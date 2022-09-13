package domain

type CurrencyItem struct {
	Currency Currency
	Value    float64
}

func NewCurrencyItem(currency Currency, value float64) *CurrencyItem {
	c := new(CurrencyItem)
	c.Value = value
	c.Currency = currency
	return c
}
