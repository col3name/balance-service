package domain

import "github.com/gofrs/uuid"

type MoneyService interface {
	GetBalance(dto *GetBalanceDTO) (*CurrencyReturn, error)
	GetTransactionListRequest(dto *GetTransactionListRequest) (*GetTransactionListReturn, error)
	TransferMoney(dto *MoneyTransferRequest) (*uuid.UUID, error)
	CreditOrDebitMoney(dto *MoneyRequest) (*uuid.UUID, error)
}

type CurrencyService interface {
	Translate(amount int64, from Currency, to Currency) (*CurrencyReturn, error)
}

type CurrencySDK interface {
	GetCurrenciesList(baseCurrency Currency) ([]*CurrencyItem, error)
}
