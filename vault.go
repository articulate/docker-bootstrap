package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/aws"
	"github.com/hashicorp/vault/api/auth/kubernetes"
)

const k8sTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token" //nolint:gosec

// Vault is a client for fetching values from Vault
type Vault struct {
	client *api.Client
}

// NewVault returns a new Vault client
func NewVault(addr string) (*Vault, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Vault{client}, nil
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
		return "", err
	}

	return secret.TokenID()
}

// getAuthMethod tries to determine the auth method to be used with Vault
func (v *Vault) getAuthMethod(role string) (api.AuthMethod, error) {
	if _, err := os.Stat(k8sTokenFile); !os.IsNotExist(err) {
		return kubernetes.NewKubernetesAuth(role)
	}

	ecs := os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	lambda := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")

	if ecs == "" && lambda == "" {
		return nil, nil
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	return aws.NewAWSAuth(aws.WithRegion(region), aws.WithRole(role))
}

// Load fetches values from the given paths
func (v *Vault) Load(path string) (Dict, error) {
	vars := make(Dict)
	client := v.client.Logical()
	secret, err := client.List(path)
	if err != nil {
		return vars, err
	}

	for _, key := range secretKeys(secret) {
		s, err := client.Read(fmt.Sprintf("%s/%s", path, key))
		if err != nil {
			return vars, err
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

// vaultToken returns the vault token from env var
func vaultToken(ctx context.Context) (string, error) {
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		return token, nil
	}

	if enc := os.Getenv("ENCRYPTED_VAULT_TOKEN"); enc != "" {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return "", err
		}

		client := kms.NewFromConfig(cfg)
		return decodeToken(ctx, client, enc)
	}

	return "", nil
}
