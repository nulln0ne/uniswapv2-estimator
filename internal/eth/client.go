// Package eth contains thin wrappers around Ethereum client functionality.
package eth

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Dial connects to the Ethereum RPC endpoint at the given URL using the
// provided context. A 15-second timeout is applied to the dialing process.
func Dial(ctx context.Context, url string) (*ethclient.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	return ethclient.DialContext(ctx, url)
}
