package uniswapv2

import (
	"math/big"
	"testing"
)

func BenchmarkGetAmountOut_NoAlloc(b *testing.B) {
	rIn := new(big.Int).SetUint64(13_451_234_567_890)
	rOut := new(big.Int).SetUint64(98_765_432_109_876)
	in := new(big.Int).SetUint64(1_000_000)
	dst := new(big.Int)
	t1 := new(big.Int)
	t2 := new(big.Int)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetAmountOut(dst, t1, t2, in, rIn, rOut)
	}
}
