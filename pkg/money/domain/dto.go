package domain

import (
	str "github.com/col3name/balance-transfer/pkg/common/infrastructure/util/strings"
	uuidUtil "github.com/col3name/balance-transfer/pkg/common/infrastructure/util/uuid"
	"github.com/gofrs/uuid"
	"math"
	"strconv"
	"strings"
	"time"
)

type Currency string

const (
	RUB Currency = "RUB"
	EUR Currency = "EUR"
	USD Currency = "USD"
)

func NewCurrency(value string) (Currency, error) {
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
	currency  Currency
	accountId uuid.UUID
}

func NewGetBalanceRequest(accountId string, currency Currency) *GetBalanceDTO {
	id, err := uuidUtil.FromString(accountId)
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

const DefaultConversionRate = 1

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
	Current  int
	Previous Cursor
	Next     Cursor
}

func NewPage(dto *GetTransactionListRequest, transactions []*Transaction, countItem int, sortData *SortDataDTO) *Page {
	var page Page
	countItemInPage := len(transactions)
	if countItemInPage == 0 {
		return nil
	}
	currentPage := sortData.Page
	isNext := sortData.IsNextDirection

	if countItemInPage > 0 {
		var firstItem string
		var lastItem string
		if countItemInPage >= 1 && dto.SortField == SortByDate {
			firstItem, lastItem = getFirstLastItems(transactions, countItemInPage)
		}
		page = initPage(dto, currentPage, countItem, isNext, firstItem, lastItem)
	}
	page.Current = getCurrentPage(dto, isNext, currentPage)
	return &page
}

func getFirstLastItems(data []*Transaction, countItemInPage int) (string, string) {
	return data[0].UpdatedAt.String(), data[countItemInPage-1].UpdatedAt.String()
}

func initPage(dto *GetTransactionListRequest, currentPage, countItem int,
	isNext bool, firstItem, lastItem string) Page {
	var page Page
	cursor := dto.Cursor
	sortField := dto.SortField
	isEmpty := cursor.Empty()

	if (sortField == SortByDate && (isEmpty || ((currentPage == 0 || currentPage == 1) && !isNext))) ||
		sortField == SortByAmount && isEmpty || currentPage < 1 {
		page = Page{Previous: ""}
	} else if (!isNext && currentPage > 0) || (sortField == SortByAmount && currentPage > 0) {
		page.SetPrevious(firstItem, currentPage-1)
	} else {
		page.SetPrevious(firstItem, currentPage)
	}
	currPage := currentPage
	if isNext {
		currPage++
	}
	if currPage < getCountPage(countItem, dto.Limit) {
		if !isNext && sortField != SortByAmount {
			page.SetNext(lastItem, currentPage)
		} else {
			page.SetNext(lastItem, currentPage+1)
		}
	}

	return page
}

func getCountPage(countItem int, limit int) int {
	return int(math.Round(float64(countItem) / float64(limit)))
}

func getCurrentPage(dto *GetTransactionListRequest, isNext bool, currentPage int) int {
	if !isNext && dto.SortField == SortByDate {
		currentPage--
	}
	if currentPage < 0 {
		currentPage = 0
	}
	return currentPage
}

func (s *Page) SetPrevious(time string, page int) {
	s.Previous = NewCursor(time, page, false)
}

func (s *Page) GetPrevious() (string, int, bool, error) {
	if s.Previous.Empty() {
		return "", 0, false, nil
	}
	return s.Previous.split()
}

func (s *Page) GetNext() (string, int, bool, error) {
	if s.Previous == "" {
		return "", 0, false, nil
	}
	return s.Next.split()
}

func (s *Page) SetNext(time string, page int) {
	s.Next = NewCursor(time, page, true)
}

type SortDataDTO struct {
	IsNextDirection bool
	Page            int
	Date            time.Time
}

type Cursor string

const CursorValueSeparator = "!"

func NewCursor(time string, page int, isNextDirection bool) Cursor {
	result := str.B64encode(time + CursorValueSeparator + strconv.Itoa(page) + CursorValueSeparator + strconv.FormatBool(isNextDirection))
	return Cursor(result)
}

func (c *Cursor) split() (string, int, bool, error) {
	if c.Empty() {
		return "", 0, false, nil
	}
	res, err := str.B64decode(string(*c))
	if err != nil {
		return "", 0, false, err
	}
	split := strings.Split(string(res), CursorValueSeparator)
	if len(split) != 3 {
		return "", 0, false, ErrInvalid
	}

	var times string
	var page int
	var isNext bool
	times = split[0]
	number, err := strconv.Atoi(split[1])
	if err != nil {
		return "", 0, false, ErrInvalid
	}
	page = number
	isNext = split[2] == "true"

	return times, page, isNext, nil
}

