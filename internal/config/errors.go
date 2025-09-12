package config

import "errors"

var (
	ErrMissingRPCEndpoint = errors.New("missing ETH_RPC_URL environment variable")
)
