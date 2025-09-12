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
	// t2 = reserveIn * 1000
	t2.Mul(reserveIn, feeDen)
	// t2 = t2 + t1  (denominator)
	t2.Add(t2, t1)
	// dst = t1 * reserveOut (numerator)
	dst.Mul(t1, reserveOut)
	// dst = dst / t2  (avoid aliasing z==y)
	return dst.Div(dst, t2)
}
