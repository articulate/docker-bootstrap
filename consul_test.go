package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsul_Load(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/v1/kv/{key...}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("key") == "foo/bar" {
			foo := base64.StdEncoding.EncodeToString([]byte("foo"))
			bar := base64.StdEncoding.EncodeToString([]byte("bar"))
			fmt.Fprintf(w, `[{"Key":"foo","Value":"%s"},{"Key":"none"},{"Key":"bar","Value":"%s"}]`, foo, bar)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	})

	c, err := NewConsul(server.URL)
	require.NoError(t, err)

	data, err := c.Load("foo/bar")
	require.NoError(t, err)
	assert.Equal(t, Dict{"FOO": "foo", "BAR": "bar"}, data)

	empty, err := c.Load("empty")
	require.ErrorContains(t, err, "could not load empty: Unexpected response code: 400")
	assert.Empty(t, empty)
}

func TestConsulKey(t *testing.T) {
	assert.Equal(t, "BAZ", consulKey("foo/bar/baz"))
	assert.Equal(t, "TEST", consulKey("test"))
}
