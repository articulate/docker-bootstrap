package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependency_Required(t *testing.T) {
	d := dependency{
		dependencyInner: dependencyInner{
			Regions: []string{"us-east-1"},
		},
	}
	assert.True(t, d.Required("us-east-1"))
	assert.False(t, d.Required("us-west-2"))

	d = dependency{}
	assert.True(t, d.Required("us-east-1"))
	assert.True(t, d.Required("us-west-2"))
}

func TestValidate(t *testing.T) { //nolint:funlen
	s := filepath.Join(t.TempDir(), "service.json")
	require.NoError(t, os.WriteFile(s, []byte(`{
		"dependencies": {
			"env_vars": {
				"required": [
					"FOO",
					{
						"key": "BAR",
						"regions": ["us-east-1"]
					},
					{
						"key":"BAZ"
					}
				],
				"optional": [
					"QUX",
					{
						"key": "FOOBAR",
						"regions": ["eu-central-1"]
					},
					{
						"key":"FOOBAZ"
					}
				]
			}
		}
}`), 0o600))

	l, log := testLogger()
	c := &Config{ServiceDefinition: s, Region: "us-east-1"}

	e := NewEnvMap()

	err := validate(context.TODO(), c, e, l)
	require.ErrorIs(t, err, ErrMissingEnvVars)
	assert.Contains(
		t,
		log.String(),
		`"ERROR","msg":"Missing required environment variables","env_vars":["FOO","BAR","BAZ"]`,
	)
	assert.Contains(t, log.String(), `"WARN","msg":"Missing optional environment variables","env_vars":["QUX","FOOBAZ"]`)

	// Empty env vars should be considered missing
	e.Add("FOO", "")
	t.Setenv("BAR", "")

	log.Reset()
	err = validate(context.TODO(), c, e, l)
	require.ErrorIs(t, err, ErrMissingEnvVars)
	assert.Contains(t, log.String(), `Missing required environment variables","env_vars":["FOO","BAR"`)

	// Set all required env vars
	c.Region = "eu-central-1"
	e.Add("FOO", "foo")
	t.Setenv("BAZ", "baz")

	log.Reset()
	err = validate(context.TODO(), c, e, l)
	require.NoError(t, err)
	assert.NotContains(t, log.String(), "Missing required environment variables")
	assert.Contains(t, log.String(), "Missing optional environment variables")

	// Set all optional env vars
	e.Add("QUX", "qux")
	e.Add("FOOBAR", "foobar")
	t.Setenv("FOOBAZ", "foobaz")

	log.Reset()
	err = validate(context.TODO(), c, e, l)
	require.NoError(t, err)
	assert.NotContains(t, log.String(), "Missing required environment variables")
	assert.NotContains(t, log.String(), "Missing optional environment variables")
}
