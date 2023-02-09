package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
)

// EnvMap represents a map for environment variables
type EnvMap struct {
	env Dict
}

// NewEnvMap returns a new EnvMap
func NewEnvMap() *EnvMap {
	return &EnvMap{env: make(Dict)}
}

// Merge adds the given map to the existing values, overwriting any existing values
func (e *EnvMap) Merge(kv map[string]string) {
	for k, v := range kv {
		e.Add(k, v)
	}
}

// Add adds the given value to the map, overwriting any existing values
func (e *EnvMap) Add(key, value string) {
	e.env[key] = value
}

// Environ returns the map in the format of "key=value", skipping any already set,
// non-empty environment variables, and expanding variables
func (e *EnvMap) Environ() []string {
	// Remove anything already set as an env var
	env := lo.OmitBy(e.env, func(k, _ string) bool {
		return os.Getenv(k) != ""
	})

	// Remove blank keys
	env = lo.OmitBy(env, func(k, _ string) bool {
		return strings.TrimSpace(k) == ""
	})

	// Expand variables
	env = lo.MapValues(env, func(v string, _ string) string {
		return os.Expand(v, func(s string) string {
			if l := os.Getenv(s); l != "" {
				return l
			}
			if v, ok := e.env[s]; ok {
				return v
			}
			return ""
		})
	})

	return lo.MapToSlice(env, func(k string, v string) string {
		return fmt.Sprintf("%s=%s", k, v)
	})
}
