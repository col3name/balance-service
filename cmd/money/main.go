package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/col3name/balance-transfer/pkg/infrastructure/logger"
	"github.com/col3name/balance-transfer/pkg/infrastructure/router"
	"github.com/col3name/balance-transfer/pkg/infrastructure/server"
	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	//TODO write test on postman
	ok := flag.Bool("load", false, "is need load .env file")
	flag.Parse()
	loggerImpl := logger.New()
	if *ok {
		err := godotenv.Load()
		if err != nil {
			loggerImpl.Fatal("Error loading .env file")
		}
	}
	conf, err := ParseConfig()
	if err != nil {
		loggerImpl.Fatal(err)
	}

	connector, err := getConnector(conf)

	if err != nil {
		loggerImpl.Fatal(err.Error())
	}
	pool, err := newConnectionPool(connector)

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

func initHandlers(connPool *pgx.ConnPool, currencyApiKey string, countConnection int) (http.Handler, error) {
	return router.Router(connPool, currencyApiKey, countConnection)
}

type Config struct {
	Port           string
	DbAddress      string
	DbName         string
	DbUser         string
	DbPassword     string
	MaxConnections int
	AcquireTimeout int
	CurrencyApiKey string
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

func ParseConfig() (*Config, error) {
	var err error
	serveRestAddress, err := parseEnvString("PORT", err)
	dbAddress, err := parseEnvString("DATABASE_ADDRESS", err)
	dbName, err := parseEnvString("DATABASE_NAME", err)
	dbUser, err := parseEnvString("DATABASE_USER", err)
	dbPassword, err := parseEnvString("DATABASE_PASSWORD", err)
	maxConnections, err := parseEnvInt("DATABASE_MAX_CONNECTION", err)
	acquireTimeout, err := parseEnvInt("DATABASE_CONNECTION_TIMEOUT", err)
	currencyApiKey, err := parseEnvString("CURRENCY_API_KEY", err)

	if err != nil {
		return nil, err
	}

	return &Config{
		serveRestAddress,
		dbAddress,
		dbName,
		dbUser,
		dbPassword,
		maxConnections,
		acquireTimeout,
		currencyApiKey,
	}, nil
}

func getConnector(config *Config) (pgx.ConnPoolConfig, error) {
	databaseUri := "postgres://" + config.DbUser + ":" + config.DbPassword + "@" + config.DbAddress + "/" + config.DbName
	pgxConnConfig, err := pgx.ParseURI(databaseUri)
	if err != nil {
		return pgx.ConnPoolConfig{}, errors.Wrap(err, "failed to parse database URI from environment variable")
	}
	pgxConnConfig.Dial = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 5 * time.Minute}).Dial
	pgxConnConfig.RuntimeParams = map[string]string{
		"standard_conforming_strings": "on",
	}
	pgxConnConfig.PreferSimpleProtocol = true

	return pgx.ConnPoolConfig{
		ConnConfig:     pgxConnConfig,
		MaxConnections: config.MaxConnections,
		AcquireTimeout: time.Duration(config.AcquireTimeout) * time.Second,
	}, nil
}

func newConnectionPool(config pgx.ConnPoolConfig) (*pgx.ConnPool, error) {
	return pgx.NewConnPool(config)
}
