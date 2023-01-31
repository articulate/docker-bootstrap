package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
)

// Consul is a client for fetching values from Consul's KV store
type Consul struct {
	kv *api.KV
}

// NewConsul returns a new instance of the Consul client using the given address
func NewConsul(addr string) (*Consul, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr

	c, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not connect to %s: %w", addr, err)
	}

	return &Consul{c.KV()}, nil
}

// Load fetches values from the given paths
func (c *Consul) Load(path string) (Dict, error) {
	kv := make(Dict)
	pairs, _, err := c.kv.List(path, &api.QueryOptions{})
	if err != nil {
		return kv, fmt.Errorf("could not load %s: %w", path, err)
	}

	for _, p := range pairs {
		if p.Value == nil {
			continue
		}
		kv[consulKey(p.Key)] = string(p.Value)
	}

	return kv, nil
}

// consulKey converts the full path to the env var name
func consulKey(path string) string {
	parts := strings.Split(path, "/")
	key := parts[len(parts)-1]
	return strings.ToUpper(key)
}
