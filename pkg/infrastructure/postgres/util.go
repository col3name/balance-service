package postgres

import (
	"context"
	"github.com/col3name/balance-transfer/pkg/domain"
	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strings"
)

func WithTransactionSQL(connPool *pgxpool.Pool, sql string, data []interface{}) error {
	tx, err := connPool.Begin(context.Background())
	if err != nil {
		if tx != nil {
			return tx.Rollback(context.Background())
		}
		return err
	}
	_, err = tx.Exec(context.Background(), sql, data...)
	if err != nil {
		if tx != nil {
			_ = tx.Rollback(context.Background())
			if strings.HasSuffix(err.Error(), "duplicate key value violates unique constraint \"transaction_pkey\"") {
				return domain.ErrDuplicateIdempotencyKey
			}
			return err
		}
		return err
	}

	return tx.Commit(context.Background())
}

func Query(connPool *pgxpool.Pool, sql string, data []interface{}, fn func(rows *pgx.Rows) (interface{}, error)) (interface{}, error) {
	rows, err := connPool.Query(context.Background(), sql, data...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	res, err := fn(&rows)
	if err != nil {
		return nil, err
	}

	return res, nil
}
