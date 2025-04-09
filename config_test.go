package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromEnv(t *testing.T) {
	t.Run("uses values from env vars", func(t *testing.T) {
		t.Setenv("SERVICE_NAME", "my-service")
		t.Setenv("SERVICE_ENV", "test")
		t.Setenv("SERVICE_PRODUCT", "foo")
		t.Setenv("AWS_REGION", "us-east-1")
		t.Setenv("SERVICE_DEFINITION", "service.yaml")
		t.Setenv("BOOTSTRAP_SKIP_VALIDATION", "true")

		c := NewFromEnv()
		assert.Equal(t, "my-service", c.Service)
		assert.Equal(t, "test", c.Environment)
		assert.Equal(t, "foo", c.Product)
		assert.Equal(t, "us-east-1", c.Region)
		assert.Equal(t, "service.yaml", c.ServiceDefinition)
		assert.True(t, c.SkipValidation)
	})

	t.Run("sets defaults", func(t *testing.T) {
		t.Setenv("SERVICE_NAME", "")
		t.Setenv("SERVICE_ENV", "")
		t.Setenv("SERVICE_PRODUCT", "foo")
		t.Setenv("AWS_REGION", "")
		t.Setenv("SERVICE_DEFINITION", "")
		t.Setenv("BOOTSTRAP_SKIP_VALIDATION", "0")

		oglog := slog.Default()
		t.Cleanup(func() {
			slog.SetDefault(oglog)
		})

		log, out := testLogger()
		slog.SetDefault(log)

		c := NewFromEnv()
		assert.Empty(t, c.Service)
		assert.Equal(t, "dev", c.Environment)
		assert.Equal(t, "foo", c.Product)
		assert.Equal(t, "us-east-1", c.Region)
		assert.Equal(t, "service.json", c.ServiceDefinition)
		assert.False(t, c.SkipValidation)

		assert.Contains(t, out.String(), `"level":"WARN","msg":"SERVICE_NAME is not set, will not load service values"`)
		assert.Contains(t, out.String(), `"level":"WARN","msg":"SERVICE_ENV is not set, defaulting to dev"`)
	})
}

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
