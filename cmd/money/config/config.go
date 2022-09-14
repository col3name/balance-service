package config

import (
	"flag"
	"fmt"
	"github.com/col3name/balance-transfer/pkg/money/app/log"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

func LoadDotEnvFileIfNeeded(loggerImpl log.Logger) {
	ok := flag.Bool("load", false, "is need load .env file")
	flag.Parse()
	if *ok {
		err := godotenv.Load()
		if err != nil {
			loggerImpl.Fatal("Error loading .env file")
		}
	}
}

func ParseEnvString(key string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	str, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("undefined environment variable %v", key)
	}
	return str, nil
}

func ParseEnvInt(key string, err error) (int, error) {
	s, err := ParseEnvString(key, err)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

func ParseConfig() (*postgres.Config, error) {
	var err error
	serveRestAddress, err := ParseEnvString("PORT", err)
	dbAddress, err := ParseEnvString("DATABASE_ADDRESS", err)
	dbName, err := ParseEnvString("DATABASE_NAME", err)
	dbUser, err := ParseEnvString("DATABASE_USER", err)
	dbPassword, err := ParseEnvString("DATABASE_PASSWORD", err)
	maxConnections, err := ParseEnvInt("DATABASE_MAX_CONNECTION", err)
	acquireTimeout, err := ParseEnvInt("DATABASE_CONNECTION_TIMEOUT", err)
	currencyApiKey, err := ParseEnvString("CURRENCY_API_KEY", err)
	migrationsPath, err := ParseEnvString("MIGRATION_PATH", err)

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
