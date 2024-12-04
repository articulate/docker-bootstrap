package main

import "log/slog"

// structuredError is an error that can be logged with additional data
type structuredError struct {
	err  error
	args []any
}

// NewStructuredError returns a new structure error that will be logged with additional data
func serror(err error, args ...any) *structuredError {
	return &structuredError{err: err, args: args}
}

func (e *structuredError) Error() string {
	return e.err.Error()
}

func (e *structuredError) Unwrap() error {
	return e.err
}

// LogValue returns a structured log value for use with slog
func (e *structuredError) LogValue() slog.Value {
	args := make([]slog.Attr, 0, (len(e.args)/2)+1)
	args = append(args, slog.String(slog.MessageKey, e.Error()))

	var c interface{}
	for _, a := range e.args {
		if c == nil {
			c = a
			continue
		}
		args = append(args, slog.Any(c.(string), a))
		c = nil
	}

	return slog.GroupValue(args...)
}
