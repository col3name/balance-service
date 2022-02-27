package money

import (
	"github.com/col3name/balance-transfer/pkg/domain"
	"github.com/gofrs/uuid"
)

type Service struct {
	repo            domain.MoneyRepo
	currencyService domain.CurrencyService
}

func NewService(repo domain.MoneyRepo, service domain.CurrencyService) *Service {
	s := new(Service)
	s.repo = repo
	s.currencyService = service
	return s
}

func (s Service) GetBalance(dto *domain.GetBalanceDTO) (*domain.CurrencyReturn, error) {
	amount, err := s.repo.GetBalance(dto.GetAccountId())
	if err != nil {
		return nil, domain.ErrNotFound
	}
	targetCurrency := dto.GetCurrency()
	result := domain.NewCurrencyReturn(float64(amount), 1)
	if targetCurrency != domain.RUB {
		result, err = s.currencyService.Translate(amount, domain.RUB, targetCurrency)
		if err != nil {
			return nil, domain.ErrFailedConvert
		}
	}
	return result, nil
}

func (s Service) GetTransactionListRequest(dto *domain.GetTransactionListRequest) (*domain.GetTransactionListReturn, error) {
	return s.repo.GetTransactionListRequest(dto)
}

func (s Service) TransferMoney(dto *domain.MoneyTransferRequest) (*uuid.UUID, error) {
	transactionId, err := s.getTransactionId(dto.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	dto.IdempotencyKey = transactionId.String()
	err = s.repo.TransferMoney(dto)
	if err != nil {
		return nil, err
	}
	return &transactionId, err
}

func (s Service) CreditOrDebitMoney(dto *domain.MoneyRequest) (*uuid.UUID, error) {
	transactionId, err := s.getTransactionId(dto.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	dto.IdempotencyKey = transactionId.String()
	err = s.repo.CreditOrDebitMoney(dto)
	if err != nil {
		return nil, err
	}
	return &transactionId, err
}

func (s Service) getTransactionId(idempotencyKey string) (uuid.UUID, error) {
	var transactionId uuid.UUID
	var err error
	if len(idempotencyKey) == 0 {
		transactionId, err = uuid.NewV4()
		if err != nil {
			return uuid.UUID{}, err
		}
		idempotencyKey = transactionId.String()
	} else {
		fromString, err := uuid.FromString(idempotencyKey)
		if err != nil {
			return uuid.UUID{}, err
		}
		transactionId = fromString
	}
	return transactionId, nil
}
