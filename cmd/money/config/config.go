package config

import (
	"github.com/col3name/balance-transfer/pkg/common/app/logger"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/util/env"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres"
)

func ParseConfig(logger logger.Logger) (*postgres.Config, error) {
	env.LoadDotEnvFileIfNeeded(logger)

	var err error
	serveRestAddress, err := env.ParseEnvString("PORT", err)
	dbAddress, err := env.ParseEnvString("DATABASE_ADDRESS", err)
	dbName, err := env.ParseEnvString("DATABASE_NAME", err)
	dbUser, err := env.ParseEnvString("DATABASE_USER", err)
	dbPassword, err := env.ParseEnvString("DATABASE_PASSWORD", err)
	maxConnections, err := env.ParseEnvInt("DATABASE_MAX_CONNECTION", err)
	acquireTimeout, err := env.ParseEnvInt("DATABASE_CONNECTION_TIMEOUT", err)
	currencyApiKey, err := env.ParseEnvString("CURRENCY_API_KEY", err)
	migrationsPath, err := env.ParseEnvString("MIGRATION_PATH", err)

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
