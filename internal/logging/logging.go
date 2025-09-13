// Package logging provides helpers to construct a configured slog.Logger.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger returns a slog.Logger configured to write text logs to stdout at
// the provided level. Supported levels: debug, info, warn, error.
func NewLogger(level string) *slog.Logger {
	lvl := parseLevel(level)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
