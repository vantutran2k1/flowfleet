package domain

import "errors"

var (
	ErrInvalidTransition = errors.New("invalid status transition: order condition not met")
	ErrOrderNotFound     = errors.New("order not found")
)
