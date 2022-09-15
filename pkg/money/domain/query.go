package domain

import "github.com/gofrs/uuid"

type MoneyQueryService interface {
	GetBalance(accountId uuid.UUID) (int64, error)
	GetTransactionListRequest(dto *GetTransactionListRequest) (*GetTransactionListReturn, error)
}
