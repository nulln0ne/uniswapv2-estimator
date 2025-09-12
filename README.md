# Uniswap V2 Estimator

A small Go HTTP service that estimates Uniswap V2 swap output without calling the smart contract directly. It reads the pair state via `eth_getStorageAt` (token0, token1 and packed reserves) and applies the Uniswap V2 formula (`x*y=k` with a 0.3% fee).

## Build & Run
```bash
# Build
make build

# Environment
export ETH_RPC_URL="https://mainnet.infura.io/v3/<your-key>"
export ADDR=":1337"           # optional, default :1337
export LOG_LEVEL="info"       # debug|info|warn|error (default info)

# Run
./bin/uniswap-estimator
```

## API
- Method: `GET /estimate`
- Query parameters:
  - `pool` — Uniswap V2 pair address (`0x...`).
  - `src` — input token address (`0x...`).
  - `dst` — output token address (`0x...`).
  - `src_amount` — input amount in raw token units (no decimals), decimal string.
- Response: plain‑text decimal `amountOut` string.

Example:
```bash
curl "http://localhost:1337/estimate?pool=0xA..&src=0xB..&dst=0xC..&src_amount=1000000000000000000"
# -> 123456789012345678
```

## How It Works
- Reads Uniswap V2 pair storage slots:
  - slot `6` — `token0`
  - slot `7` — `token1`
  - slot `8` — packed word: `uint112 reserve0 | uint112 reserve1 | uint32 blockTimestampLast`
- Unpacks the two `uint112` reserves.
- Picks `reserveIn/reserveOut` according to direction (`src` → `dst`).
- Computes `amountOut` using the Uniswap V2 formula (0.3% fee).

## Benchmarks
Result from `make bench` on Apple M1 Pro (darwin/arm64):
```
goos: darwin
goarch: arm64
pkg: github.com/nulln0ne/uniswap-estimator/pkg/uniswapv2
cpu: Apple M1 Pro
BenchmarkGetAmountOut_NoAlloc-8      20183210            59.39 ns/op            8 B/op          1 allocs/op
PASS
ok      github.com/nulln0ne/uniswap-estimator/pkg/uniswapv2    2.708s
```