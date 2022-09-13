package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/logger"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/transport/router"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/transport/server"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	loggerImpl := logger.New()

	ok := flag.Bool("load", false, "is need load .env file")
	flag.Parse()
	if *ok {
		err := godotenv.Load()
		if err != nil {
			loggerImpl.Fatal("Error loading .env file")
		}
	}
	conf, err := parseConfig()
	if err != nil {
		loggerImpl.Fatal(err)
	}
	migration := postgres.NewMigration("", conf.MigrationsPath)
	pool, err := postgres.GetReadyConnectionToDB(conf, migration)

	if err != nil {
		loggerImpl.Fatal(err.Error())
	}
	handler, err := initHandlers(pool, conf.CurrencyApiKey, 128)
	if err != nil {
		loggerImpl.Fatal(err.Error())
	}
	loggerImpl.Info("Start at" + time.Now().String())
	httpServer := server.HttpServer{Logger: loggerImpl}
	srv := httpServer.StartServer(conf.Port, handler)
	killSignalChan := httpServer.GetKillSignalChan()
	httpServer.WaitForKillSignal(killSignalChan)
	err = srv.Shutdown(context.TODO())
	loggerImpl.Info("Stop at" + time.Now().String())

	if err != nil {
		loggerImpl.Error(err)
		return
	}
}

func initHandlers(connPool *pgxpool.Pool, currencyApiKey string, countConnection int) (http.Handler, error) {
	return router.Router(connPool, currencyApiKey, countConnection)
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
