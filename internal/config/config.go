// Package config loads and validates service configuration from environment
// variables.
package config

import "os"

// Config holds runtime configuration values for the service.
type Config struct {
	Addr        string
	RPCEndpoint string
	LogLevel    string
}

// FromEnv reads configuration from environment variables and returns a
// populated Config. It applies sensible defaults for optional settings and
// validates required values.
//
// Required:
//   - ETH_RPC_URL: Ethereum node RPC URL
//
// Optional:
//   - ADDR (default ":1337"): listen address for the HTTP server
//   - LOG_LEVEL (default "info"): one of debug, info, warn, error
func FromEnv() (*Config, error) {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":1337"
	}

	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		return nil, ErrMissingRPCEndpoint
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	cfg := &Config{
		Addr:        addr,
		RPCEndpoint: rpcURL,
		LogLevel:    logLevel,
	}

	return cfg, nil
}
