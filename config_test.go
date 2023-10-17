package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ConsulPaths(t *testing.T) {
	t.Parallel()

	t.Run("full config", func(t *testing.T) {
		t.Parallel()

		c := Config{Service: "foo", Environment: "stage"}
		assert.Equal(t, []string{
			"global/env_vars",
			"global/stage/env_vars",
			"services/foo/env_vars",
			"services/foo/stage/env_vars",
		}, c.ConsulPaths())
	})

	t.Run("no service", func(t *testing.T) {
		t.Parallel()

		c := Config{Environment: "stage"}
		assert.Equal(t, []string{
			"global/env_vars",
			"global/stage/env_vars",
		}, c.ConsulPaths())
	})
}

func TestConfig_VaultPaths(t *testing.T) {
	t.Parallel()

	t.Run("stage", func(t *testing.T) {
		t.Parallel()

		c := Config{Service: "foo", Environment: "stage"}
		assert.Equal(t, []string{
			"secret/global/env_vars",
			"secret/global/stage/env_vars",
			"secret/services/foo/env_vars",
			"secret/services/foo/stage/env_vars",
		}, c.VaultPaths())
	})

	t.Run("prod", func(t *testing.T) {
		t.Parallel()

		c := Config{Service: "foo", Environment: "prod"}
		assert.Equal(t, []string{
			"secret/global/env_vars",
			"secret/global/prod/env_vars",
			"secret/services/foo/env_vars",
			"secret/services/foo/prod/env_vars",
		}, c.VaultPaths())
	})

	t.Run("dev", func(t *testing.T) {
		t.Parallel()

		c := Config{Service: "foo", Environment: "dev"}
		assert.Equal(t, []string{
			"secret/global/dev/env_vars",
			"secret/services/foo/dev/env_vars",
		}, c.VaultPaths())
	})

	t.Run("no service", func(t *testing.T) {
		t.Parallel()

		c := Config{Environment: "prod"}

		assert.Equal(t, []string{
			"secret/global/env_vars",
			"secret/global/prod/env_vars",
		}, c.VaultPaths())
	})

	t.Run("no service dev", func(t *testing.T) {
		t.Parallel()

		c := Config{Environment: "dev"}
		assert.Equal(t, []string{
			"secret/global/dev/env_vars",
		}, c.VaultPaths())
	})
}
