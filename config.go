package main

import "fmt"

type Config struct {
	Service     string
	Product     string
	Environment string
}

// ConsulPaths returns the paths from Consul to load
func (c *Config) ConsulPaths() []string {
	return []string{
		fmt.Sprintf("global/%s/env_vars", c.Environment), // DEPRECATED
		"global/env_vars",
		fmt.Sprintf("products/%s/env_vars", c.Product),               // DEPRECATED
		fmt.Sprintf("apps/%s/%s/env_vars", c.Service, c.Environment), // DEPRECATED
		fmt.Sprintf("services/%s/env_vars", c.Service),
	}
}

// VaultPaths returns the paths from Vault to load
func (c *Config) VaultPaths() []string {
	if c.Environment == "stage" || c.Environment == "prod" {
		return []string{
			fmt.Sprintf("secret/global/%s/env_vars", c.Environment), // DEPRECATED
			"secret/global/env_vars",
			fmt.Sprintf("secret/products/%s/env_vars", c.Product),               // DEPRECATED
			fmt.Sprintf("secret/apps/%s/%s/env_vars", c.Service, c.Environment), // DEPRECATED
			fmt.Sprintf("secret/services/%s/env_vars", c.Service),
		}
	}

	return []string{
		fmt.Sprintf("secret/global/%s/%s/env_vars", c.Environment, c.Environment), // DEPRECATED
		fmt.Sprintf("secret/products/%s/%s/env_vars", c.Product, c.Environment),   // DEPRECATED
		fmt.Sprintf("secret/apps/%s/%s/env_vars", c.Service, c.Environment),       // DEPRECATED
	}
}
