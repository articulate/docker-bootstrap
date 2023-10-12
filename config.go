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
	return []string{
		"global/env_vars",
		fmt.Sprintf("global/%s/env_vars", c.Environment),
		fmt.Sprintf("services/%s/env_vars", c.Service),
		fmt.Sprintf("services/%s/%s/env_vars", c.Service, c.Environment),
	}
}

// VaultPaths returns the paths from Vault to load
func (c *Config) VaultPaths() []string {
	if c.Environment == "stage" || c.Environment == "prod" {
		return []string{
			"secret/global/env_vars",
			fmt.Sprintf("secret/global/%s/env_vars", c.Environment),
			fmt.Sprintf("secret/services/%s/env_vars", c.Service),
			fmt.Sprintf("secret/services/%s/%s/env_vars", c.Service, c.Environment),
		}
	}

	return []string{
		fmt.Sprintf("secret/global/%s/env_vars", c.Environment),
		fmt.Sprintf("secret/services/%s/%s/env_vars", c.Service, c.Environment),
	}
}
