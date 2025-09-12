package uniswapv2

import (
	"math/big"
	"testing"
)

func TestGetAmountOut(t *testing.T) {
	// Example: reserves 1000000 : 1000000, amountIn 1000
	rIn := big.NewInt(1_000_000)
	rOut := big.NewInt(1_000_000)
	amountIn := big.NewInt(1_000)

	// dst/t1/t2 are re-used temporaries
	var dst, t1, t2 big.Int
	out := GetAmountOut(&dst, &t1, &t2, amountIn, rIn, rOut)

	// compute expected using same formula to assert determinism and non-zero
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(amountInWithFee, rOut)
	denominator := new(big.Int).Mul(rIn, big.NewInt(1000))
	denominator.Add(denominator, amountInWithFee)
	expected := new(big.Int).Div(numerator, denominator)

	if out.Cmp(expected) != 0 {
		t.Fatalf("unexpected: got %s want %s", out, expected)
	}
	if out.Sign() <= 0 {
		t.Fatalf("amountOut should be positive")
	}
}
