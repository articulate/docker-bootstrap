package main

import (
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

	d, err := loadValues(t.Context(), m, logger, []string{"foo", "bar"}, map[string]origin{})
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

	d, err := loadValues(t.Context(), m, logger, []string{"none"}, map[string]origin{})
	require.ErrorContains(t, err, "could not load values: test error")
	assert.Equal(t, Dict{}, d)

	assert.JSONEq(t, `{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"none"}`, log.String())

	m.AssertExpectations(t)
}

func TestLoadValues_Overwrite(t *testing.T) {
	logger, log := testLogger()
	sources := map[string]origin{}

	m := new(mockClient)
	m.On("Load", "first").Return("KEY", "one", nil)
	m.On("Load", "second").Return("KEY", "two", nil)
	m.On("Load", "third").Return("KEY", "two", nil)

	d, err := loadValues(t.Context(), m, logger, []string{"first", "second"}, sources)
	require.NoError(t, err)
	assert.Equal(t, Dict{"KEY": "two"}, d)

	// A later source with the same value is a harmless duplicate.
	d, err = loadValues(t.Context(), m, logger, []string{"third"}, sources)
	require.NoError(t, err)
	assert.Equal(t, Dict{"KEY": "two"}, d)

	//nolint:testifylint
	assert.Equal(t, `{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"first"}
{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"second"}
{"time":"test-time","level":"WARN","msg":"Overwriting value","key":"KEY","path":"second","previous_path":"first"}
{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"third"}
`, log.String())

	m.AssertExpectations(t)
}

func TestLoadValues_HostEnv(t *testing.T) {
	t.Setenv("HOST_KEY", "host")
	logger, log := testLogger()

	m := new(mockClient)
	m.On("Load", "first").Return("HOST_KEY", "loaded", nil)
	m.On("Load", "second").Return("HOST_KEY", "loaded", nil)

	_, err := loadValues(t.Context(), m, logger, []string{"first", "second"}, map[string]origin{})
	require.NoError(t, err)

	//nolint:testifylint
	assert.Equal(t, `{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"first"}
{"time":"test-time","level":"WARN","msg":"Host env var takes precedence","key":"HOST_KEY","path":"first"}
{"time":"test-time","level":"DEBUG","msg":"Loading values","path":"second"}
`, log.String())

	m.AssertExpectations(t)
}
