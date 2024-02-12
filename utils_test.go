package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Load(path string) (Dict, error) {
	args := m.Called(path)

	return map[string]string{args.String(0): args.String(1)}, args.Error(2) //nolint:wrapcheck
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
		m.On("Load", "none").Return("", "", errors.New("test error")) //nolint:goerr113

		loadValues(m, zerolog.New(os.Stderr), []string{"none"})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+t.Name()) //nolint:gosec
	cmd.Env = append(cmd.Env, "TEST_FATAL=true")

	var exit *exec.ExitError
	_, err := cmd.Output()
	require.ErrorAs(t, err, &exit)
	assert.Equal(t, 1, exit.ExitCode())
	assert.Equal(t, []byte(`{"level":"debug","path":"none","message":"Loading values"}
{"level":"fatal","error":"test error","path":"none","message":"Could not load values"}
`), exit.Stderr)
}
