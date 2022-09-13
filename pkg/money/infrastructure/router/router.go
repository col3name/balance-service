package router

import (
	"errors"
	"fmt"
	"github.com/col3name/balance-transfer/pkg/money/app/currency"
	"github.com/col3name/balance-transfer/pkg/money/app/money"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/adapter/freecurrency"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/transport"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var ErrFailedInitRouter = errors.New("failed init router")

func Router(pool *pgxpool.Pool, freeCurrencyApiKey string, maxIdleConnection int) (http.Handler, error) {
	router := mux.NewRouter()

	router.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)
	router.HandleFunc("/ready", readyCheckHandler).Methods(http.MethodGet)

	apiV1Route := router.PathPrefix("/api/v1").Subrouter()

	moneyRepo := postgres.NewMoneyRepo(pool)
	sdk := freecurrency.NewAdapter(freeCurrencyApiKey, maxIdleConnection)
	currencyService := currency.NewService(sdk, domain.RUB)
	if currencyService == nil {
		return nil, ErrFailedInitRouter
	}
	service := money.NewService(moneyRepo, currencyService)
	moneyController := transport.NewMoneyController(*service)

	apiV1Route.HandleFunc("/money/{accountId}", moneyController.GetBalance()).Methods(http.MethodGet, http.MethodOptions)
	apiV1Route.HandleFunc("/money/{accountId}/transactions", moneyController.GetTransactionList()).Methods(http.MethodGet, http.MethodOptions)
	apiV1Route.HandleFunc("/money/transfer", moneyController.TransferMoney()).Methods(http.MethodPost)
	apiV1Route.HandleFunc("/money", moneyController.CreditOrDebitMoney()).Methods(http.MethodPost)
	return logMiddleware(router), nil
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "{\"status\": \"OK\"}")
}

func readyCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
