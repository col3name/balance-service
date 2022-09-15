package postgres

import (
	"context"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
	"strings"
)

type MoneyRepo struct {
	connPool *pgxpool.Pool
}

func NewMoneyRepo(pool *pgxpool.Pool) *MoneyRepo {
	r := new(MoneyRepo)
	r.connPool = pool
	return r
}

func (r *MoneyRepo) GetBalance(accountId uuid.UUID) (int64, error) {
	const sql = "SELECT balance FROM account WHERE id = $1"
	rows, err := r.connPool.Query(context.Background(), sql, accountId)
	if err != nil {
		return 0, domain.ErrNotFound
	}
	if rows.Err() != nil {
		return 0, domain.ErrNotFound
	}
	defer rows.Close()
	var balance int64
	if rows.Next() {
		err = rows.Scan(&balance)
		if err != nil {
			return 0, err
		}
	}
	return balance, nil
}

func (r *MoneyRepo) getTransactionOnPage(dto *domain.GetTransactionListRequest) ([]*domain.Transaction, *domain.SortDataDTO, error) {
	sql, args, sortData, err := r.getSqlQueryAndData(dto)
	if err != nil {
		return nil, nil, err
	}
	rows, err := r.connPool.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, nil, err

	}
	if rows.Err() != nil {
		return nil, nil, err

	}
	defer rows.Close()

	transactions, err := r.parseGetTransactionListRequest(rows)
	if err != nil {
		return nil, nil, err

	}
	return transactions, sortData, nil
}

func (r *MoneyRepo) GetTransactionListRequest(dto *domain.GetTransactionListRequest) (*domain.GetTransactionListReturn, error) {
	countItem, err := r.getCountAllTransactions(dto)
	if err != nil {
		return nil, err
	}
	transactions, sortData, err := r.getTransactionOnPage(dto)
	if err != nil {
		return nil, err
	}

	result := &domain.GetTransactionListReturn{
		CountItem:    countItem,
		Transactions: transactions,
		Page:         domain.NewPage(dto, transactions, countItem, sortData),
	}

	return result, nil
}

func (r *MoneyRepo) parseGetTransactionListRequest(rows pgx.Rows) ([]*domain.Transaction, error) {
	data := make([]*domain.Transaction, 10)

	var transaction domain.Transaction
	var isDebit bool
	var err error
	for rows.Next() {
		err = rows.Scan(&transaction.Id, &transaction.Description, &transaction.Amount, &transaction.OtherAccountID, &transaction.UpdatedAt, &isDebit)
		if err != nil {
			return nil, err
		}

		transaction.TransferMoney(isDebit)

		data = append(data, &transaction)
	}
	return data, err
}

