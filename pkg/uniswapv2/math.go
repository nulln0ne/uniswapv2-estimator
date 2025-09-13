// Package uniswapv2 provides pure math utilities for Uniswap V2 style AMMs.
package uniswapv2

import "math/big"

// fee: 0.3% => multiplier 997/1000
var (
	feeMul = big.NewInt(997)
	feeDen = big.NewInt(1000)
)

// GetAmountOut computes the output amount for a constant-product AMM swap
// using the Uniswap V2 formula with a 0.3% fee (997/1000).
//
// To minimize allocations, callers must provide three scratch big.Int values
// (dst, t1, t2) that are reused internally. The result is returned in dst and
// also as the function's return value.
//
// The formula is:
//
//	amountInWithFee = amountIn * 997
//	numerator       = amountInWithFee * reserveOut
//	denominator     = reserveIn*1000 + amountInWithFee
//	amountOut       = numerator / denominator
func GetAmountOut(dst, t1, t2 *big.Int, amountIn, reserveIn, reserveOut *big.Int) *big.Int {
	// t1 = amountIn * 997
	t1.Mul(amountIn, feeMul)
	// dst = reserveIn * 1000 (will use as denominator temp)
	dst.Mul(reserveIn, feeDen)
	// t2 = dst + t1  (denominator in t2; avoids z==y alias later)
	t2.Add(dst, t1)
	// dst = t1 * reserveOut (numerator)
	dst.Mul(t1, reserveOut)
	// dst = dst / t2 using QuoRem; remainder goes into t1
	dst.QuoRem(dst, t2, t1)
	return dst
}
