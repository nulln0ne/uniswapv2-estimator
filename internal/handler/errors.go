package handler

import "github.com/gofiber/fiber/v3"

// ErrInvalidQueryParameters indicates that the request query string could not
// be parsed into the expected structure.
var ErrInvalidQueryParameters = fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")

// ErrSameAddresses is returned when src and dst addresses are identical.
var ErrSameAddresses = fiber.NewError(fiber.StatusBadRequest, "src and dst addresses cannot be the same")

// ErrAmountRequired is returned when the amount parameter is missing.
var ErrAmountRequired = fiber.NewError(fiber.StatusBadRequest, "amount is required")

// ErrInvalidAmountFormat is returned when the amount cannot be parsed as a
// base-10 integer.
var ErrInvalidAmountFormat = fiber.NewError(fiber.StatusBadRequest, "invalid amount format")

// ErrAmountNonPositive is returned when the amount is zero or negative.
var ErrAmountNonPositive = fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")

// ErrSameTokenBadRequest maps a same-token validation failure to a 400 error.
var ErrSameTokenBadRequest = fiber.NewError(fiber.StatusBadRequest, "src and dst tokens cannot be the same")

// ErrEmptyReservesBadRequest maps empty-reserve pool state to a 400 error.
var ErrEmptyReservesBadRequest = fiber.NewError(fiber.StatusBadRequest, "pool has insufficient reserves")

// ErrEstimationFailedInternal signals a generic server-side estimation error.
var ErrEstimationFailedInternal = fiber.NewError(fiber.StatusInternalServerError, "estimation failed")

// NewInvalidAmountIn wraps an amount parsing error into a 400 Bad Request with
// a descriptive message.
func NewInvalidAmountIn(err error) error {
	return fiber.NewError(fiber.StatusBadRequest, "invalid amount_in: "+err.Error())
}

// NewAddressRequired returns a 400 Bad Request for a missing address field.
func NewAddressRequired(field string) error {
	return fiber.NewError(fiber.StatusBadRequest, field+" address is required")
}

// NewInvalidAddress returns a 400 Bad Request for an invalid address format.
func NewInvalidAddress(field string) error {
	return fiber.NewError(fiber.StatusBadRequest, "invalid "+field+" address")
}
