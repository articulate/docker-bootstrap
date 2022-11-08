package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConsul(t *testing.T) {
	c, err := NewConsul("foo:b\ar;baz")
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "parse")
}

func TestConsulKey(t *testing.T) {
	assert.Equal(t, consulKey("foo/bar/baz"), "BAZ")
	assert.Equal(t, consulKey("test"), "TEST")
}
