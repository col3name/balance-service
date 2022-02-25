package postgres

import (
	"github.com/jackc/pgx"
	"money-transfer/pkg/domain"
	"strings"
)

func WithTransactionSQL(connPool *pgx.ConnPool, sql string, data []interface{}) error {
	tx, err := connPool.Begin()
	if err != nil {
		if tx != nil {
			return tx.Rollback()
		}
		return err
	}
	_, err = tx.Exec(sql, data...)
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
			if strings.HasSuffix(err.Error(), "duplicate key value violates unique constraint \"transaction_pkey\"") {
				return domain.ErrDuplicateIdempotencyKey
			}
			return err
		}
		return err
	}

	return tx.Commit()
}

func Query(connPool *pgx.ConnPool, sql string, data []interface{}, fn func(rows *pgx.Rows) (interface{}, error)) (interface{}, error) {
	rows, err := connPool.Query(sql, data...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	res, err := fn(rows)
	if err != nil {
		return nil, err
	}

	return res, nil
}
