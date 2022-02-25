package transport

import (
	"errors"
	"money-transfer/pkg/domain"
	"net/http"
)

var ErrUnexpected = errors.New("unexpected error")

func TranslateError(err error) *Error {
	if errorIs(err, ErrBadRouting) {
		return NewError(http.StatusNotFound, 100, err)
	} else if errorIs(err, ErrBadRequest) {
		return NewError(http.StatusBadRequest, 101, err)
	} else if errorIs(err, domain.ErrNotFound) {
		return NewError(http.StatusNotFound, 102, err)
	} else if errorIs(err, domain.ErrInvalidAccountId) {
		return NewError(http.StatusBadRequest, 103, err)
	} else if errorIs(err, domain.ErrNotSupportedCurrency) {
		return NewError(http.StatusBadRequest, 104, err)
	} else if errorIs(err, domain.ErrInvalidRequest) {
		return NewError(http.StatusBadRequest, 104, err)
	} else if errorIs(err, domain.ErrDuplicateIdempotencyKey) {
		return NewError(http.StatusOK, 104, err)
	} else if errorIs(err, domain.ErrInvalidCursor) {
		return NewError(http.StatusBadRequest, 104, err)
	} else if errorIs(err, domain.ErrUnsupportedSortField) {
		return NewError(http.StatusBadRequest, 104, err)
	} else if errorIs(err, domain.ErrUnsupportedSortDirection) {
		return NewError(http.StatusBadRequest, 104, err)
	}

	return NewError(http.StatusInternalServerError, 100, ErrUnexpected)
}

func errorIs(err, target error) bool {
	return errors.Is(err, target)
}
