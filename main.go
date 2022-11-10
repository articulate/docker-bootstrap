package main

import (
	"context"
	"os"
	"os/exec"
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
	}

	logger := log.With().
		Str("env", cfg.Environment).
		Str("service", cfg.Service).
		Str("product", cfg.Product).
		Logger()

	// handles peer environments (peer-some-thing => peer), which loads stage vars
	if strings.HasPrefix(cfg.Environment, "peer") {
		cfg.Environment = "stage"
	}

	if len(os.Args) < 2 {
		logger.Fatal().Msg("Missing command")
	}

	env := NewEnvMap()

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

	exit := run(os.Args[1], os.Args[2:], env.Environ(), logger)
	os.Exit(exit)
}

func loadConsul(addr string, c Config, l zerolog.Logger) Dict {
	l.Debug().Msg("Loading values from Consul")

	client, err := NewConsul(addr)
	if err != nil {
		l.Fatal().Err(err).Str("addr", addr).Msg("Could not connect to Consul")
	}

	return loadValues(client, l, c.ConsulPaths())
}

func loadVault(ctx context.Context, addr string, c Config, l zerolog.Logger) Dict {
	l.Debug().Msg("Loading values from Vault")

	client, err := NewVault(addr)
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

	values := loadValues(client, l, c.VaultPaths())
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
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		l.Fatal().Err(err).Str("cmd", cmd.String()).Msg("Unknown error while running command")
	}

	return 0
}