func (c *Cursor) ToPage() (int, bool, error) {
	var currentPage int
	isNextDirection := true
	var err error

	if !c.Empty() {
		_, currentPage, isNextDirection, err = c.split()
		if err != nil {
			return 0, false, ErrInvalidCursor
		}
		if currentPage < 0 {
			currentPage = 0
		}
	}

	return currentPage, isNextDirection, nil
}

func (c *Cursor) Empty() bool {
	return len(*c) == 0
}

func (c *Cursor) ToSortData(sortDirection SortDirection) (*SortDataDTO, error) {
	sortData := &SortDataDTO{
		Date:            time.Time{},
		Page:            0,
		IsNextDirection: true,
	}

	var err error
	var cursorTime string
	if !c.Empty() {
		cursorTime, sortData.Page, sortData.IsNextDirection, err = c.split()
		if err != nil {
			return nil, ErrInvalidCursor
		}
		sortData.Date, err = c.getSortDateByDate(cursorTime)
		if err != nil {
			return nil, ErrInvalidCursor
		}
	} else {
		sortData.Date = c.getSortDateBySortDirection(sortDirection)
	}

	return sortData, nil
}

func (c *Cursor) getSortDateBySortDirection(sortDirection SortDirection) time.Time {
	var sortDate time.Time

	if sortDirection == SortDesc {
		sortDate = time.Now().Add(25 * time.Hour)
	} else {
		sortDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	return sortDate
}

func (c *Cursor) getSortDateByDate(times string) (time.Time, error) {
	const baseCursorLayout = "2006-01-02 15:04:05.000000 +0000 UTC"
	const secondCursorLayout = "2006-01-02 15:04:05.00000 +0000 UTC"

	sortDate, err := time.Parse(baseCursorLayout, times)
	if err != nil {
		sortDate, err = time.Parse(secondCursorLayout, times)
	}
	return sortDate, err
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

func (d *SortDirection) Toggle() {
	if *d == SortDesc {
		*d = SortAsc
	} else {
		*d = SortDesc
	}
}

type GetTransactionListRequest struct {
	Limit         int
	SortField     SortField
	SortDirection SortDirection
	Cursor        Cursor
	AccountId     uuid.UUID
}

func NewGetTransactionListRequest(accountId, cursor string, field SortField, direction SortDirection, limit int) *GetTransactionListRequest {
	id, err := uuidUtil.FromString(accountId)
	if err != nil || limit < 1 || limit > 30 {
		return nil
	}

	return &GetTransactionListRequest{
		SortField:     field,
		SortDirection: direction,
		Cursor:        Cursor(cursor),
		AccountId:     id,
		Limit:         limit,
	}
}

func (s *GetTransactionListRequest) SetCursor(cursor string) {
	s.Cursor = Cursor(cursor)
}

const EmptyUUID = "00000000-0000-0000-0000-000000000000"

type Transaction struct {
	Amount         int64
	Id             string
	Description    string
	OtherAccountID string
	UpdatedAt      time.Time
}

func (t *Transaction) TransferMoney(isDebit bool) {
	if t.OtherAccountID == EmptyUUID {
		t.OtherAccountID = ""
	} else if isDebit {
		t.Amount = -t.Amount
	}
}

type GetTransactionListReturn struct {
	CountItem    int
	Page         *Page
	Transactions []*Transaction
}

type MoneyTransferRequest struct {
	Amount         int64
	IdempotencyKey string
	From           string
	To             string
	Description    string
}

func NewMoneyTransferRequest(idempotencyKey string, from string, to string, amount int64, description string) (*MoneyTransferRequest, error) {
	s := new(MoneyTransferRequest)

	err := s.isValidTransferRequest(amount, from, to)
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

func (s *MoneyTransferRequest) isValidTransferRequest(amount int64, from string, to string) error {
	if amount <= 0 || !uuidUtil.IsValid(from) || !uuidUtil.IsValid(to) {
		return ErrInvalidRequest
	}
	if from == to {
		return ErrTransferMoneyToThemself
	}
	return nil
}

type IdempotencyKey string

type MoneyRequest struct {
	Amount         int64
	IdempotencyKey string
	Account        string
	Description    string
}

func GetTransactionId(idempotencyKey string) (uuid.UUID, error) {
	var transactionId uuid.UUID
	var err error
	if len(idempotencyKey) == 0 {
		transactionId, err = uuid.NewV4()
		if err != nil {
			return uuid.UUID{}, err
		}
		idempotencyKey = transactionId.String()
	}
	return uuid.FromString(idempotencyKey)
}

func NewMoneyRequest(idempotencyKey string, account string, amount int64, description string) *MoneyRequest {
	s := new(MoneyRequest)
	if amount == 0 || !uuidUtil.IsValid(account) {
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

func parseIdempotencyKey(idempotencyKey string) (string, error) {
	if len(idempotencyKey) > 0 && !uuidUtil.IsValid(idempotencyKey) {
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
