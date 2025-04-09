package main

import (
	"context"
	"errors"
	"testing"

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
	logger, log := testLogger()

	m := new(mockClient)
	m.On("Load", "foo").Return("foo", "bar", nil)
	m.On("Load", "bar").Return("bar", "baz", nil)

	d, err := loadValues(context.TODO(), m, logger, []string{"foo", "bar"})
	require.NoError(t, err)
	assert.Equal(t, Dict{
		"foo": "bar",
		"bar": "baz",
	}, d)

	//nolint:testifylint
	assert.Equal(t, `{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"foo"}
{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"bar"}
`, log.String())

	m.AssertExpectations(t)
}

func TestLoadValues_Error(t *testing.T) {
	logger, log := testLogger()

	m := new(mockClient)
	m.On("Load", "none").Return("", "", errors.New("test error")) //nolint:err113

	d, err := loadValues(context.TODO(), m, logger, []string{"none"})
	require.ErrorContains(t, err, "could not load values: test error")
	assert.Equal(t, Dict{}, d)

	assert.JSONEq(t, `{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"none"}`, log.String())

	m.AssertExpectations(t)
}
