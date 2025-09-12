package eth

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

func Dial(ctx context.Context, url string) (*ethclient.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	return ethclient.DialContext(ctx, url)
}
