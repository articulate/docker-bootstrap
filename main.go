package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if v, ok := os.LookupEnv("DEBUG_BOOTSTRAP"); ok && v != "false" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	cfg := Config{
		Service:     os.Getenv("SERVICE_NAME"),
		Product:     os.Getenv("SERVICE_PRODUCT"),
		Environment: os.Getenv("SERVICE_ENV"),
		Region:      os.Getenv("AWS_REGION"),
	}

	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	logger := log.With().
		Str("env", cfg.Environment).
		Str("service", cfg.Service).
		Str("product", cfg.Product).
		Str("region", cfg.Region).
		Logger()

	// handles peer environments (peer-some-thing => peer), which loads stage vars
	if strings.HasPrefix(cfg.Environment, "peer") {
		cfg.Environment = "stage"
	}

	if len(os.Args) < 2 {
		logger.Fatal().Msg("Missing command")
	}

	env := NewEnvMap()
	pwd, err := os.Getwd()
	if err != nil {
		logger.Warn().Err(err).Msg("Cannot determine PWD")
	}
	env.Add("PWD", pwd)
	env.Add("AWS_REGION", cfg.Region)

	if addr := os.Getenv("CONSUL_ADDR"); addr != "" {
		env.Merge(loadConsul(addr, cfg, logger))
	} else {
		logger.Warn().Msg("Not loading values from Consul. CONSUL_ADDR is not set")
	}

	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		env.Merge(loadVault(ctx, addr, cfg, logger))
	} else {
		logger.Warn().Msg("Not loading values from Vault. VAULT_ADDR is not set")
	}

	env.Add("PROCESSOR_COUNT", strconv.Itoa(runtime.NumCPU()))

	exit := run(os.Args[1], os.Args[2:], env.Environ(), logger)
	os.Exit(exit)
}

func loadConsul(addr string, c Config, l zerolog.Logger) Dict {
	l.Debug().Msg("Loading values from Consul")

	client, err := NewConsul(addr)
	if err != nil {
		l.Fatal().Err(err).Str("addr", addr).Msg("Could not connect to Consul")
	}

	paths := c.ConsulPaths()
	if p := os.Getenv("CONSUL_PATHS"); p != "" {
		paths = append(paths, strings.Split(p, ",")...)
	}

	return loadValues(client, l, paths)
}

func loadVault(ctx context.Context, addr string, c Config, l zerolog.Logger) Dict {
	l.Debug().Msg("Loading values from Vault")

	client, err := NewVault(addr, c.Region)
	if err != nil {
		l.Fatal().Err(err).Str("addr", addr).Msg("Could not connect to Vault")
	}

	token, err := vaultToken(ctx)
	if err != nil {
		l.Fatal().Err(err).Msg("Could not get Vault token")
	}

	role := os.Getenv("VAULT_ROLE")
	if role == "" {
		role = c.Service
	}

	auth, err := client.Authenticate(ctx, token, role)
	if err != nil {
		l.Fatal().Err(err).Msg("Could not authenticate Vault")
	}

	if auth == "" {
		l.Warn().Msg("Not loading values from Vault. Unable to authenticate Vault")
		return make(Dict)
	}

	paths := c.VaultPaths()
	if p := os.Getenv("VAULT_PATHS"); p != "" {
		paths = append(paths, strings.Split(p, ",")...)
	}

	values := loadValues(client, l, paths)
	values["VAULT_TOKEN"] = auth
	return values
}

func run(name string, args, env []string, l zerolog.Logger) int {
	cmd := exec.Command(name, args...) //nolint:gosec
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		l.Fatal().Err(err).Str("cmd", cmd.String()).Msg("Could not start command")
	}

	if err := cmd.Wait(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		l.Fatal().Err(err).Str("cmd", cmd.String()).Msg("Unknown error while running command")
	}

	return 0
}
