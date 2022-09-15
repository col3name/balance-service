package currency

import (
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"sync"
	"time"
)

type Service struct {
	currencySDK domain.CurrencySDK
	updatedAt   time.Time
	currencyMap map[domain.Currency]float64
	mx          sync.RWMutex
}

func NewService(sdk domain.CurrencySDK, from domain.Currency) *Service {
	s := new(Service)
	s.currencySDK = sdk
	s.currencyMap = make(map[domain.Currency]float64, 2)
	err := s.updateCurrencies(from)
	if err != nil {
		return nil
	}
	return s
}

func (s *Service) Translate(amount int64, from domain.Currency, to domain.Currency) (*domain.CurrencyReturn, error) {
	err := s.updateIfNeeded(from)
	if err != nil {
		return nil, err
	}
	s.mx.RLock()
	number, ok := s.currencyMap[to]
	s.mx.RUnlock()
	if !ok {
		return nil, domain.ErrFailedConvert
	}
	return domain.NewCurrencyReturn(float64(amount)*number, 1/number), nil
}

func (s *Service) updateIfNeeded(from domain.Currency) error {
	if !s.isNeedUpdate() {
		return nil
	}
	return s.updateCurrencies(from)
}

func (s *Service) updateCurrencies(from domain.Currency) error {
	currencyItems, err := s.currencySDK.GetCurrenciesList(from)
	if err != nil {
		return err
	}

	s.fillCurrencyMap(currencyItems)

	return nil
}

func (s *Service) fillCurrencyMap(currencyItems []*domain.CurrencyItem) {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, currencyItem := range currencyItems {
		s.currencyMap[currencyItem.Currency] = currencyItem.Value
	}

	s.updatedAt = time.Now()
}

func (s *Service) isNeedUpdate() bool {
	return time.Since(s.updatedAt).Minutes() <= 10
}
