package service

import "errors"

var (
	ErrSameToken     = errors.New("src and dst are equal")
	ErrPairMismatch  = errors.New("pair does not match src/dst")
	ErrEmptyReserves = errors.New("empty reserves")
)
