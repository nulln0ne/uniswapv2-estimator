package handler

import (
	"context"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/gofiber/fiber/v3"
	"github.com/nulln0ne/uniswap-estimator/internal/service"
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

func TestEstimateHandler_OK(t *testing.T) {
	token0 := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	token1 := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	pool := common.HexToAddress("0x0000000000000000000000000000000000000abc")

	fe := &fakeEth{blockNumber: 42, storage: map[common.Address]map[common.Hash][]byte{pool: {common.BigToHash(new(big.Int).SetUint64(6)): rightPadAddress(token0), common.BigToHash(new(big.Int).SetUint64(7)): rightPadAddress(token1), common.BigToHash(new(big.Int).SetUint64(8)): packReserves(1_000_000, 2_000_000, 0)}}}
	ec := newInprocEthClient(t, fe)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewEstimateService(logger, *ec)
	h := NewEstimateHandler(logger, svc)

	app := fiber.New()
	app.Get("/estimate", h.Handle())

	req := httptest.NewRequest(http.MethodGet, "/estimate?pool="+pool.Hex()+"&src="+token0.Hex()+"&dst="+token1.Hex()+"&src_amount=1000", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestEstimateHandler_Validation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fe := &fakeEth{blockNumber: 1, storage: map[common.Address]map[common.Hash][]byte{}}
	ec := newInprocEthClient(t, fe)
	svc := service.NewEstimateService(logger, *ec)
	h := NewEstimateHandler(logger, svc)

	app := fiber.New()
	app.Get("/estimate", h.Handle())

	req := httptest.NewRequest(http.MethodGet, "/estimate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing params, got %d", resp.StatusCode)
	}
}
