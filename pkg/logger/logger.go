package logger

import (
	"log/slog"
	"os"
)

// New creates a JSON logger writing to stdout.
//
// Why JSON:
// - easy to read in Docker logs
// - easy to parse by log aggregation tools later
func New() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}
