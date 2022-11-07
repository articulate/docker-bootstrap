package vault

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/aws"
	"github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/samber/lo"
)

type Vault struct {
	client *api.Client
}

const k8sTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token" //nolint:gosec

func New(addr string) (*Vault, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Vault{client}, nil
}

func (v *Vault) Authenticate(ctx context.Context, token, role string) (string, error) {
	if token != "" {
		v.client.SetToken(token)
		return token, nil
	}

	auth, err := getAuth(role)
	if err != nil {
		return "", err
	}

	secret, err := v.client.Auth().Login(ctx, auth)
	if err != nil {
		return "", err
	}

	return secret.TokenID()
}

func getAuth(role string) (api.AuthMethod, error) {
	if _, err := os.Stat(k8sTokenFile); !os.IsNotExist(err) {
		return kubernetes.NewKubernetesAuth(role)
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	return aws.NewAWSAuth(aws.WithRegion(region), aws.WithRole(role))
}

func (v *Vault) Load(path string) (map[string]string, error) {
	vars := map[string]string{}
	client := v.client.Logical()
	secret, err := client.List(path)
	if err != nil {
		return vars, err
	}

	for _, key := range keys(secret) {
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

func keys(s *api.Secret) []string {
	if s == nil {
		return []string{}
	}

	k, ok := s.Data["keys"].([]interface{})
	if !ok {
		return []string{}
	}

	return lo.Map[interface{}, string](k, func(x interface{}, i int) string {
		return x.(string)
	})
}
