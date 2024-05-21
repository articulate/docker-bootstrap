package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var logLevel = new(slog.LevelVar)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
	if v, err := strconv.ParseBool(os.Getenv("DEBUG_BOOTSTRAP")); err == nil && v {
		logLevel.Set(slog.LevelDebug)
	}

	cfg := NewFromEnv()
	logger = logger.With(
		slog.String("env", cfg.Environment),
		slog.String("service", cfg.Service),
		slog.String("product", cfg.Product),
		slog.String("region", cfg.Region),
		slog.String("program", cfg.Program),
	)
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		logger.ErrorContext(ctx, "Missing command")
		os.Exit(1)
	}

	env, err := loadEnvVars(ctx, cfg, logger)
	if err != nil {
		logger.ErrorContext(ctx, "Could not load environment variables", "error", err)
		os.Exit(1)
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

	cmd := os.Args[1]
	args := os.Args[2:]

	if cmd == "yarn" || cmd == "npm" {
		logger.WarnContext(ctx, cmd+" is not recommended. You might see unexpected behavior. Use node instead.")
	}

	if err := run(cmd, args, env.Environ()); err != nil {
		logger.ErrorContext(ctx, "Could not run command", "error", err)
		os.Exit(1)
	}
}

func loadEnvVars(ctx context.Context, cfg *Config, l *slog.Logger) (*EnvMap, error) {
	env := NewEnvMap()
	if addr := os.Getenv("CONSUL_ADDR"); addr != "" {
		c, err := loadConsul(ctx, addr, cfg, l)
		if err != nil {
			return env, fmt.Errorf("could not load values from Consul: %w", err)
		}
		env.Merge(c)
	} else {
		l.WarnContext(ctx, "Not loading values from Consul. CONSUL_ADDR is not set")
	}

	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		v, err := loadVault(ctx, addr, cfg, l)
		if err != nil {
			return env, fmt.Errorf("could not load values from Vault: %w", err)
		}
		env.Merge(v)
	} else {
		l.WarnContext(ctx, "Not loading values from Vault. VAULT_ADDR is not set")
	}

	return env, nil
}

func loadConsul(ctx context.Context, addr string, c *Config, l *slog.Logger) (Dict, error) {
	l.DebugContext(ctx, "Loading values from Consul")

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

func run(name string, args, env []string) error {
	bin, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("could not find %s: %w", name, err)
	}

	if err := syscall.Exec(bin, args, env); err != nil {
		return fmt.Errorf("could not execute %s %s: %w", name, strings.Join(args, " "), err)
	}

	return nil
}
