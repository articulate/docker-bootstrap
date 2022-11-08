package main

import (
	"fmt"
	"os"
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
// non-empty environment variables
func (e *EnvMap) Environ() []string {
	env := []string{}
	for k, v := range e.env {
		if x := os.Getenv(k); x == "" {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return env
}
