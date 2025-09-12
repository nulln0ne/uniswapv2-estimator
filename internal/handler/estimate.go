package handler

import (
	"context"
	"math/big"

	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v3"
	"github.com/nulln0ne/uniswap-estimator/internal/service"
)

type EstimateHandler struct {
	BaseHandler
	service *service.EstimateService
}

func NewEstimateHandler(logger *slog.Logger, svc *service.EstimateService) *EstimateHandler {
	return &EstimateHandler{
		BaseHandler: BaseHandler{
			logger: logger,
		},
		service: svc,
	}
}

type EstimateRequest struct {
	Pool     string `query:"pool" json:"pool"`
	Src      string `query:"src" json:"src"`
	Dst      string `query:"dst" json:"dst"`
	AmountIn string `query:"src_amount" json:"amount_in"`
}

func (h *EstimateHandler) Handle() fiber.Handler {
	return func(c fiber.Ctx) error {
		req, err := h.parseAndValidateRequest(c)
		if err != nil {
			return err
		}

		pool := common.HexToAddress(req.Pool)
		src := common.HexToAddress(req.Src)
		dst := common.HexToAddress(req.Dst)

		amountIn, err := h.parseAmount(req.AmountIn)
		if err != nil {
			return NewInvalidAmountIn(err)
		}

		amountOut, err := h.service.Estimate(context.Background(), pool, src, dst, amountIn)
		if err != nil {
			return h.handleServiceError(err)
		}

		h.logger.Debug("estimate computed", "pool", req.Pool, "src", req.Src, "dst", req.Dst, "in", amountIn.String(), "out", amountOut.String())
		return c.SendString(amountOut.String())
	}
}

func (h *EstimateHandler) parseAndValidateRequest(c fiber.Ctx) (*EstimateRequest, error) {
	var req EstimateRequest

	if err := c.Bind().Query(&req); err != nil {
		h.logger.Debug("failed to bind query parameters", "err", err)
		return nil, ErrInvalidQueryParameters
	}

	if err := h.validateAddresses(&req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (h *EstimateHandler) validateAddresses(req *EstimateRequest) error {
	addresses := map[string]string{
		"pool": req.Pool,
		"src":  req.Src,
		"dst":  req.Dst,
	}

	for field, addr := range addresses {
		if addr == "" {
			return NewAddressRequired(field)
		}
		if !common.IsHexAddress(addr) {
			return NewInvalidAddress(field)
		}
	}

	if req.Src == req.Dst {
		return ErrSameAddresses
	}

	return nil
}

func (h *EstimateHandler) parseAmount(amountStr string) (*big.Int, error) {
	if amountStr == "" {
		return nil, ErrAmountRequired
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return nil, ErrInvalidAmountFormat
	}

	if amount.Sign() <= 0 {
		return nil, ErrAmountNonPositive
	}

	return amount, nil
}

func (h *EstimateHandler) handleServiceError(err error) error {
	switch err {
	case service.ErrSameToken:
		return ErrSameTokenBadRequest
	case service.ErrEmptyReserves:
		return ErrEmptyReservesBadRequest
	default:
		h.logger.Error("service estimate failed", "err", err)
		return ErrEstimationFailedInternal
	}
}
