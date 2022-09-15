package postgres

import (
	"context"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/db"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Config struct {
	MaxConnections int
	AcquireTimeout int
	Port           string
	DbAddress      string
	DbName         string
	DbUser         string
	DbPassword     string
	CurrencyApiKey string
	MigrationsPath string
}

func GetReadyConnectionToDB(config *Config, migration db.Migration) (*pgxpool.Pool, error) {
	databaseUrl := getDatabaseURL(config)
	connect, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		return nil, err
	}
	migration.SetDatabaseURL(databaseUrl)
	err = migration.Migrate()
	if err != nil {
		return connect, err
	}
	return connect, nil
}

func getDatabaseURL(config *Config) string {
	return "postgres://" + config.DbUser + ":" + config.DbPassword + "@" + config.DbAddress + "/" + config.DbName + "?sslmode=disable"
}
