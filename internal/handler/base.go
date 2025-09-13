// Package handler defines HTTP request handlers and related utilities.
package handler

import "log/slog"

// BaseHandler provides common dependencies for HTTP handlers.
type BaseHandler struct {
	logger *slog.Logger
}
