package domain

type MoneyRepo interface {
	TransferMoney(dto *MoneyTransferRequest) error
	CreditOrDebitMoney(dto *MoneyRequest) error
}
