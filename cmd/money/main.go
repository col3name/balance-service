package main

import (
	"context"
	"github.com/col3name/balance-transfer/cmd/money/config"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/server"
	"github.com/col3name/balance-transfer/pkg/money/app/log"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/logger"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/transport/http/router"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"time"
)

func main() {
	loggerImpl := logger.New()

	conf, err := config.ParseConfig(loggerImpl)
	if err != nil {
		loggerImpl.Fatal(err)
	}
	pool, err := prepareDbPool(conf)
	if err != nil {
		loggerImpl.Fatal(err.Error())
	}
	handler, err := router.Router(pool, conf.CurrencyApiKey, 128)
	if err != nil {
		loggerImpl.Fatal(err.Error())
	}
	runHttpServer(loggerImpl, conf, handler)
}

func prepareDbPool(conf *postgres.Config) (*pgxpool.Pool, error) {
	migration := postgres.NewMigration("", conf.MigrationsPath)

	return postgres.GetReadyConnectionToDB(conf, migration)
}

func runHttpServer(loggerImpl log.Logger, conf *postgres.Config, handler http.Handler) {
	loggerImpl.Info("Start at" + time.Now().String())
	defer loggerImpl.Info("Stop at" + time.Now().String())

	httpServer := server.HttpServer{Logger: loggerImpl}
	srv := httpServer.StartServer(conf.Port, handler)
	httpServer.WaitForKillSignal()

	err := srv.Shutdown(context.Background())
	if err != nil {
		loggerImpl.Error(err)
		return
	}
}
