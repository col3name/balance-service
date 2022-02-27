package transport

import (
	"encoding/json"
	"github.com/col3name/balance-transfer/pkg/domain"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type MoneyController struct {
	BaseController
	moneyService domain.MoneyService
}

func NewMoneyController(service domain.MoneyService) *MoneyController {
	c := new(MoneyController)
	c.moneyService = service
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
	idempotencyKey := req.Header.Get("Idempotency-Key")
	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	var r moneyRequest
	err = json.Unmarshal(rawData, &r)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}

	request := domain.NewMoneyRequest(idempotencyKey, r.Account, r.Amount, r.Description)
	if request == nil {
		return nil, domain.ErrInvalidRequest
	}
	return request, nil
}

func decodeMoneyTransferRequest(req *http.Request) (*domain.MoneyTransferRequest, error) {
	idempotencyKey := req.Header.Get("Idempotency-Key")
	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	var r moneyTransferRequest
	err = json.Unmarshal(rawData, &r)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}

	request, err := domain.NewMoneyTransferRequest(idempotencyKey, r.From, r.To, r.Amount, r.Description)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func decodeGetTransactionListRequest(req *http.Request) (*domain.GetTransactionListRequest, error) {
	vars := mux.Vars(req)
	idStr := vars["accountId"]
	query := req.URL.Query()

	cursor := query.Get("cursor")
	sort := query.Get("sort")
	sortField := domain.SortByDate
	if len(sort) > 0 {
		sortFieldVal, err := strconv.Atoi(sort)
		if err != nil {
			return nil, domain.ErrUnsupportedSortField
		}
		switch sortFieldVal {
		case int(domain.SortByAmount):
			sortField = domain.SortByAmount
		case int(domain.SortByDate):
			sortField = domain.SortByDate
		default:
			if sortFieldVal > 1 {
				return nil, domain.ErrUnsupportedSortField
			}
		}
	}

	sortDirection := domain.SortDesc
	order := query.Get("order")
	if len(order) > 0 {
		sortDirVal, err := strconv.Atoi(order)
		if err != nil {
			return nil, domain.ErrUnsupportedSortDirection
		}
		switch sortDirVal {
		case int(domain.SortAsc):
			sortDirection = domain.SortAsc
		case int(domain.SortDesc):
			sortDirection = domain.SortDesc
		default:
			if sortDirVal > 1 {
				return nil, domain.ErrUnsupportedSortDirection
			}
		}
	}

	limitStr := query.Get("limit")
	limit := 2
	if len(limitStr) > 0 {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, domain.ErrInvalidRequest
		}
		limit = l
	}

	getDto := domain.NewGetTransactionListRequest(idStr, cursor, sortField, sortDirection, limit)
	if getDto == nil {
		return nil, domain.ErrInvalidAccountId
	}
	return getDto, nil
}

func decodeGetBalanceRequest(req *http.Request) (*domain.GetBalanceDTO, error) {
	vars := mux.Vars(req)
	idStr := vars["accountId"]
	query := req.URL.Query()

	currencyParam := query.Get("currency")
	var currency domain.Currency
	switch strings.ToUpper(currencyParam) {
	case string(domain.USD):
		currency = domain.USD
	case string(domain.EUR):
		currency = domain.EUR
	case string(domain.RUB):
		currency = domain.RUB
	default:
		if len(currencyParam) > 0 && string(domain.RUB) != currencyParam {
			return nil, domain.ErrNotSupportedCurrency
		}
		currency = domain.RUB
	}

	getBalanceDTO := domain.NewGetBalanceRequest(idStr, currency)
	if getBalanceDTO == nil {
		return nil, domain.ErrInvalidAccountId
	}
	return getBalanceDTO, nil
}
