package uniswapv2

import "math/big"

// fee: 0.3% => multiplier 997/1000
var (
	feeMul = big.NewInt(997)
	feeDen = big.NewInt(1000)
)

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
