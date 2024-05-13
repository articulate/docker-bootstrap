package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	logLevel = new(slog.LevelVar)
	// killWait is the time to wait before forcefully terminating the child process
	killWait = 5 * time.Second
	// signals contains the signals that will be forwarded to the child process
	signals = []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	}
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
	if v, err := strconv.ParseBool(os.Getenv("DEBUG_BOOTSTRAP")); err == nil && v {
		logLevel.Set(slog.LevelDebug)
	}

	cfg := NewFromEnv()
	logger = logger.With("env", cfg.Environment, "service", cfg.Service, "region", cfg.Region)
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		logger.ErrorContext(ctx, "Missing command")
		os.Exit(1)
	}

	env := NewEnvMap()
	if addr := os.Getenv("CONSUL_ADDR"); addr != "" {
		c, err := loadConsul(ctx, addr, cfg, logger)
		if err != nil {
			logger.ErrorContext(ctx, "Could not load values from Consul", "error", err)
			os.Exit(1)
		}
		env.Merge(c)
	} else {
		logger.WarnContext(ctx, "Not loading values from Consul. CONSUL_ADDR is not set")
	}

	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		v, err := loadVault(ctx, addr, cfg, logger)
		if err != nil {
			logger.ErrorContext(ctx, "Could not load values from Vault", "error", err)
			os.Exit(1)
		}
		env.Merge(v)
	} else {
		logger.WarnContext(ctx, "Not loading values from Vault. VAULT_ADDR is not set")
	}

	pwd, err := os.Getwd()
	if err != nil {
		logger.WarnContext(ctx, "Cannot determine PWD", "error", err)
	}
	env.Add("PWD", pwd)
	env.Add("AWS_REGION", cfg.Region)
	env.Add("SERVICE_ENV", cfg.Environment)
	env.Add("PROCESSOR_COUNT", strconv.Itoa(runtime.NumCPU()))

	if err := validate(ctx, cfg, env, logger); err != nil {
		logger.ErrorContext(ctx, "Missing dependencies", "error", err)
		os.Exit(4)
	}

	os.Exit(run(ctx, os.Args[1], os.Args[2:], env.Environ(), logger))
}

func loadConsul(ctx context.Context, addr string, c *Config, l *slog.Logger) (Dict, error) {
	l.Debug("Loading values from Consul")

	client, err := NewConsul(addr)
	if err != nil {
		return nil, serror(fmt.Errorf("Could not connect to Consul: %w", err), "addr", addr)
	}

	paths := c.ConsulPaths()
	if p := os.Getenv("CONSUL_PATHS"); p != "" {
		paths = append(paths, strings.Split(p, ",")...)
	}

	return loadValues(ctx, client, l, paths)
}

func loadVault(ctx context.Context, addr string, c *Config, l *slog.Logger) (Dict, error) {
	l.DebugContext(ctx, "Loading values from Vault")

	client, err := NewVault(addr, c.Region)
	if err != nil {
		return nil, serror(fmt.Errorf("Could not connect to Vault: %w", err), "addr", addr)
	}

	token, err := vaultToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("Could not get Vault token: %w", err)
	}

	role := os.Getenv("VAULT_ROLE")
	if role == "" {
		role = c.Service
	}

	auth, err := client.Authenticate(ctx, token, role)
	if err != nil {
		return nil, fmt.Errorf("Could not authenticate Vault: %w", err)
	}

	if auth == "" {
		l.WarnContext(ctx, "Not loading values from Vault. Unable to authenticate Vault")
		return make(Dict), nil
	}

	paths := c.VaultPaths()
	if p := os.Getenv("VAULT_PATHS"); p != "" {
		paths = append(paths, strings.Split(p, ",")...)
	}

	values, err := loadValues(ctx, client, l, paths)
	values["VAULT_TOKEN"] = auth
	return values, err
}

func run(ctx context.Context, name string, args, env []string, l *slog.Logger) int {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if name != "bash" && name != "sh" && name != "zsh" && name != "fish" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := cmd.Start(); err != nil {
		l.ErrorContext(ctx, "Could not start command", "error", err, "cmd", cmd.String())
		return 1
	}

	sigch := make(chan os.Signal, 1)
	exitch := make(chan os.Signal, 1)
	signal.Notify(sigch, signals...)
	signal.Notify(exitch, syscall.SIGINT)
	defer signal.Stop(sigch)
	defer signal.Stop(exitch)

	// forward signals to the child process
	go func() {
		for {
			s := <-sigch
			l.DebugContext(ctx, "Sending signal", "signal", s.String())
			if err := cmd.Process.Signal(s); err != nil {
				l.ErrorContext(ctx, "Could not send signal to command", "error", err, "cmd", cmd.String(), "signal", s.String())
			}
		}
	}()

	// handle forceful termination
	go func() {
		<-exitch
		time.Sleep(killWait)
		l.DebugContext(ctx, "Terminating unrespnsive process", "cmd", cmd.String())
		cancel()
	}()

	if err := cmd.Wait(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		l.ErrorContext(ctx, "Unknown error while running command", "error", err, "cmd", cmd.String())
		return 3
	}

	return 0
}
