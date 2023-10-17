package main

import (
	"fmt"
)

type Config struct {
	Service     string
	Environment string
	Region      string
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
