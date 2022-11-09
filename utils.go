package main

import "github.com/rs/zerolog"

type (
	Dict map[string]string

	// Client represents a client used to load values
	Client interface {
		Load(string) (Dict, error)
	}
)

func loadValues(c Client, l zerolog.Logger, paths []string) Dict {
	values := map[string]string{}
	for _, path := range paths {
		l.Debug().Str("path", path).Msg("Loading values")

		kv, err := c.Load(path)
		if err != nil {
			l.Fatal().Err(err).Str("path", path).Msg("Could not load values")
		}

		for k, v := range kv {
			values[k] = v
		}
	}
	return values
}
