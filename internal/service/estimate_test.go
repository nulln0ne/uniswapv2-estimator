package service

import (
	"context"
	"log/slog"
	"math/big"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
)

type fakeEth struct {
	blockNumber uint64
	// storage[address][positionHash] = 32-byte value
	storage map[common.Address]map[common.Hash][]byte
}

func (f *fakeEth) BlockNumber(ctx context.Context) (hexutil.Uint64, error) {
	return hexutil.Uint64(f.blockNumber), nil
}

func (f *fakeEth) GetStorageAt(ctx context.Context, addr common.Address, position common.Hash, _ gethrpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	if m, ok := f.storage[addr]; ok {
		if v, ok2 := m[position]; ok2 {
			return hexutil.Bytes(v), nil
		}
	}
	// default empty 32 bytes
	return hexutil.Bytes(make([]byte, 32)), nil
}

func newInprocEthClient(t *testing.T, fe *fakeEth) *ethclient.Client {
	t.Helper()
	srv := gethrpc.NewServer()
	// Register under the standard "eth" namespace so methods map to eth_*
	if err := srv.RegisterName("eth", fe); err != nil {
		t.Fatalf("register rpc service: %v", err)
	}
	c := gethrpc.DialInProc(srv)
	return ethclient.NewClient(c)
}

func u256Bytes(v *big.Int) []byte {
	b := v.Bytes()
	if len(b) > 32 {
		panic("value does not fit in 32 bytes")
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

func packReserves(r0, r1 uint64, ts uint32) []byte {
	v := new(big.Int).SetUint64(uint64(ts))
	v.Lsh(v, 112)
	v.Or(v, new(big.Int).SetUint64(r1))
	v.Lsh(v, 112)
	v.Or(v, new(big.Int).SetUint64(r0))
	return u256Bytes(v)
}

func rightPadAddress(addr common.Address) []byte {
	// Address is right-aligned in 32 bytes when read from storage
	out := make([]byte, 32)
	copy(out[12:], addr.Bytes())
	return out
}

func TestEstimate_Success(t *testing.T) {
	t.Parallel()

	token0 := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	token1 := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	pool := common.HexToAddress("0x0000000000000000000000000000000000000abc")

	// reserves: 1_000_000 : 2_000_000
	r0, r1 := uint64(1_000_000), uint64(2_000_000)
	amountIn := big.NewInt(1_000)

	fe := &fakeEth{
		blockNumber: 123,
		storage: map[common.Address]map[common.Hash][]byte{
			pool: {
				common.BigToHash(new(big.Int).SetUint64(6)): rightPadAddress(token0),
				common.BigToHash(new(big.Int).SetUint64(7)): rightPadAddress(token1),
				common.BigToHash(new(big.Int).SetUint64(8)): packReserves(r0, r1, 0),
			},
		},
	}
	ec := newInprocEthClient(t, fe)

	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	svc := NewEstimateService(logger, *ec)

	out, err := svc.Estimate(context.Background(), pool, token0, token1, amountIn)
	if err != nil {
		t.Fatalf("Estimate error: %v", err)
	}

	// compute expected
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(amountInWithFee, new(big.Int).SetUint64(r1))
	denominator := new(big.Int).Add(new(big.Int).Mul(new(big.Int).SetUint64(r0), big.NewInt(1000)), amountInWithFee)
	expected := new(big.Int).Div(numerator, denominator)

	if out.Cmp(expected) != 0 {
		t.Fatalf("unexpected amountOut: got %s want %s", out, expected)
	}
}

func TestEstimate_PairMismatch(t *testing.T) {
	t.Parallel()

	token0 := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	token1 := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	wrong := common.HexToAddress("0x00000000000000000000000000000000000000cc")
	pool := common.HexToAddress("0x0000000000000000000000000000000000000abc")

	fe := &fakeEth{
		blockNumber: 1,
		storage: map[common.Address]map[common.Hash][]byte{
			pool: {
				common.BigToHash(new(big.Int).SetUint64(6)): rightPadAddress(token0),
				common.BigToHash(new(big.Int).SetUint64(7)): rightPadAddress(token1),
				common.BigToHash(new(big.Int).SetUint64(8)): packReserves(1, 1, 0),
			},
		},
	}
	ec := newInprocEthClient(t, fe)
	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	svc := NewEstimateService(logger, *ec)

	_, err := svc.Estimate(context.Background(), pool, token0, wrong, big.NewInt(1))
	if err == nil || err != ErrPairMismatch {
		t.Fatalf("expected ErrPairMismatch, got %v", err)
	}
}

func TestEstimate_SameToken(t *testing.T) {
	t.Parallel()

	token := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	pool := common.HexToAddress("0x0000000000000000000000000000000000000abc")

	fe := &fakeEth{blockNumber: 1, storage: map[common.Address]map[common.Hash][]byte{pool: {common.BigToHash(new(big.Int).SetUint64(6)): rightPadAddress(token), common.BigToHash(new(big.Int).SetUint64(7)): rightPadAddress(token), common.BigToHash(new(big.Int).SetUint64(8)): packReserves(1, 1, 0)}}}
	ec := newInprocEthClient(t, fe)
	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	svc := NewEstimateService(logger, *ec)

	_, err := svc.Estimate(context.Background(), pool, token, token, big.NewInt(1))
	if err == nil || err != ErrSameToken {
		t.Fatalf("expected ErrSameToken, got %v", err)
	}
}

func TestEstimate_EmptyReserves(t *testing.T) {
	t.Parallel()

	token0 := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	token1 := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	pool := common.HexToAddress("0x0000000000000000000000000000000000000abc")

	fe := &fakeEth{blockNumber: 1, storage: map[common.Address]map[common.Hash][]byte{pool: {common.BigToHash(new(big.Int).SetUint64(6)): rightPadAddress(token0), common.BigToHash(new(big.Int).SetUint64(7)): rightPadAddress(token1), common.BigToHash(new(big.Int).SetUint64(8)): packReserves(0, 0, 0)}}}
	ec := newInprocEthClient(t, fe)
	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	svc := NewEstimateService(logger, *ec)

	_, err := svc.Estimate(context.Background(), pool, token0, token1, big.NewInt(1))
	if err == nil || err != ErrEmptyReserves {
		t.Fatalf("expected ErrEmptyReserves, got %v", err)
	}
}
