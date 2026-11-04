package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type (
	Dict map[string]string

	// Client represents a client used to load values
	Client interface {
		Load(string) (Dict, error)
	}

	// origin records which path last set a key, and to what value
	origin struct {
		path  string
		value string
	}
)

// loadValues loads each path in order, later paths overwriting earlier ones.
// sources is shared across calls so overwrites between Consul and Vault are logged too.
func loadValues(
	ctx context.Context, c Client, l *slog.Logger, paths []string, sources map[string]origin,
) (Dict, error) {
	values := map[string]string{}
	for _, path := range paths {
		l.DebugContext(ctx, "Loading values", "path", path)

		kv, err := c.Load(path)
		if err != nil {
			return values, serror(fmt.Errorf("could not load values: %w", err), "path", path)
		}

		for k, v := range kv {
			// Values are never logged; they may be secrets.
			prev, seen := sources[k]
			switch {
			case seen && prev.value != v:
				l.WarnContext(ctx, "Overwriting value", "key", k, "path", path, "previous_path", prev.path)
			case !seen && os.Getenv(k) != "":
				l.WarnContext(ctx, "Host env var takes precedence", "key", k, "path", path)
			}
			sources[k] = origin{path: path, value: v}
			values[k] = v
		}
	}
	return values, nil
}
