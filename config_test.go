package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ConsulPaths(t *testing.T) {
	c := Config{
		Service:     "foo",
		Product:     "bar",
		Environment: "stage",
	}

	assert.Equal(t, []string{
		"global/stage/env_vars",
		"global/env_vars",
		"products/bar/env_vars",
		"apps/foo/stage/env_vars",
		"services/foo/env_vars",
	}, c.ConsulPaths())
}

func TestConfig_VaultPaths(t *testing.T) {
	c := Config{
		Service:     "foo",
		Product:     "bar",
		Environment: "stage",
	}

	assert.Equal(t, []string{
		"secret/global/stage/env_vars",
		"secret/global/env_vars",
		"secret/products/bar/env_vars",
		"secret/apps/foo/stage/env_vars",
		"secret/services/foo/env_vars",
	}, c.VaultPaths())

	c.Environment = "prod"
	assert.Equal(t, []string{
		"secret/global/prod/env_vars",
		"secret/global/env_vars",
		"secret/products/bar/env_vars",
		"secret/apps/foo/prod/env_vars",
		"secret/services/foo/env_vars",
	}, c.VaultPaths())

	c.Environment = "dev"
	assert.Equal(t, []string{
		"secret/global/dev/env_vars",
		"secret/products/bar/dev/env_vars",
		"secret/apps/foo/dev/env_vars",
	}, c.VaultPaths())
}
