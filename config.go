package main

import (
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	Service           string
	Environment       string
	Region            string
	ServiceDefinition string
}

// NewFromEnv creates a new Config from environment variables and defaults
func NewFromEnv() *Config {
	cfg := &Config{
		Service:           os.Getenv("SERVICE_NAME"),
		Environment:       os.Getenv("SERVICE_ENV"),
		Region:            os.Getenv("AWS_REGION"),
		ServiceDefinition: os.Getenv("SERVICE_DEFINITION"),
	}

	if cfg.Service == "" {
		slog.Warn("SERVICE_NAME is not set, will not load service values")
	}

	if cfg.Environment == "" {
		slog.Warn("SERVICE_ENV is not set, defaulting to dev")
		cfg.Environment = "dev"
	}

	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	if cfg.ServiceDefinition == "" {
		cfg.ServiceDefinition = "service.json"
	}

	return cfg
}

// ConsulPaths returns the paths from Consul to load
func (c *Config) ConsulPaths() []string {
	paths := []string{
		"global/env_vars",
		fmt.Sprintf("global/%s/env_vars", c.Environment),
	}

	if c.Service != "" {
		paths = append(
			paths,
			fmt.Sprintf("services/%s/env_vars", c.Service),
			fmt.Sprintf("services/%s/%s/env_vars", c.Service, c.Environment),
		)
	}

	return paths
}

// VaultPaths returns the paths from Vault to load
func (c *Config) VaultPaths() []string {
	isPublic := c.Environment == "stage" || c.Environment == "prod"
	paths := []string{}

	if isPublic {
		paths = append(paths, "secret/global/env_vars")
	}

	paths = append(paths, fmt.Sprintf("secret/global/%s/env_vars", c.Environment))

	if c.Service != "" {
		if isPublic {
			paths = append(paths, fmt.Sprintf("secret/services/%s/env_vars", c.Service))
		}

		paths = append(paths, fmt.Sprintf("secret/services/%s/%s/env_vars", c.Service, c.Environment))
	}

	return paths
}
