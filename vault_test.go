package main

import (
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestSecretKeys(t *testing.T) {
	secret := &api.Secret{
		Data: map[string]interface{}{
			"keys": []interface{}{"foo", "bar", "baz"},
		},
	}

	assert.ElementsMatch(t, []string{"foo", "bar", "baz"}, secretKeys(secret))
	assert.Equal(t, []string{}, secretKeys(nil))
	assert.Equal(t, []string{}, secretKeys(&api.Secret{}))
}
