package config

import "os"

type Config struct {
	Addr        string
	RPCEndpoint string
	LogLevel    string
}

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
