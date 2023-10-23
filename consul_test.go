package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsulKey(t *testing.T) {
	assert.Equal(t, "BAZ", consulKey("foo/bar/baz"))
	assert.Equal(t, "TEST", consulKey("test"))
}
