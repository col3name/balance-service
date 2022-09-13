package db

type Migration interface {
	Migrate() error
	SetDatabaseURL(string)
}
