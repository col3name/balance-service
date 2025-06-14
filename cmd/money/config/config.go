package config

import (
	"github.com/col3name/balance-transfer/pkg/common/app/logger"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/util/env"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres"
)

func ParseConfig(logger logger.Logger) (*postgres.Config, error) {
	env.LoadDotEnvFileIfNeeded(logger)

	var err error
	serveRestAddress, err := env.ParseEnvString("PORT", err, "8000")
	dbAddress, err := env.ParseEnvString("DATABASE_ADDRESS", err, "localhost:5432")
	dbName, err := env.ParseEnvString("DATABASE_NAME", err, "payment")
	dbUser, err := env.ParseEnvString("DATABASE_USER", err, "payment")
	dbPassword, err := env.ParseEnvString("DATABASE_PASSWORD", err, "1234")
	maxConnections, err := env.ParseEnvInt("DATABASE_MAX_CONNECTION", err, 64)
	acquireTimeout, err := env.ParseEnvInt("DATABASE_CONNECTION_TIMEOUT", err, 500)
	currencyApiKey, err := env.ParseEnvString("CURRENCY_API_KEY", err, "89103730-9489-11ec-bd80-b16cd9bfc243")
	migrationsPath, err := env.ParseEnvString("MIGRATION_PATH", err, "./data/postgres/migrations/money")

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
