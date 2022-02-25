package currency

import (
	"money-transfer/pkg/domain"
	"sync"
	"time"
)

type Service struct {
	currencySDK domain.CurrencySDK
	updatedAt   time.Time
	data        map[domain.Currency]float64
	mx          sync.RWMutex
}

func NewService(sdk domain.CurrencySDK, from domain.Currency) *Service {
	s := new(Service)
	s.currencySDK = sdk
	s.data = make(map[domain.Currency]float64, 2)
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
	val, ok := s.data[to]
	s.mx.RUnlock()
	if !ok {
		return nil, domain.ErrFailedConvert
	}
	return domain.NewCurrencyReturn(float64(amount)*val, 1/val), nil
}

func (s *Service) updateIfNeeded(from domain.Currency) error {
	if s.isNeedUpdate() {
		err := s.updateCurrencies(from)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) updateCurrencies(from domain.Currency) error {
	list, err := s.currencySDK.GetCurrenciesList(from)
	if err != nil {
		return err
	}

	s.mx.Lock()
	for _, item := range list {
		s.data[item.Currency] = item.Value
	}

	s.updatedAt = time.Now()
	s.mx.Unlock()

	return nil
}

func (s *Service) isNeedUpdate() bool {
	return time.Since(s.updatedAt).Minutes() >= 10
}
