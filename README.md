# Uniswap V2 Estimator

> A lightweight Go HTTP service for estimating Uniswap V2 swap outputs without direct smart contract calls

## Overview

This service estimates Uniswap V2 swap outputs by reading pair state directly via `eth_getStorageAt` calls (token0, token1, and packed reserves) and applying the Uniswap V2 AMM formula (`x*y=k` with 0.3% fee).

## Quick Start

### Environment Configuration

Create a `.env` file based on the example:

```env
ADDR=:1337
ETH_RPC_URL=https://mainnet.infura.io/v3/YOUR_PROJECT_ID
LOG_LEVEL=info # debug, info, warn, error (default: info)
```

### Build & Run

```bash
# Build the application
make build

# Run the service
./bin/uniswap-estimator
```

## API Reference

### Estimate Swap Output

**Endpoint:** `GET /estimate`

**Query Parameters:**
- `pool` **(required)** — Uniswap V2 pair address (format: `0x...`)
- `src` **(required)** — Input token address (format: `0x...`)
- `dst` **(required)** — Output token address (format: `0x...`)
- `src_amount` **(required)** — Input amount in raw token units (decimal string, no decimals applied)

**Response:** Plain-text decimal string representing `amountOut`

### Example Usage

```bash
curl "http://localhost:1337/estimate?pool=0xA43fe16908251ee70EF74718545e4FE6C5cCEc9f&src=0xA0b86a33E6441d0C95D9C02DAd5c8dE47a5e67E8&dst=0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2&src_amount=1000000000000000000"

# Response:
# 123456789012345678
```

## Technical Implementation

### Storage Reading Strategy

The service is directly reading Uniswap V2 contract storage:

| Slot | Content | Description |
|------|---------|-------------|
| `6` | `token0` | First token address in the pair |
| `7` | `token1` | Second token address in the pair |
| `8` | Packed data | `uint112 reserve0 \| uint112 reserve1 \| uint32 blockTimestampLast` |

### Calculation Process

1. **Storage Extraction** — Read and unpack the two `uint112` reserves from slot 8
2. **Direction Mapping** — Determine `reserveIn`/`reserveOut` based on `src` -> `dst` direction
3. **AMM Formula** — Apply Uniswap V2 formula with 0.3% fee deduction

## Performance Benchmarks

Benchmark results on Apple M1 Pro (darwin/arm64):

```
goos: darwin
goarch: arm64
pkg: github.com/nulln0ne/uniswap-estimator/pkg/uniswapv2
cpu: Apple M1 Pro

BenchmarkGetAmountOut_NoAlloc-8    20,183,210    59.39 ns/op    8 B/op    1 allocs/op
```