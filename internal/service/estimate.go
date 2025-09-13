package service

import (
	"context"
	"fmt"
	"math/big"

	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nulln0ne/uniswap-estimator/pkg/uniswapv2"
)

// EstimateService provides Uniswap V2 output amount estimations by reading
// on-chain pair storage directly.
type EstimateService struct {
	BaseService
	ethereumClient *ethclient.Client
}

// NewEstimateService constructs an EstimateService using the provided logger
// and Ethereum client.
func NewEstimateService(logger *slog.Logger, ec ethclient.Client) *EstimateService {
	return &EstimateService{
		BaseService:    BaseService{logger: logger},
		ethereumClient: &ec,
	}
}

// contract UniswapV2Pair is IUniswapV2Pair, UniswapV2ERC20 {
//     using SafeMath  for uint;
//     using UQ112x112 for uint224;
//
//     uint public constant MINIMUM_LIQUIDITY = 10**3;
//     bytes4 private constant SELECTOR = bytes4(keccak256(bytes('transfer(address,uint256)')));
//
//     address public factory;
//     address public token0;
//     address public token1;
//
//     uint112 private reserve0;           // uses single storage slot, accessible via getReserves
//     uint112 private reserve1;           // uses single storage slot, accessible via getReserves
//     uint32  private blockTimestampLast; // uses single storage slot, accessible via getReserves

// Estimate computes the expected output amount for swapping amountIn of src to
// dst in the provided pool at the latest block. It validates the token pair,
// reads reserves from storage and applies the Uniswap V2 formula.
func (e *EstimateService) Estimate(ctx context.Context, pool, src, dst common.Address, amountIn *big.Int) (*big.Int, error) {
	e.logger.Debug("estimating swap", "pool", pool.Hex(), "src", src.Hex(), "dst", dst.Hex(), "in", amountIn.String())

	if src == dst {
		return nil, ErrSameToken
	}

	bn, err := e.ethereumClient.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("block number: %w", err)
	}
	blockNum := new(big.Int).SetUint64(bn)

	token0, token1, err := e.loadTokens(ctx, pool, blockNum)
	if err != nil {
		return nil, err
	}

	// reserves (uint112 | uint112 | uint32) are packed into a single 32‑byte slot (slot 8)
	br, err := e.readSlot(ctx, pool, blockNum, 8)
	if err != nil {
		return nil, err
	}
	reserve0, reserve1 := parseReserves(br)

	var reserveIn, reserveOut *big.Int
	switch {
	case src == token0 && dst == token1:
		reserveIn, reserveOut = reserve0, reserve1
	case src == token1 && dst == token0:
		reserveIn, reserveOut = reserve1, reserve0
	default:
		return nil, ErrPairMismatch
	}

	if reserveIn.Sign() == 0 || reserveOut.Sign() == 0 {
		return nil, ErrEmptyReserves
	}

	var outAmt, tmp1, tmp2 big.Int
	out := uniswapv2.GetAmountOut(&outAmt, &tmp1, &tmp2, amountIn, reserveIn, reserveOut)
	e.logger.Debug("amount out computed", "out", out.String())
	return out, nil
}

func (e *EstimateService) readSlot(ctx context.Context, pool common.Address, blockNum *big.Int, slot uint64) ([]byte, error) {
	key := common.BigToHash(new(big.Int).SetUint64(slot))
	b, err := e.ethereumClient.StorageAt(ctx, pool, key, blockNum)
	if err != nil {
		return nil, fmt.Errorf("storageAt slot %d (pool %s, block %s): %w",
			slot, pool.Hex(), blockNum.String(), err)
	}
	return b, nil
}

// loadTokens reads token0 and token1 from Uniswap V2 pair storage (slots 6 and 7).
func (e *EstimateService) loadTokens(ctx context.Context, pool common.Address, blockNum *big.Int) (common.Address, common.Address, error) {
	b0, err := e.readSlot(ctx, pool, blockNum, 6)
	if err != nil {
		return common.Address{}, common.Address{}, err
	}
	token0 := common.BytesToAddress(b0)

	b1, err := e.readSlot(ctx, pool, blockNum, 7)
	if err != nil {
		return common.Address{}, common.Address{}, err
	}
	token1 := common.BytesToAddress(b1)

	return token0, token1, nil
}

// parseReserves unpacks two uint112 reserves from the 32‑byte storage word
// used by Uniswap V2 pairs. The layout is:
//
//	[ 112 bits reserve0 | 112 bits reserve1 | 32 bits timestamp ]
//
// Values are treated as big‑endian within the 256‑bit word.
func parseReserves(b []byte) (reserve0, reserve1 *big.Int) {
	v := new(big.Int).SetBytes(b)
	one := big.NewInt(1)
	mask112 := new(big.Int).Sub(new(big.Int).Lsh(one, 112), one)

	reserve0 = new(big.Int).And(v, mask112)
	tmp := new(big.Int).Rsh(v, 112)
	reserve1 = new(big.Int).And(tmp, mask112)
	return
}
