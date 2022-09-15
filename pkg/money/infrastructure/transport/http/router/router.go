package router

import (
	"errors"
	"fmt"
	"github.com/col3name/balance-transfer/pkg/common/app/logger"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/http/types"
	"github.com/col3name/balance-transfer/pkg/money/app/currency"
	"github.com/col3name/balance-transfer/pkg/money/app/money"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/adapter/freecurrency"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres/query"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres/repo"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/transport/http/handler"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const Version = "v1"

var ErrFailedInitRouter = errors.New("failed init router")

func Router(pool *pgxpool.Pool, logger logger.Logger, freeCurrencyApiKey string, maxIdleConnection int) (http.Handler, error) {
	moneyController, err := setupMoneyController(pool, logger, freeCurrencyApiKey, maxIdleConnection)
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()

	setupDefaultMethod(router)

	apiV1Route := getApiVersionSubRoute(router, Version)
	setupMoneyRoutes(apiV1Route, moneyController)

	return logMiddleware(router), nil
}

func setupMoneyRoutes(apiV1Route *mux.Router, moneyController *handler.MoneyController) {
	apiV1Route.HandleFunc("/money/{accountId}", moneyController.GetBalance()).Methods(http.MethodGet, http.MethodOptions)
	apiV1Route.HandleFunc("/money/{accountId}/transactions", moneyController.GetTransactionList()).Methods(http.MethodGet, http.MethodOptions)
	apiV1Route.HandleFunc("/money/transfer", moneyController.TransferMoney()).Methods(http.MethodPost)
	apiV1Route.HandleFunc("/money", moneyController.CreditOrDebitMoney()).Methods(http.MethodPost)
}

func getApiVersionSubRoute(router *mux.Router, version string) *mux.Router {
	return router.PathPrefix("/api/" + version).Subrouter()
}

func setupDefaultMethod(router *mux.Router) {
	router.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)
	router.HandleFunc("/ready", readyCheckHandler).Methods(http.MethodGet)
}

func setupMoneyController(pool *pgxpool.Pool, logger logger.Logger, freeCurrencyApiKey string, maxIdleConnection int) (*handler.MoneyController, error) {
	unitOfWork := repo.NewUnitOfWork(pool, logger)
	moneyQuery := query.NewMoneyQueryService(pool)
	sdk := freecurrency.NewAdapter(freeCurrencyApiKey, maxIdleConnection)
	currencyService := currency.NewService(sdk, domain.RUB)
	if currencyService == nil {
		return nil, ErrFailedInitRouter
	}
	service := money.NewService(unitOfWork, moneyQuery, currencyService)
	moneyController := handler.NewMoneyController(*service)
	return moneyController, nil
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(types.ContentType, types.ApplicationJson)
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "{\"status\": \"OK\"}")
}

func readyCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(types.ContentType, types.ApplicationJson)
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "{\"host\": \"%v\"}", r.Host)
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		log.WithFields(log.Fields{
			"method":     r.Method,
			"url":        r.URL,
			"remoteAddr": r.RemoteAddr,
			"userAgent":  r.UserAgent(),
		}).Info("got a new request")
		h.ServeHTTP(w, r)
	})
}