func (r *MoneyRepo) TransferMoney(dto *domain.MoneyTransferRequest) error {
	sql := `SELECT (balance >= $1) AS valid
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
	err := WithTransactionSQL(r.connPool, sql, data)
	if err != nil {
		if r.isNotEnoughMoneyErr(err) {
			return domain.ErrNotEnoughMoney
		}
		if r.isAccountNotExist(err) {
			return domain.ErrAccountNotExist
		}
	}
	return err
}

func (r *MoneyRepo) isNotEnoughMoneyErr(err error) bool {
	return strings.Contains(err.Error(), "new row for relation \"account\" violates check constraint \"account_balance_check\"")
}

func (r *MoneyRepo) isAccountNotExist(err error) bool {
	return strings.HasSuffix(err.Error(), " (SQLSTATE 23503)")
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

	err := WithTransactionSQL(r.connPool, sql, data)
	if err != nil {
		if r.isNotEnoughMoneyErr(err) {
			return domain.ErrNotEnoughMoney
		}
		if r.isAccountNotExist(err) {
			return domain.ErrAccountNotExist
		}
	}
	return err
}

func (r *MoneyRepo) getSortField(field domain.SortField) string {
	if field == domain.SortByDate {
		return " datetimestamp "
	} else if field == domain.SortByAmount {
		return " amounts "
	}
	return ""
}

const NextCharacter = " > "
const PreviousCharacter = " < "

func (r *MoneyRepo) getCompareChar(direction domain.SortDirection, next bool) string {
	if direction == domain.SortAsc {
		if next {
			return NextCharacter
		}
		return PreviousCharacter
	}
	if next {
		return PreviousCharacter
	}
	return NextCharacter
}

func (r *MoneyRepo) getOrderBy(field domain.SortField, direction domain.SortDirection) string {
	res := ` ORDER BY ` + r.getSortField(field)
	if direction == domain.SortDesc {
		res += " DESC "
	} else {
		res += " ASC "

	}
	return res
}

func (r *MoneyRepo) getSqlQueryAndData(dto *domain.GetTransactionListRequest) (string, []interface{}, *domain.SortDataDTO, error) {
	switch dto.SortField {
	case domain.SortByDate:
		return r.getSortFieldSqlAndData(dto)
	case domain.SortByAmount:
		sql := r.getSortDateSql(dto.SortDirection)
		args, sortData, err := r.getSortDateArgs(dto)
		return sql, args, sortData, err
	}
	return "", nil, nil, domain.ErrUnsupportedSortField
}

func (r *MoneyRepo) getSortFieldSqlAndData(dto *domain.GetTransactionListRequest) (string, []interface{}, *domain.SortDataDTO, error) {
	sortData, err := dto.Cursor.ToSortData(dto.SortDirection)
	if err != nil {
		return "", nil, nil, err
	}

	cursorSortByDate := sortData.Date
	currentPage := sortData.Page
	isNext := sortData.IsNextDirection
	i := 1
	var args []interface{}

	sql := `SELECT ft.id, ft.description,  tmp.value, tmp.otherAccount, tmp.datetimestamp, tmp.isDebit `
	if dto.SortField == domain.SortByAmount {
		sql += `, abs(amount) AS amounts`
	}
	sql += `
FROM (
         SELECT id, datetimestamp, amount AS value, COALESCE(to_id, uuid_nil()) AS otherAccount, true AS isDebit `
	if dto.SortField == domain.SortByAmount {
		sql += `, abs(amount) AS amounts `
	}

	sql += ` FROM financial_transaction
         WHERE from_id = $` + strconv.Itoa(i)

	args = append(args, dto.AccountId.String())
	i++
	pageCondition := ` AND datetimestamp` + r.getCompareChar(dto.SortDirection, isNext) + ` $` + strconv.Itoa(i)
	if !dto.Cursor.Empty() {
		sql += pageCondition
		args = append(args, cursorSortByDate)
	}
	sql += `UNION ALL
         SELECT id, datetimestamp, amount AS value, COALESCE(from_id, uuid_nil()) AS otherAccount, false AS isDebit `
	if dto.SortField == domain.SortByAmount {
		sql += `, abs(amount) AS amounts `
	}
	sql += `FROM financial_transaction
         WHERE to_id = $1`
	if !dto.Cursor.Empty() {
		sql += pageCondition
		i++
	}

	sortDir := dto.SortDirection
	if !isNext {
		sortDir.Toggle()
	}

	sql += r.getOrderBy(dto.SortField, sortDir) + `
         LIMIT $` + strconv.Itoa(i) + `
     ) AS tmp
         JOIN financial_transaction ft ON tmp.id = ft.id
` + r.getOrderBy(dto.SortField, dto.SortDirection) + `
LIMIT $` + strconv.Itoa(i) + `;`

	args = append(args, dto.Limit)

	return sql, args, &domain.SortDataDTO{
		Page:            currentPage,
		IsNextDirection: isNext,
	}, nil
}

func (r *MoneyRepo) getSortDateSql(sortDirection domain.SortDirection) string {
	sql := `SELECT ft.id, ft.description, tmp.amount, tmp.otherAccount, ft.datetimestamp, tmp.isDebit
FROM (
         SELECT id, amount, otherAccount, true AS isDebit
         FROM (
                  SELECT id, amount, COALESCE(to_id, uuid_nil()) AS otherAccount, true AS isDebit
                  FROM financial_transaction
                  WHERE from_id = $1
                  UNION ALL
                  SELECT id, amount, COALESCE(from_id, uuid_nil()) AS otherAccount, false AS isDebit
                  FROM financial_transaction
                  WHERE to_id = $1
              ) AS t
         ORDER BY ABS(amount)`
	if sortDirection == domain.SortDesc {
		sql += " DESC "
	}
	sql += `
         OFFSET $2 LIMIT $3
     ) AS tmp
         JOIN financial_transaction ft ON tmp.id = ft.id`
	if sortDirection == domain.SortDesc {
		sql += " ORDER BY ABS(tmp.amount) DESC ;"
	} else {
		sql += ";"
	}
	return sql
}

func (r *MoneyRepo) getSortDateArgs(dto *domain.GetTransactionListRequest) ([]interface{}, *domain.SortDataDTO, error) {
	var err error
	result := &domain.SortDataDTO{}
	result.Page, result.IsNextDirection, err = dto.Cursor.ToPage()
	if err != nil {
		return nil, nil, err
	}

	args := make([]interface{}, 0, 3)

	args = append(args, dto.AccountId, result.Page*dto.Limit, dto.Limit)

	return args, result, nil
}

func (r *MoneyRepo) getCountAllTransactions(dto *domain.GetTransactionListRequest) (int, error) {
	sql := r.getCountTransactionSQL()

	args := make([]interface{}, 0)
	args = append(args, dto.AccountId)

	countItem, err := Query(r.connPool, sql, args, r.getTransactionCountFn)
	if err != nil {
		return 0, err
	}
	return countItem.(int), nil
}

func (r *MoneyRepo) getCountTransactionSQL() string {
	return `SELECT SUM(tmp.count) FROM (
                           SELECT COUNT(id)
                           FROM financial_transaction
                           WHERE from_id = $1
                           UNION ALL
                           SELECT COUNT(id)
                           FROM financial_transaction
                           WHERE to_id = $1
                           ) tmp;`
}

func (r *MoneyRepo) getTransactionCountFn(rows *pgx.Rows) (interface{}, error) {
	var countItem int
	if !(*rows).Next() {
		err := (*rows).Scan(&countItem)
		if err != nil {
			return 0, err
		}
	}
	return countItem, nil
}
