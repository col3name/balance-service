package domain

import "errors"

var ErrNotFound = errors.New("not exist")
var ErrApiLimit = errors.New("api limit reached")
var ErrFailedConvert = errors.New("failed convert")
var ErrNotSupportedCurrency = errors.New("not supported currency")
var ErrInvalidRequest = errors.New("invalid request")
var ErrUnsupportedSortField = errors.New("unsupported sort field")
var ErrUnsupportedSortDirection = errors.New("unsupported sort direction")
var ErrInvalid = errors.New("invalid")
var ErrInvalidCursor = errors.New("invalid cursor")
var ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")
