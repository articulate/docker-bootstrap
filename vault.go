package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/aws"
	"github.com/hashicorp/vault/api/auth/kubernetes"
)

const k8sTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token" //nolint:gosec

// Vault is a client for fetching values from Vault
type Vault struct {
	client       *api.Client
	region       string
	k8sTokenFile string
}

// NewVault returns a new Vault client
func NewVault(addr, region string) (*Vault, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not connect to %s: %w", addr, err)
	}

	return &Vault{client, region, k8sTokenFile}, nil
}

// Authenticate authenticates the client with Vault
func (v *Vault) Authenticate(ctx context.Context, token, role string) (string, error) {
	if token != "" {
		v.client.SetToken(token)
		return token, nil
	}

	auth, err := v.getAuthMethod(role)
	if auth == nil || err != nil {
		return "", err
	}

	secret, err := v.client.Auth().Login(ctx, auth)
	if err != nil {
		return "", fmt.Errorf("could not authenticate: %w", err)
	}

	id, err := secret.TokenID()
	if err != nil {
		return "", fmt.Errorf("could not get token: %w", err)
	}
	return id, nil
}

// getAuthMethod tries to determine the auth method to be used with Vault
func (v *Vault) getAuthMethod(role string) (api.AuthMethod, error) {
	if _, err := os.Stat(v.k8sTokenFile); !os.IsNotExist(err) {
		auth, err := kubernetes.NewKubernetesAuth(role, kubernetes.WithServiceAccountTokenPath(v.k8sTokenFile))
		if err != nil {
			return nil, fmt.Errorf("could not authenticate with kubernetes: %w", err)
		}
		return auth, nil
	}

	ecs := os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	lambda := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")

	if ecs == "" && lambda == "" {
		return nil, nil
	}

	auth, err := aws.NewAWSAuth(aws.WithRegion(v.region), aws.WithRole(role))
	if err != nil {
		return nil, fmt.Errorf("could not authenticate with IAM: %w", err)
	}

	return auth, nil
}

// Load fetches values from the given paths
func (v *Vault) Load(path string) (Dict, error) {
	vars := make(Dict)
	client := v.client.Logical()
	secret, err := client.List(path)
	if err != nil {
		return vars, fmt.Errorf("could not list %s: %w", path, err)
	}

	for _, key := range secretKeys(secret) {
		k := fmt.Sprintf("%s/%s", path, key)
		s, err := client.Read(k)
		if err != nil {
			return vars, fmt.Errorf("could not read %s: %w", k, err)
		}
		if s == nil {
			continue
		}
		if val, ok := s.Data["value"].(string); ok {
			vars[key] = val
		}
	}

	return vars, nil
}

// secretKeys returns a list of keys for the given secret
func secretKeys(s *api.Secret) []string {
	if s == nil {
		return []string{}
	}

	keys, ok := s.Data["keys"].([]interface{})
	if !ok {
		return []string{}
	}

	list := []string{}
	for _, k := range keys {
		list = append(list, k.(string))
	}

	return list
}
