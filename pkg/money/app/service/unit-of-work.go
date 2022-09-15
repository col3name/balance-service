package service

import "github.com/col3name/balance-transfer/pkg/money/domain"

type RepositoryProvider interface {
	MoneyRepo() domain.MoneyRepo
}

type Job func(RepositoryProvider) error

type UnitOfWork interface {
	Execute(fn Job) error
}
