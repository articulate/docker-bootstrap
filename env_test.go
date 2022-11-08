package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvMap(t *testing.T) {
	t.Setenv("DOCKER_CONSUL_BOOTSTRAP_TEST_FOO", "ignore")
	t.Setenv("DOCKER_CONSUL_BOOTSTRAP_TEST_BAR", "ignored")

	e := NewEnvMap()
	e.Merge(map[string]string{
		"DOCKER_CONSUL_BOOTSTRAP_TEST_FOO":  "changed",
		"DOCKER_CONSUL_BOOTSTRAP_TEST_TEST": "testing",
		"DOCKER_CONSUL_BOOTSTRAP_TEST_BAZ":  "foo",
	})
	e.Merge(map[string]string{
		"DOCKER_CONSUL_BOOTSTRAP_TEST_BAZ": "bar",
	})
	e.Add("DOCKER_CONSUL_BOOTSTRAP_TEST_BAR", "testing")

	assert.ElementsMatch(t, []string{
		"DOCKER_CONSUL_BOOTSTRAP_TEST_TEST=testing",
		"DOCKER_CONSUL_BOOTSTRAP_TEST_BAZ=bar",
	}, e.Environ())
}
