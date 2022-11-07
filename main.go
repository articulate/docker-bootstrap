package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/articulate/docker-consul-template-bootstrap/pkg/consul"
	"github.com/articulate/docker-consul-template-bootstrap/pkg/vault"
)

func main() { //nolint:gocyclo
	ctx := context.Background()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if v, ok := os.LookupEnv("DEBUG_BOOTSTRAP"); ok && v != "false" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// handles peer environments (peer-some-thing => peer), which loads stage vars
	env := strings.Split(os.Getenv("SERVICE_ENV"), "-")[0]
	if env == "peer" {
		env = "stage"
	}

	service := os.Getenv("SERVICE_NAME")
	product := os.Getenv("SERVICE_PRODUCT")

	logger := log.With().
		Str("env", env).
		Str("service", service).
		Str("product", product).
		Logger()

	vars := make(map[string]string)

	if consulAddr := os.Getenv("CONSUL_ADDR"); consulAddr != "" {
		logger.Debug().Msg("loading values from Consul")
		c, err := consul.New(consulAddr)
		if err != nil {
			logger.Fatal().Err(err).Str("addr", consulAddr).Msg("could not connect to Consul")
		}

		paths := []string{
			fmt.Sprintf("global/%s/env_vars", env), // DEPRECATED
			"global/env_vars",
			fmt.Sprintf("products/%s/env_vars", product),     // DEPRECATED
			fmt.Sprintf("apps/%s/%s/env_vars", service, env), // DEPRECATED
			fmt.Sprintf("services/%s/env_vars", service),
		}

		for _, path := range paths {
			kv, err := c.Load(path)
			if err != nil {
				logger.Fatal().Err(err).Str("path", path).Msg("could not load env vars from Consul")
			}

			for k, v := range kv {
				vars[k] = string(v)
			}
		}
	} else {
		logger.Warn().Msg("CONSUL_ADDR is not set")
	}

	if vaultAddr := os.Getenv("VAULT_ADDR"); vaultAddr != "" {
		logger.Debug().Msg("loading values from Vault")

		token, err := vaultToken(ctx)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not get vault token")
		}

		v, err := vault.New(vaultAddr)
		if err != nil {
			logger.Fatal().Err(err).Str("addr", vaultAddr).Msg("could not connect to Vault")
		}

		role := os.Getenv("VAULT_ROLE")
		if role == "" {
			role = service
		}

		t, err := v.Authenticate(ctx, token, role)
		if err != nil {
			logger.Fatal().Err(err).Str("role", role).Msg("could not authenticate with Vault")
		}
		vars["VAULT_TOKEN"] = t

		paths := []string{
			fmt.Sprintf("secret/global/%s/env_vars", env), // DEPRECATED
			"secret/global/env_vars",
			fmt.Sprintf("secret/products/%s/env_vars", product),     // DEPRECATED
			fmt.Sprintf("secret/apps/%s/%s/env_vars", service, env), // DEPRECATED
			fmt.Sprintf("secret/services/%s/env_vars", service),
		}
		if env != "stage" && env != "prod" {
			paths = []string{
				fmt.Sprintf("secret/global/%s/env_vars", env),           // DEPRECATED
				fmt.Sprintf("secret/products/%s/env_vars", product),     // DEPRECATED
				fmt.Sprintf("secret/apps/%s/%s/env_vars", service, env), // DEPRECATED
			}
		}

		for _, path := range paths {
			logger.Debug().Str("path", path).Msg("loading env vars from Vault")
			kv, err := v.Load(path)
			if err != nil {
				logger.Fatal().Err(err).Str("path", path).Msg("could not load env vars from Vault")
			}

			for k, v := range kv {
				vars[k] = v
			}
		}
	} else {
		logger.Warn().Msg("VAULT_ADDR is not set")
	}

	envs := os.Environ()
	for k, v := range vars {
		if e := os.Getenv(k); e == "" {
			envs = append(envs, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cmd := exec.Command(os.Args[1], os.Args[2:]...) //nolint:gosec
	cmd.Env = envs
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		logger.Fatal().Err(err).Str("command", cmd.String()).Msg("could not start command")
	}

	if err := cmd.Wait(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		logger.Fatal().Err(err).Str("command", cmd.String()).Msg("unknown error during command")
	}
}

func vaultToken(ctx context.Context) (string, error) {
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		return token, nil
	}

	if encToken := os.Getenv("ENCRYPTED_VAULT_TOKEN"); encToken != "" {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return "", err
		}

		client := kms.NewFromConfig(cfg)
		return vault.DecodeToken(ctx, client, encToken)
	}

	return "", nil
}
