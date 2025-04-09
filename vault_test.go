package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVault_Authenticate(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	v, err := NewVault(server.URL, "us-east-1")
	require.NoError(t, err)

	// Static token
	token, err := v.Authenticate(t.Context(), "foobar", "")
	require.NoError(t, err)
	assert.Equal(t, "foobar", token)
	assert.Equal(t, "foobar", v.client.Token())

	// k8s auth
	mux.HandleFunc("/v1/auth/kubernetes/login", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"auth":{"client_token":"k8s-auth-token"}}`)
	})
	v.k8sTokenFile = filepath.Join(t.TempDir(), "k8s-token")
	require.NoError(t, os.WriteFile(v.k8sTokenFile, []byte("k8s-jwt"), 0o600))
	token, err = v.Authenticate(t.Context(), "", "k8s")
	require.NoError(t, err)
	assert.Equal(t, "k8s-auth-token", token)
	assert.Equal(t, "k8s-auth-token", v.client.Token())
	require.NoError(t, os.Remove(v.k8sTokenFile))

	// aws auth
	mux.HandleFunc("/v1/auth/aws/login", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"auth":{"client_token":"aws-auth-token"}}`)
	})
	t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "foo")
	t.Setenv("AWS_ACCESS_KEY_ID", "my-access-token")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "my-secret-token")
	t.Setenv("AWS_SESSION_TOKEN", "my-session-token")
	token, err = v.Authenticate(t.Context(), "", "aws")
	require.NoError(t, err)
	assert.Equal(t, "aws-auth-token", token)
	assert.Equal(t, "aws-auth-token", v.client.Token())
}

func TestVault_Load(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/v1/{key...}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list") == "true" {
			fmt.Fprint(w, `{"data":{"keys":["FOO","BAR","BAZ"]}}`)
			return
		}

		switch path.Base(r.PathValue("key")) {
		case "FOO":
			fmt.Fprint(w, `{"data":{"value":"foo"}}`)
		case "BAR":
			fmt.Fprint(w, `{"data":{"value":"bar"}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	v, err := NewVault(server.URL, "us-east-1")
	require.NoError(t, err)

	data, err := v.Load("foo/bar")
	require.NoError(t, err)
	require.Equal(t, Dict{"FOO": "foo", "BAR": "bar"}, data)
}

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
