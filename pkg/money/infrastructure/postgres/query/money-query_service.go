package query

import (
	"context"
	"github.com/col3name/balance-transfer/pkg/money/domain"
	"github.com/col3name/balance-transfer/pkg/money/infrastructure/postgres"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
)

type MoneyQueryService struct {
	connPool *pgxpool.Pool
}

func NewMoneyQueryService(pool *pgxpool.Pool) *MoneyQueryService {
	r := new(MoneyQueryService)
	r.connPool = pool
	return r
}

func (q *MoneyQueryService) GetBalance(accountId uuid.UUID) (int64, error) {
	const sql = "SELECT balance FROM account WHERE id = $1"
	rows, err := q.connPool.Query(context.Background(), sql, accountId)
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

func (q *MoneyQueryService) GetTransactionListRequest(dto *domain.GetTransactionListRequest) (*domain.GetTransactionListReturn, error) {
	countItem, err := q.getCountAllTransactions(dto)
	if err != nil {
		return nil, err
	}
	transactions, sortData, err := q.getTransactionOnPage(dto)
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

func (q *MoneyQueryService) getCountAllTransactions(dto *domain.GetTransactionListRequest) (int, error) {
	sql := q.getCountTransactionSQL()

	args := make([]interface{}, 0, 1)
	args = append(args, dto.AccountId)

	countItem, err := postgres.Query(q.connPool, sql, args, q.getTransactionCountFn)
	if err != nil {
		return 0, err
	}
	return countItem.(int), nil
}

func (q *MoneyQueryService) getCountTransactionSQL() string {
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

func (q *MoneyQueryService) getTransactionCountFn(rows *pgx.Rows) (interface{}, error) {
	var countItem int
	if !(*rows).Next() {
		err := (*rows).Scan(&countItem)
		if err != nil {
			return 0, err
		}
	}
	return countItem, nil
}

func (q *MoneyQueryService) getTransactionOnPage(dto *domain.GetTransactionListRequest) ([]*domain.Transaction, *domain.SortDataDTO, error) {
	sql, args, sortData, err := q.getSqlQueryAndData(dto)
	if err != nil {
		return nil, nil, err
	}
	rows, err := q.connPool.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, nil, err

	}
	if rows.Err() != nil {
		return nil, nil, err

	}
	defer rows.Close()

	transactions, err := q.parseGetTransactionListRequest(rows)
	if err != nil {
		return nil, nil, err

	}
	return transactions, sortData, nil
}

func (q *MoneyQueryService) getSqlQueryAndData(dto *domain.GetTransactionListRequest) (string, []interface{}, *domain.SortDataDTO, error) {
	switch dto.SortField {
	case domain.SortByDate:
		return q.getSortFieldSqlAndData(dto)
	case domain.SortByAmount:
		sql := q.getSortDateSql(dto.SortDirection)
		args, sortData, err := q.getSortDateArgs(dto)
		return sql, args, sortData, err
	}
	return "", nil, nil, domain.ErrUnsupportedSortField
}

func (q *MoneyQueryService) getSortFieldSqlAndData(dto *domain.GetTransactionListRequest) (string, []interface{}, *domain.SortDataDTO, error) {
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
	pageCondition := ` AND datetimestamp` + q.getCompareChar(dto.SortDirection, isNext) + ` $` + strconv.Itoa(i)
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

	sql += q.getOrderBy(dto.SortField, sortDir) + `
         LIMIT $` + strconv.Itoa(i) + `
     ) AS tmp
         JOIN financial_transaction ft ON tmp.id = ft.id
` + q.getOrderBy(dto.SortField, dto.SortDirection) + `
LIMIT $` + strconv.Itoa(i) + `;`

	args = append(args, dto.Limit)

	return sql, args, &domain.SortDataDTO{
		Page:            currentPage,
		IsNextDirection: isNext,
	}, nil
}

const (
	NextCharacter     = " > "
	PreviousCharacter = " < "
)

func (q *MoneyQueryService) getCompareChar(direction domain.SortDirection, next bool) string {
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

func (q *MoneyQueryService) getSortDateSql(sortDirection domain.SortDirection) string {
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

func (q *MoneyQueryService) getSortDateArgs(dto *domain.GetTransactionListRequest) ([]interface{}, *domain.SortDataDTO, error) {
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

func (q *MoneyQueryService) getOrderBy(field domain.SortField, direction domain.SortDirection) string {
	res := ` ORDER BY ` + field.ToString()
	if direction == domain.SortDesc {
		res += " DESC "
	} else {
		res += " ASC "

	}
	return res
}

func (q *MoneyQueryService) parseGetTransactionListRequest(rows pgx.Rows) ([]*domain.Transaction, error) {
	data := make([]*domain.Transaction, 0, 10)
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
