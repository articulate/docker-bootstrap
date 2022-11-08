package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Load(path string) (Dict, error) {
	args := m.Called(path)

	return map[string]string{args.String(0): args.String(1)}, args.Error(2)
}

func TestLoadValues(t *testing.T) {
	log := &bytes.Buffer{}
	logger := zerolog.New(log)

	m := new(mockClient)
	m.On("Load", "foo").Return("foo", "bar", nil)
	m.On("Load", "bar").Return("bar", "baz", nil)

	assert.Equal(t, Dict{
		"foo": "bar",
		"bar": "baz",
	}, loadValues(m, logger, []string{"foo", "bar"}))

	assert.Equal(t, []byte(`{"level":"debug","path":"foo","message":"Loading values"}
{"level":"debug","path":"bar","message":"Loading values"}
`), log.Bytes())

	m.AssertExpectations(t)
}

func TestLoadValues_Fatal(t *testing.T) {
	if os.Getenv("TEST_FATAL") == "true" {
		m := new(mockClient)
		m.On("Load", "none").Return("", "", fmt.Errorf("test error"))

		loadValues(m, zerolog.New(os.Stderr), []string{"none"})
		return
	}

	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name())) //nolint:gosec
	cmd.Env = append(cmd.Env, "TEST_FATAL=true")

	_, err := cmd.Output()
	assert.Error(t, err)
	assert.Equal(t, 1, err.(*exec.ExitError).ExitCode())
	assert.Equal(t, []byte(`{"level":"debug","path":"none","message":"Loading values"}
{"level":"fatal","error":"test error","path":"none","message":"Could not load values"}
`), err.(*exec.ExitError).Stderr)
}
