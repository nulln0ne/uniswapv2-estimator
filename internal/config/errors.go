package config

import "errors"

// ErrMissingRPCEndpoint indicates that the required ETH_RPC_URL variable is
// not set in the environment.
var ErrMissingRPCEndpoint = errors.New("missing ETH_RPC_URL environment variable")
