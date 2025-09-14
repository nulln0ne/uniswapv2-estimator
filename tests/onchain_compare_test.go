package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	gethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/nulln0ne/uniswap-estimator/pkg/uniswapv2"
)

// TestGetAmountOut_Onchain compares our math implementation to Uniswap V2 Router02's
// getAmountOut via an on-chain eth_call. Skips if ETH_RPC_URL is not set.
func TestGetAmountOut_Onchain(t *testing.T) {
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		t.Skip("ETH_RPC_URL not set; skipping on-chain comparison test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		t.Fatalf("dial eth rpc: %v", err)
	}

	// Load ABI for Uniswap V2 Router02 from tests/data/abi.json
	abiPath := filepath.Join("data", "abi.json")
	data, err := os.ReadFile(abiPath)
	if err != nil {
		t.Fatalf("read abi: %v", err)
	}
	// The file contains a JSON array; accounts/abi expects a single contract object.
	// Build an ABI by marshaling methods of interest into a Contract object.
	var arr []map[string]any
	if err := json.Unmarshal(data, &arr); err != nil {
		t.Fatalf("parse abi json: %v", err)
	}
	// Re-marshal only the getAmountOut entry to a minimal ABI that geth can parse.
	// This avoids relying on the full Router ABI shape and keeps the test focused.
	var methodEntries []map[string]any
	for _, e := range arr {
		if name, _ := e["name"].(string); name == "getAmountOut" {
			methodEntries = append(methodEntries, e)
		}
	}
	if len(methodEntries) == 0 {
		t.Fatalf("getAmountOut not found in ABI file")
	}
	minimalJSON, err := json.Marshal(methodEntries)
	if err != nil {
		t.Fatalf("marshal minimal abi: %v", err)
	}
	contractABI, err := gethabi.JSON(bytes.NewReader(minimalJSON))
	if err != nil {
		t.Fatalf("parse abi: %v", err)
	}

	// Uniswap V2 Router02 mainnet address
	router := common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")

	// Test vectors (amountIn, reserveIn, reserveOut)
	cases := []struct {
		name       string
		amountIn   *big.Int
		reserveIn  *big.Int
		reserveOut *big.Int
	}{
		{"small_balanced", big.NewInt(1_000), big.NewInt(1_000_000), big.NewInt(1_000_000)},
		{"skewed_reserves", big.NewInt(50_000_000_000_000), new(big.Int).SetUint64(5_000_000_000_000_000), new(big.Int).SetUint64(100_000_000_000_000_000)},
		{"large_values", new(big.Int).SetUint64(1_000_000_000_000_000), new(big.Int).SetUint64(50_000_000_000_000_000), new(big.Int).SetUint64(75_000_000_000_000_000)},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Local math
			var dst, t1, t2 big.Int
			local := uniswapv2.GetAmountOut(&dst, &t1, &t2, tc.amountIn, tc.reserveIn, tc.reserveOut)

			// Encode and perform eth_call to router.getAmountOut(amountIn, reserveIn, reserveOut)
			input, err := contractABI.Pack("getAmountOut", tc.amountIn, tc.reserveIn, tc.reserveOut)
			if err != nil {
				t.Fatalf("abi pack: %v", err)
			}

			call := ethereum.CallMsg{To: &router, Data: input}
			out, err := client.CallContract(ctx, call, nil)
			if err != nil {
				t.Fatalf("eth_call getAmountOut: %v", err)
			}
			// Unpack return value
			values, err := contractABI.Unpack("getAmountOut", out)
			if err != nil {
				t.Fatalf("abi unpack: %v", err)
			}
			if len(values) != 1 {
				t.Fatalf("unexpected outputs: %d", len(values))
			}
			onchain, ok := values[0].(*big.Int)
			if !ok {
				t.Fatalf("unexpected output type: %T", values[0])
			}

			if local.Cmp(onchain) != 0 {
				t.Fatalf("mismatch: local=%s onchain=%s (in=%s rIn=%s rOut=%s)", local, onchain, tc.amountIn, tc.reserveIn, tc.reserveOut)
			}
		})
	}
}
