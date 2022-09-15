package currency

import (
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"sync"
	"time"
)

type Service struct {
	updatedAt   time.Time
	mx          sync.RWMutex
	currencySDK domain.CurrencySDK
	currencyMap map[domain.Currency]float64
}

func NewService(sdk domain.CurrencySDK, from domain.Currency) *Service {
	s := new(Service)
	s.currencySDK = sdk
	err := s.initCurrencyMap(from)
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
	return s.initCurrencyMap(from)
}

func (s *Service) initCurrencyMap(from domain.Currency) error {
	s.currencyMap = make(map[domain.Currency]float64, 2)

	currencyItems, err := s.currencySDK.GetCurrenciesList(from)
	if err != nil {
		return err
	}

	s.fillCurrencyMap(currencyItems)
	s.updatedAt = time.Now()

	return nil
}

func (s *Service) fillCurrencyMap(currencyItems []*domain.CurrencyItem) {
	for _, currencyItem := range currencyItems {
		s.currencyMap[currencyItem.Currency] = currencyItem.Value
	}
}

func (s *Service) isNeedUpdate() bool {
	const CurrencyUpdatePeriod = 10
	return time.Since(s.updatedAt).Minutes() <= CurrencyUpdatePeriod
}
