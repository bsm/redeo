package redeo

import (
	"errors"
)

// Common errors
var (
	ErrInvalidRequest    = errors.New("redeo: invalid request")
	ErrWrongNumberOfArgs = errors.New("redeo: wrong number of arguments")
	ErrUnknownCommand    = errors.New("redeo: unknown command")
)
