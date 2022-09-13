package postgres

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"strings"
)

const ErrNotChanged = "no change"

type Migration struct {
	databaseUrl  string
	migrationDir string
}

func NewMigration(databaseUrl string, migrationDir string) *Migration {
	return &Migration{
		databaseUrl:  databaseUrl,
		migrationDir: migrationDir,
	}
}

func (migration *Migration) SetDatabaseURL(databaseUrl string) {
	migration.databaseUrl = databaseUrl
}

func (migration *Migration) Migrate() error {
	m, err := migrate.New(migration.migrationDir, migration.databaseUrl)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !strings.Contains(err.Error(), ErrNotChanged) {
		return err
	}
	return nil
}
