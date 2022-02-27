package domain

import "errors"

var ErrNotFound = errors.New("not exist")
var ErrApiLimit = errors.New("api limit reached")
var ErrFailedConvert = errors.New("failed convert")
var ErrNotSupportedCurrency = errors.New("not supported currency")
var ErrInvalidRequest = errors.New("invalid request")
var ErrInvalidIdempotencyKey = errors.New("invalid idempotency key, must be uuid v4")
var ErrUnsupportedSortField = errors.New("unsupported sort field")
var ErrUnsupportedSortDirection = errors.New("unsupported sort direction")
var ErrInvalid = errors.New("invalid")
var ErrInvalidCursor = errors.New("invalid cursor")
var ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")
var ErrTransferMoneyToThemself = errors.New("can't transfer money to themself")
var ErrNotEnoughMoney = errors.New("not enough money on account")
