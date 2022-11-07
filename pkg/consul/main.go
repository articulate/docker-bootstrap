package consul

import (
	"net/url"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/samber/lo"
)

type Consul struct {
	kv *api.KV
}

func New(addr string) (*Consul, error) {
	a, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	c, err := api.NewClient(&api.Config{
		Address: a.Host,
		Scheme:  a.Scheme,
	})
	if err != nil {
		return nil, err
	}

	return &Consul{c.KV()}, nil
}

func (c *Consul) Load(path string) (map[string][]byte, error) {
	pairs, _, err := c.kv.List(path, &api.QueryOptions{})
	if err != nil {
		return map[string][]byte{}, err
	}

	return lo.Associate[*api.KVPair, string, []byte](pairs, func(kv *api.KVPair) (string, []byte) {
		return pathToKey(kv.Key), kv.Value
	}), nil
}

func pathToKey(path string) string {
	parts := strings.Split(path, "/")
	key := parts[len(parts)-1]
	return strings.ToUpper(key)
}
