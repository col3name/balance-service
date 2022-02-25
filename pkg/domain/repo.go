package domain

import "github.com/gofrs/uuid"

type MoneyRepo interface {
	GetBalance(accountId uuid.UUID) (int64, error)
	GetTransactionListRequest(dto *GetTransactionListRequest) (*GetTransactionListReturn, error)
	TransferMoney(dto *MoneyTransferRequest) error
	CreditOrDebitMoney(dto *MoneyRequest) error
}
