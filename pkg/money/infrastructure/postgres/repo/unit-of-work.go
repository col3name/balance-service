package repo

import (
	"github.com/col3name/balance-transfer/pkg/common/app/logger"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/db"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/postgres"
	"github.com/col3name/balance-transfer/pkg/money/app/service"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/jackc/pgx/v4"
)

type unitOfWork struct {
	db     *db.Database
	logger logger.Logger
}

func NewUnitOfWork(conn postgres.PgxPoolIface, logger logger.Logger) service.UnitOfWork {
	return &unitOfWork{
		db:     db.NewDatabase(conn),
		logger: logger,
	}
}

func (u *unitOfWork) Execute(fn service.Job) error {
	cancelFunc, err := u.db.WithTx(func(tx pgx.Tx) error {
		return fn(&repositoryProvider{tx: tx})
	}, u.logger)
	if err != nil {
		return infrastructure.InternalError(u.logger, err)
	}
	defer cancelFunc()
	return nil
}

type repositoryProvider struct {
	tx pgx.Tx
}

func (r *repositoryProvider) MoneyRepo() domain.MoneyRepo {
	return NewMoneyRepo(r.tx)
}
