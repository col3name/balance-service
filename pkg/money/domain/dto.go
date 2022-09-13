package domain

import (
	b64 "encoding/base64"
	"errors"
	"github.com/gofrs/uuid"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidAccountId = errors.New("invalid account id")

type Currency string

const (
	RUB Currency = "RUB"
	EUR Currency = "EUR"
	USD Currency = "USD"
)

func CurrencyFromString(value string) (Currency, error) {
	var currency Currency
	switch strings.ToUpper(value) {
	case string(USD):
		currency = USD
	case string(EUR):
		currency = EUR
	case string(RUB):
		currency = RUB
	default:
		if len(value) > 0 && string(RUB) != value {
			return "", ErrNotSupportedCurrency
		}
		currency = RUB
	}
	return currency, nil
}

type GetBalanceDTO struct {
	accountId uuid.UUID
	currency  Currency
}

func NewGetBalanceRequest(accountId string, currency Currency) *GetBalanceDTO {
	id, err := getUUID(accountId)
	if err != nil {
		return nil
	}

	return &GetBalanceDTO{
		accountId: id,
		currency:  currency,
	}
}

func (s *GetBalanceDTO) GetAccountId() uuid.UUID {
	return s.accountId
}

func (s *GetBalanceDTO) GetCurrency() Currency {
	return s.currency
}

type CurrencyReturn struct {
	Amount           float64
	ConversationRate float64
}

func NewCurrencyReturn(amount float64, conversationRate float64) *CurrencyReturn {
	if amount < 0 || conversationRate <= 0 {
		return nil
	}
	c := new(CurrencyReturn)
	c.Amount = amount
	c.ConversationRate = conversationRate
	return c
}

type Page struct {
	Prev    string
	Next    string
	Current int
}

func (s *Page) SetPrev(val string, page int) {
	s.Prev = s.generateCursor(val, page, false)
}

func (s *Page) GetPrev() (string, int, bool, error) {
	if s.Prev == "" {
		return "", 0, false, nil
	}
	return GetValue(s.Prev)
}

func (s *Page) GetNext() (string, int, bool, error) {
	if s.Prev == "" {
		return "", 0, false, nil
	}
	return GetValue(s.Prev)
}

func GetValue(cursor string) (string, int, bool, error) {
	res, err := decode(cursor)
	if err != nil {
		return "", 0, false, err
	}
	const ValueSeparator = "!"
	split := strings.Split(string(res), ValueSeparator)
	if len(split) != 3 {
		return "", 0, false, ErrInvalid
	}
	var value string
	var page int
	var isNext bool
	value = split[0]
	number, err := strconv.Atoi(split[1])
	if err != nil {
		return "", 0, false, ErrInvalid
	}
	page = number
	isNext = split[2] == "true"

	return value, page, isNext, nil
}

func (s *Page) generateCursor(val string, page int, isNext bool) string {
	return encode(val + "!" + strconv.Itoa(page) + "!" + strconv.FormatBool(isNext))
}

func (s *Page) SetNext(val string, page int) {
	s.Next = s.generateCursor(val, page, true)
}

func encode(val string) string {
	return b64.StdEncoding.EncodeToString([]byte(val))
}

func decode(val string) ([]byte, error) {
	return b64.StdEncoding.DecodeString(val)
}

type SortField int

const (
	SortByDate SortField = iota
	SortByAmount
)

type SortDirection int

const (
	SortAsc SortDirection = iota
	SortDesc
)

type GetTransactionListRequest struct {
	SortField     SortField
	SortDirection SortDirection
	Cursor        string
	AccountId     uuid.UUID
	Limit         int
}

func NewGetTransactionListRequest(accountId, cursor string, field SortField, direction SortDirection, limit int) *GetTransactionListRequest {
	id, err := getUUID(accountId)
	if err != nil || limit < 1 || limit > 30 {
		return nil
	}

	return &GetTransactionListRequest{
		SortField:     field,
		SortDirection: direction,
		Cursor:        cursor,
		AccountId:     id,
		Limit:         limit,
	}
}

func (s *GetTransactionListRequest) SetCursor(cursor string) {
	s.Cursor = cursor
}

type Transaction struct {
	Id             string
	Amount         int64
	UpdatedAt      time.Time
	Description    string
	OtherAccountID string
}

type GetTransactionListReturn struct {
	Data      []Transaction
	Page      Page
	CountItem int
}

type MoneyTransferRequest struct {
	IdempotencyKey string
	From           string
	To             string
	Amount         int64
	Description    string
}

func NewMoneyTransferRequest(idempotencyKey string, from string, to string, amount int64, description string) (*MoneyTransferRequest, error) {
	s := new(MoneyTransferRequest)
	err := isValidData(amount, from, to)
	if err != nil {
		return nil, err
	}

	idempotencyKey, err = parseIdempotencyKey(idempotencyKey)
	if err != nil {
		return nil, err
	}

	s.IdempotencyKey = idempotencyKey
	s.Amount = amount
	s.From = from
	s.To = to
	s.Description = description
	return s, nil
}

type MoneyRequest struct {
	IdempotencyKey string
	Account        string
	Amount         int64
	Description    string
}

func NewMoneyRequest(idempotencyKey string, account string, amount int64, description string) *MoneyRequest {
	s := new(MoneyRequest)
	if amount == 0 || !isValidUUID(account) {
		return nil
	}

	idempotencyKey, err := parseIdempotencyKey(idempotencyKey)
	if err != nil {
		return nil
	}
	s.IdempotencyKey = idempotencyKey
	s.Amount = amount
	s.Account = account
	s.Description = description
	return s
}

func getUUID(accountId string) (uuid.UUID, error) {
	return uuid.FromString(accountId)
}

func parseIdempotencyKey(idempotencyKey string) (string, error) {
	if len(idempotencyKey) > 0 && !isValidUUID(idempotencyKey) {
		return "", ErrInvalidIdempotencyKey
	} else if len(idempotencyKey) == 0 {
		newKey, err := uuid.NewGen().NewV4()
		if err != nil {
			return "", ErrInvalidRequest
		}
		idempotencyKey = newKey.String()
	}
	return idempotencyKey, nil
}

func isValidData(amount int64, from string, to string) error {
	if amount <= 0 || !isValidUUID(from) || !isValidUUID(to) {
		return ErrInvalidRequest
	}
	if from == to {
		return ErrTransferMoneyToThemself
	}
	return nil
}

func isValidUUID(val string) bool {
	_, err := getUUID(val)
	return err == nil
}
