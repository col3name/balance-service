package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/server"
	"github.com/col3name/balance-transfer/pkg/money/app/log"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/logger"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/transport/router"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	loggerImpl := logger.New()

	loadDotEnvFileIfNeeded(loggerImpl)
	conf, err := parseConfig()
	if err != nil {
		loggerImpl.Fatal(err)
	}
	pool, err := prepareDbPool(conf, err)
	if err != nil {
		loggerImpl.Fatal(err.Error())
	}
	handler, err := initHandlers(pool, conf.CurrencyApiKey, 128)
	if err != nil {
		loggerImpl.Fatal(err.Error())
	}
	runHttpServer(loggerImpl, conf, handler, err)
}

func loadDotEnvFileIfNeeded(loggerImpl log.Logger) {
	ok := flag.Bool("load", false, "is need load .env file")
	flag.Parse()
	if *ok {
		err := godotenv.Load()
		if err != nil {
			loggerImpl.Fatal("Error loading .env file")
		}
	}
}

func prepareDbPool(conf *postgres.Config, err error) (*pgxpool.Pool, error) {
	migration := postgres.NewMigration("", conf.MigrationsPath)
	return postgres.GetReadyConnectionToDB(conf, migration)
}

func initHandlers(connPool *pgxpool.Pool, currencyApiKey string, countConnection int) (http.Handler, error) {
	return router.Router(connPool, currencyApiKey, countConnection)
}

func runHttpServer(loggerImpl log.Logger, conf *postgres.Config, handler http.Handler, err error) {
	loggerImpl.Info("Start at" + time.Now().String())

	httpServer := server.HttpServer{Logger: loggerImpl}
	srv := httpServer.StartServer(conf.Port, handler)
	killSignalChan := httpServer.GetKillSignalChan()
	httpServer.WaitForKillSignal(killSignalChan)
	err = srv.Shutdown(context.Background())

	loggerImpl.Info("Stop at" + time.Now().String())

	if err != nil {
		loggerImpl.Error(err)
		return
	}
}

func parseEnvString(key string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	str, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("undefined environment variable %v", key)
	}
	return str, nil
}

func parseEnvInt(key string, err error) (int, error) {
	s, err := parseEnvString(key, err)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

func parseConfig() (*postgres.Config, error) {
	var err error
	serveRestAddress, err := parseEnvString("PORT", err)
	dbAddress, err := parseEnvString("DATABASE_ADDRESS", err)
	dbName, err := parseEnvString("DATABASE_NAME", err)
	dbUser, err := parseEnvString("DATABASE_USER", err)
	dbPassword, err := parseEnvString("DATABASE_PASSWORD", err)
	maxConnections, err := parseEnvInt("DATABASE_MAX_CONNECTION", err)
	acquireTimeout, err := parseEnvInt("DATABASE_CONNECTION_TIMEOUT", err)
	currencyApiKey, err := parseEnvString("CURRENCY_API_KEY", err)
	migrationsPath, err := parseEnvString("MIGRATION_PATH", err)

	if err != nil {
		return nil, err
	}

	return &postgres.Config{
		Port:           serveRestAddress,
		DbAddress:      dbAddress,
		DbName:         dbName,
		DbUser:         dbUser,
		DbPassword:     dbPassword,
		MaxConnections: maxConnections,
		AcquireTimeout: acquireTimeout,
		CurrencyApiKey: currencyApiKey,
		MigrationsPath: migrationsPath,
	}, nil
}
