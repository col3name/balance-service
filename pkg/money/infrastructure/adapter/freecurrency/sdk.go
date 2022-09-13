package freecurrency

import (
	"crypto/tls"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/mailru/easyjson"
	"io"
	"net/http"
)

type Adapter struct {
	apiKey  string
	baseUrl string
	client  *http.Client
}

func NewAdapter(apiKey string, maxIdleConnsPerHost int) *Adapter {
	s := new(Adapter)
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

func (s *Adapter) GetCurrenciesList(baseCurrency domain.Currency) ([]*domain.CurrencyItem, error) {
	resp, err := s.doRequest(baseCurrency)
	if err != nil {
		return nil, domain.ErrApiLimit
	}
	defer resp.Body.Close()

	currencyResp, err := s.unmarshallResponse(resp)
	if err != nil {
		return nil, err
	}
	return s.buildResult(currencyResp)
}

func (s *Adapter) doRequest(baseCurrency domain.Currency) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, s.getApiUrl(baseCurrency), nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(req)
}

func (s *Adapter) getApiUrl(baseCurrency domain.Currency) string {
	return s.baseUrl + "apikey=" + s.apiKey + "&base_currency=" + string(baseCurrency)
}

func (s *Adapter) unmarshallResponse(resp *http.Response) (*CurrencyResponse, error) {
	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	currencyResp := &CurrencyResponse{}
	err = easyjson.Unmarshal(rawBytes, currencyResp)
	if err != nil {
		return nil, err
	}
	return currencyResp, nil
}

func (s *Adapter) buildResult(currencyResp *CurrencyResponse) ([]*domain.CurrencyItem, error) {
	var result []*domain.CurrencyItem
	result = append(result,
		domain.NewCurrencyItem(domain.USD, currencyResp.Data.USD),
		domain.NewCurrencyItem(domain.EUR, currencyResp.Data.EUR),
	)

	return result, nil
}
