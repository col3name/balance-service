package freecurrency

import (
	"crypto/tls"
	"github.com/col3name/balance-transfer/pkg/domain"
	"github.com/mailru/easyjson"
	"io/ioutil"
	"net/http"
)

type SDK struct {
	apiKey  string
	baseUrl string
	client  *http.Client
}

func NewSDK(apiKey string, maxIdleConnsPerHost int) *SDK {
	s := new(SDK)
	s.apiKey = apiKey
	s.baseUrl = "https://freecurrencyapi.net/api/v2/latest?"
	s.client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			TLSClientConfig: &tls.Config{
				ClientSessionCache: tls.NewLRUClientSessionCache(maxIdleConnsPerHost),
			},
		},
	}

	return s
}

func (s *SDK) GetCurrenciesList(baseCurrency domain.Currency) ([]domain.CurrencyItem, error) {
	req, err := http.NewRequest(http.MethodGet, s.getApiUrl(baseCurrency), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, domain.ErrApiLimit
	}
	defer resp.Body.Close()
	rawBytes, err := ioutil.ReadAll(resp.Body)
	someStruct := &CurrencyResponse{}
	err = easyjson.Unmarshal(rawBytes, someStruct)
	if err != nil {
		return nil, err
	}
	var result []domain.CurrencyItem
	result = append(result,
		*domain.NewCurrencyItem(domain.USD, someStruct.Data.USD),
		*domain.NewCurrencyItem(domain.EUR, someStruct.Data.EUR))

	return result, nil
}

func (s *SDK) getApiUrl(baseCurrency domain.Currency) string {
	return s.baseUrl + "apikey=" + s.apiKey + "&base_currency=" + string(baseCurrency)
}
