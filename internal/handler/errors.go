package handler

import "github.com/gofiber/fiber/v3"

var (
	ErrInvalidQueryParameters   = fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	ErrSameAddresses            = fiber.NewError(fiber.StatusBadRequest, "src and dst addresses cannot be the same")
	ErrAmountRequired           = fiber.NewError(fiber.StatusBadRequest, "amount is required")
	ErrInvalidAmountFormat      = fiber.NewError(fiber.StatusBadRequest, "invalid amount format")
	ErrAmountNonPositive        = fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	ErrSameTokenBadRequest      = fiber.NewError(fiber.StatusBadRequest, "src and dst tokens cannot be the same")
	ErrEmptyReservesBadRequest  = fiber.NewError(fiber.StatusBadRequest, "pool has insufficient reserves")
	ErrEstimationFailedInternal = fiber.NewError(fiber.StatusInternalServerError, "estimation failed")
)

func NewInvalidAmountIn(err error) error {
	return fiber.NewError(fiber.StatusBadRequest, "invalid amount_in: "+err.Error())
}

func NewAddressRequired(field string) error {
	return fiber.NewError(fiber.StatusBadRequest, field+" address is required")
}

func NewInvalidAddress(field string) error {
	return fiber.NewError(fiber.StatusBadRequest, "invalid "+field+" address")
}
