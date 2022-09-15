package money

import (
	"github.com/col3name/balance-transfer/pkg/money/app/service"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/gofrs/uuid"
)

type Service struct {
	query           domain.MoneyQueryService
	unitOfWork      service.UnitOfWork
	currencyService domain.CurrencyService
}

func NewService(unitOfWork service.UnitOfWork, query domain.MoneyQueryService, service domain.CurrencyService) *Service {
	s := new(Service)
	s.query = query
	s.unitOfWork = unitOfWork
	s.currencyService = service
	return s
}

func (s Service) GetBalance(dto *domain.GetBalanceDTO) (*domain.CurrencyReturn, error) {
	amount, err := s.query.GetBalance(dto.GetAccountId())
	if err != nil {
		return nil, domain.ErrNotFound
	}
	currency := domain.NewCurrencyReturn(float64(amount), domain.DefaultConversionRate)
	targetCurrency := dto.GetCurrency()
	if targetCurrency == domain.RUB {
		return currency, nil
	}

	currency, err = s.currencyService.Translate(amount, domain.RUB, targetCurrency)
	if err != nil {
		return nil, domain.ErrFailedConvert
	}
	return currency, nil
}

func (s Service) GetTransactionListRequest(dto *domain.GetTransactionListRequest) (*domain.GetTransactionListReturn, error) {
	return s.query.GetTransactionListRequest(dto)
}

func (s Service) TransferMoney(dto *domain.MoneyTransferRequest) (*uuid.UUID, error) {
	transactionId, err := domain.GetTransactionId(dto.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	dto.IdempotencyKey = transactionId.String()

	err = s.unitOfWork.Execute(func(provider service.RepositoryProvider) error {
		repo := provider.MoneyRepo()
		return repo.TransferMoney(dto)
	})
	if err != nil {
		return nil, err
	}
	return &transactionId, err
}

func (s Service) CreditOrDebitMoney(dto *domain.MoneyRequest) (*uuid.UUID, error) {
	transactionId, err := domain.GetTransactionId(dto.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	dto.IdempotencyKey = transactionId.String()
	err = s.unitOfWork.Execute(func(provider service.RepositoryProvider) error {
		repo := provider.MoneyRepo()
		return repo.CreditOrDebitMoney(dto)
	})
	if err != nil {
		return nil, err
	}
	return &transactionId, err
}
