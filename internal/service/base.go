// Package service contains business logic and integrations backing HTTP handlers.
package service

import "log/slog"

// BaseService provides common dependencies for service types.
type BaseService struct {
	logger *slog.Logger
}
