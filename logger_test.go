package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredError(t *testing.T) {
	wrapped := fmt.Errorf("test error: %w", os.ErrNotExist)
	err := serror(wrapped, "foo", "bar")

	require.ErrorContains(t, err, "test error")
	require.ErrorIs(t, err, os.ErrNotExist)

	l, log := testLogger()
	l.Error("test", "error", err)
	assert.Contains(t, log.String(), `"error":{"msg":"test error: file does not exist","foo":"bar"}`)
}

func testLogger() (*slog.Logger, *bytes.Buffer) {
	log := &bytes.Buffer{}
	return slog.New(slog.NewJSONHandler(log, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				a.Value = slog.StringValue("test-time")
			}
			return a
		},
	})), log
}
