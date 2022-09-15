package repo

import (
	"context"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/jackc/pgx/v4"
	"strings"
)

type MoneyRepo struct {
	tx pgx.Tx
}

func NewMoneyRepo(tx pgx.Tx) *MoneyRepo {
	r := new(MoneyRepo)
	r.tx = tx
	return r
}

func (r *MoneyRepo) TransferMoney(dto *domain.MoneyTransferRequest) error {
	const sql = `SELECT (balance >= $1) AS valid
FROM account
WHERE id = $2;
UPDATE account
SET balance = balance - $1
WHERE id = $2;
UPDATE account
    SET balance = balance + $1
WHERE id = $3;
INSERT INTO financial_transaction (id, description, amount, from_id, to_id)
VALUES ($4, $5, $1, $2, $3);
`
	var data []interface{}
	data = append(data, dto.Amount, dto.From, dto.To, dto.IdempotencyKey, dto.Description)
	return r.exec(sql, data)
}

func (r *MoneyRepo) CreditOrDebitMoney(dto *domain.MoneyRequest) error {
	var sql string
	var data []interface{}

	if dto.Amount < 0 {
		sql = `SELECT (balance >= $1) AS valid
FROM account
WHERE id = $2;
UPDATE account
SET balance = balance + $1
WHERE id = $2;
INSERT INTO financial_transaction (id, description, amount, from_id)
VALUES ($3, $4, $1, $2);
`
		data = append(data, dto.Amount, dto.Account, dto.IdempotencyKey, dto.Description)
	}
	return r.exec(sql, data)
}

func (r *MoneyRepo) isNotEnoughMoneyErr(err error) bool {
	return strings.Contains(err.Error(), "new row for relation \"account\" violates check constraint \"account_balance_check\"")
}

func (r *MoneyRepo) isAccountNotExist(err error) bool {
	return strings.HasSuffix(err.Error(), " (SQLSTATE 23503)")
}

func (r *MoneyRepo) exec(sql string, data []interface{}) error {
	result, err := r.tx.Exec(context.Background(), sql, data...)
	if err != nil {
		if r.isNotEnoughMoneyErr(err) {
			return domain.ErrNotEnoughMoney
		}
		if r.isAccountNotExist(err) {
			return domain.ErrAccountNotExist
		}
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAccountNotExist
	}
	return err
}
