package main

import (
	"context"
	"fmt"
	"log/slog"
)

type (
	Dict map[string]string

	// Client represents a client used to load values
	Client interface {
		Load(string) (Dict, error)
	}
)

func loadValues(ctx context.Context, c Client, l *slog.Logger, paths []string) (Dict, error) {
	values := map[string]string{}
	for _, path := range paths {
		l.DebugContext(ctx, "Loading values", "path", path)

		kv, err := c.Load(path)
		if err != nil {
			return values, serror(fmt.Errorf("Could not load values: %w", err), "path", path)
		}

		for k, v := range kv {
			values[k] = v
		}
	}
	return values, nil
}
