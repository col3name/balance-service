package transport

import (
	"encoding/json"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strconv"
)

type MoneyController struct {
	BaseController
	moneyService domain.MoneyService
}

func NewMoneyController(moneyService domain.MoneyService) *MoneyController {
	c := new(MoneyController)
	c.moneyService = moneyService
	return c
}

func (c *MoneyController) GetBalance() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		reqDTO, err := decodeGetBalanceRequest(req)
		if err != nil {
			c.WriteError(w, err, TranslateError(err))
			return
		}
		balance, err := c.moneyService.GetBalance(reqDTO)
		if err != nil {
			c.WriteError(w, err, TranslateError(err))
			return
		}
		c.WriteJsonResponse(w, &balance)
	}
}

func (c *MoneyController) GetTransactionList() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (*req).Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		reqDTO, err := decodeGetTransactionListRequest(req)
		if err != nil {
			c.WriteError(w, err, TranslateError(err))
			return
		}
		resultDTO, err := c.moneyService.GetTransactionListRequest(reqDTO)
		if err != nil {
			c.WriteError(w, err, TranslateError(err))
			return
		}
		c.WriteJsonResponse(w, &resultDTO)
	}
}

func (c *MoneyController) TransferMoney() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		reqDTO, err := decodeMoneyTransferRequest(req)
		if err != nil {
			c.WriteError(w, err, TranslateError(err))
			return
		}
		transactionId, err := c.moneyService.TransferMoney(reqDTO)
		if err != nil {
			c.WriteError(w, err, TranslateError(err))
			return
		}
		c.WriteJsonResponse(w, response{
			Data: (*transactionId).String(),
		})
	}
}

func (c *MoneyController) CreditOrDebitMoney() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		reqDto, err := decodeMoneyRequest(req)
		if err != nil {
			c.WriteError(w, err, TranslateError(err))
			return
		}
		transactionId, err := c.moneyService.CreditOrDebitMoney(reqDto)
		if err != nil {
			c.WriteError(w, err, TranslateError(err))
			return
		}
		c.WriteJsonResponse(w, response{
			Data: (*transactionId).String(),
		})
	}
}

func decodeMoneyRequest(req *http.Request) (*domain.MoneyRequest, error) {

	var r moneyRequest
	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	idempotencyKey := getIdempotencyKey(req)
	request := domain.NewMoneyRequest(idempotencyKey, r.Account, r.Amount, r.Description)
	if request == nil {
		return nil, domain.ErrInvalidRequest
	}
	return request, nil
}

func decodeMoneyTransferRequest(req *http.Request) (*domain.MoneyTransferRequest, error) {

	var r moneyTransferRequest
	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	idempotencyKey := getIdempotencyKey(req)
	return domain.NewMoneyTransferRequest(idempotencyKey, r.From, r.To, r.Amount, r.Description)
}

func getIdempotencyKey(req *http.Request) string {
	return req.Header.Get("Idempotency-Key")
}

const FieldCurrency = "currency"
const FieldSort = "sort"
const FiledOrder = "order"
const FieldLimit = "limit"
const FieldCursor = "cursor"

func decodeGetTransactionListRequest(req *http.Request) (*domain.GetTransactionListRequest, error) {
	vars := mux.Vars(req)
	accountId := vars["accountId"]
	query := req.URL.Query()

	cursor := query.Get(FieldCursor)
	sortField, err := getSortFieldParameter(query)
	if err != nil {
		return nil, err
	}
	sortDirection, err := getSortDirectionParameter(query)
	if err != nil {
		return nil, err
	}
	var limit int
	limit, err = getLimitParameter(query)
	if err != nil {
		return nil, err
	}
	getDto := domain.NewGetTransactionListRequest(accountId, cursor, sortField, sortDirection, limit)
	if getDto == nil {
		return nil, domain.ErrInvalidAccountId
	}
	return getDto, nil
}

func getLimitParameter(query url.Values) (int, error) {
	limitStr := query.Get(FieldLimit)
	limit := 2
	if len(limitStr) > 0 {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			return 0, domain.ErrInvalidRequest
		}
		limit = l
	}
	return limit, nil
}

func getSortDirectionParameter(query url.Values) (domain.SortDirection, error) {
	order := query.Get(FiledOrder)
	sortDirection := domain.SortDesc
	if len(order) > 0 {
		sortDirVal, err := strconv.Atoi(order)
		if err != nil {
			return 0, domain.ErrUnsupportedSortDirection
		}
		switch sortDirVal {
		case int(domain.SortAsc):
			sortDirection = domain.SortAsc
		case int(domain.SortDesc):
			sortDirection = domain.SortDesc
		default:
			if sortDirVal > 1 {
				return 0, domain.ErrUnsupportedSortDirection
			}
		}
	}
	return sortDirection, nil
}

func getSortFieldParameter(query url.Values) (domain.SortField, error) {
	sort := query.Get(FieldSort)
	sortField := domain.SortByDate
	if len(sort) > 0 {
		sortFieldVal, err := strconv.Atoi(sort)
		if err != nil {
			return 0, domain.ErrUnsupportedSortField
		}
		switch sortFieldVal {
		case int(domain.SortByAmount):
			sortField = domain.SortByAmount
		case int(domain.SortByDate):
			sortField = domain.SortByDate
		default:
			if sortFieldVal > 1 {
				return 0, domain.ErrUnsupportedSortField
			}
		}
	}
	return sortField, nil
}

func decodeGetBalanceRequest(req *http.Request) (*domain.GetBalanceDTO, error) {
	vars := mux.Vars(req)
	idStr := vars["accountId"]
	query := req.URL.Query()

	currency, err := domain.CurrencyFromString(query.Get(FieldCurrency))
	if err != nil {
		return nil, err
	}

	getBalanceDTO := domain.NewGetBalanceRequest(idStr, currency)
	if getBalanceDTO == nil {
		return nil, domain.ErrInvalidAccountId
	}
	return getBalanceDTO, nil
}
