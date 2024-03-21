package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/samber/lo"
)

type (
	serviceConfig struct {
		Dependencies struct {
			EnvVars struct {
				Required []dependency `json:"required"`
				Optional []dependency `json:"optional"`
			} `json:"env_vars"`
		} `json:"dependencies"`
	}
	dependency struct {
		dependencyInner
		Partial bool `json:"-"`
	}
	dependencyInner struct {
		Key     string   `json:"key"`
		Regions []string `json:"regions"`
	}
)

var ErrMissingEnvVars = errors.New("missing required environment variables")

// Required returns true if the dependency is required for the given region
func (d *dependency) Required(region string) bool {
	return d.Regions == nil || lo.Contains(d.Regions, region)
}

// UnmarshalJSON handles the dependency being a string or an object
func (d *dependency) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		d.Key = str
		d.Partial = true
		return nil
	}

	var dep dependencyInner
	if err := json.Unmarshal(data, &dep); err != nil {
		return fmt.Errorf("could not decode dependency: %w", err)
	}

	d.dependencyInner = dep
	return nil
}

func validate(ctx context.Context, c *Config, e *EnvMap, l *slog.Logger) error {
	f, err := os.ReadFile(c.ServiceDefinition)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("could not read service definition: %w", err)
	}

	var cfg serviceConfig
	if err := json.Unmarshal(f, &cfg); err != nil {
		return fmt.Errorf("could not decode service definition: %w", err)
	}

	req := missing(cfg.Dependencies.EnvVars.Required, c, e)
	opt := missing(cfg.Dependencies.EnvVars.Optional, c, e)

	if len(opt) != 0 {
		l.WarnContext(ctx, "Missing optional environment variables", "env_vars", opt)
	}

	if len(req) != 0 {
		l.ErrorContext(ctx, "Missing required environment variables", "env_vars", req)
		return ErrMissingEnvVars
	}

	return nil
}

func missing(deps []dependency, c *Config, e *EnvMap) []string {
	res := []string{}
	for _, d := range deps {
		if !d.Required(c.Region) {
			continue
		}

		if v := os.Getenv(d.Key); v == "" && !e.Has(d.Key) {
			res = append(res, d.Key)
		}
	}
	return res
}
