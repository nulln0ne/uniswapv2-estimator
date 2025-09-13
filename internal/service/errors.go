package service

import "errors"

// ErrSameToken indicates src and dst token addresses are equal.
var ErrSameToken = errors.New("src and dst are equal")

// ErrPairMismatch indicates the provided src/dst tokens do not match the
// pool's token0/token1 pair.
var ErrPairMismatch = errors.New("pair does not match src/dst")

// ErrEmptyReserves indicates one or both reserves are zero for the pool.
var ErrEmptyReserves = errors.New("empty reserves")
