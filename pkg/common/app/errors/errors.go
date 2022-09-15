package errors

import (
	"errors"
)

var (
	ErrInternal = errors.New("internalServerError")
	ErrExternal = errors.New("externalServerError")
)
