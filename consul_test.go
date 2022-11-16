package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsulKey(t *testing.T) {
	assert.Equal(t, consulKey("foo/bar/baz"), "BAZ")
	assert.Equal(t, consulKey("test"), "TEST")
}
