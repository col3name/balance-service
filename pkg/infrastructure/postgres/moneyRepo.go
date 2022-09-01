package postgres

import (
	"context"
	"github.com/col3name/balance-transfer/pkg/domain"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"math"
	"strconv"
	"strings"
	"time"
)

type MoneyRepo struct {
	pool *pgxpool.Pool
}

func NewMoneyRepo(pool *pgxpool.Pool) *MoneyRepo {
	r := new(MoneyRepo)
	r.pool = pool
	return r
}

func (r *MoneyRepo) GetBalance(accountId uuid.UUID) (int64, error) {
	rows, err := r.pool.Query(context.Background(), "SELECT balance FROM account WHERE id = $1", accountId)
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

func (r *MoneyRepo) setupFieldCursor(dto *domain.GetTransactionListRequest) (time.Time, int, bool, error) {
	var cursorSortByDate time.Time
	var fieldVal string
	var currentPage int
	var err error
	isNext := true
	if len(dto.Cursor) > 0 {
		fieldVal, currentPage, isNext, err = domain.GetVal(dto.Cursor)
		if err != nil {
			return time.Time{}, 0, false, domain.ErrInvalidCursor
		}
		cursorSortByDate, err = time.Parse("2006-01-02 15:04:05.000000 +0000 UTC", fieldVal)
		if err != nil {
			cursorSortByDate, err = time.Parse("2006-01-02 15:04:05.00000 +0000 UTC", fieldVal)
			if err != nil {
				return time.Time{}, 0, false, domain.ErrInvalidCursor
			}
		}
	} else {
		if dto.SortDirection == domain.SortDesc {
			cursorSortByDate = time.Now().Add(25 * time.Hour)
		} else {
			cursorSortByDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return cursorSortByDate, currentPage, isNext, nil
}

func (r *MoneyRepo) GetTransactionListRequest(dto *domain.GetTransactionListRequest) (*domain.GetTransactionListReturn, error) {
	var result domain.GetTransactionListReturn
	countItem, err := r.getCountAllTransaction(dto)
	if err != nil {
		return nil, err
	}
	result.CountItem = countItem

	sql, data, currentPage, isNext, err := r.getSqlQueryAndData(dto)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(context.Background(), sql, data...)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()
	var item domain.Transaction
	var isDebit bool
	result.Data = make([]domain.Transaction, 0)

	for rows.Next() {
		err = rows.Scan(&item.Id, &item.Description, &item.Amount, &item.OtherAccountID, &item.UpdatedAt, &isDebit)
		if err != nil {
			return nil, err
		}
		if item.OtherAccountID == "00000000-0000-0000-0000-000000000000" {
			item.OtherAccountID = ""
		} else if isDebit {
			item.Amount = -item.Amount
		}
		result.Data = append(result.Data, item)
	}

	result.Page = r.setupPage(dto, result, currentPage, isNext)
	return &result, nil
}

func (r *MoneyRepo) setupPage(dto *domain.GetTransactionListRequest, result domain.GetTransactionListReturn, currentPage int, isNext bool) domain.Page {
	var page domain.Page
	countItemInPage := len(result.Data)
	if countItemInPage == 0 {
		return page
	}
	if countItemInPage > 0 {
		var firstItem string
		var lastItem string
		if countItemInPage >= 1 && dto.SortField == domain.SortByDate {
			firstItem, lastItem = r.getFirstLastItems(result.Data, countItemInPage)
		}
		page = r.initPage(dto, currentPage, result.CountItem, isNext, firstItem, lastItem)
	}
	page.Current = r.getCurrentPage(dto, isNext, currentPage)
	return page
}

func (r *MoneyRepo) getCurrentPage(dto *domain.GetTransactionListRequest, isNext bool, currentPage int) int {
	if !isNext && dto.SortField == domain.SortByDate {
		currentPage--
	}
	if currentPage < 0 {
		currentPage = 0
	}
	return currentPage
}

func (r *MoneyRepo) getFirstLastItems(data []domain.Transaction, countItemInPage int) (string, string) {
	return data[0].UpdatedAt.String(), data[countItemInPage-1].UpdatedAt.String()
}
func (r *MoneyRepo) initPage(dto *domain.GetTransactionListRequest, currentPage, countItem int,
	isNext bool, firstItem, lastItem string) domain.Page {
	var page domain.Page
	if (dto.SortField == domain.SortByDate && (len(dto.Cursor) == 0 || ((currentPage == 0 || currentPage == 1) && !isNext))) ||
		dto.SortField == domain.SortByAmount && len(dto.Cursor) == 0 || currentPage < 1 {
		page = domain.Page{Prev: ""}
	} else if (!isNext && currentPage > 0) || (dto.SortField == domain.SortByAmount && currentPage > 0) {
		page.SetPrev(firstItem, currentPage-1)
	} else {
		page.SetPrev(firstItem, currentPage)
	}
	maxPage := int(math.Round(float64(countItem) / float64(dto.Limit)))
	j := currentPage
	if isNext {
		j++
	}
	if j < maxPage {
		if !isNext && dto.SortField != domain.SortByAmount {
			page.SetNext(lastItem, currentPage)
		} else {
			page.SetNext(lastItem, currentPage+1)
		}
	}

	return page
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
	err := WithTransactionSQL(r.pool, sql, data)
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

	err := WithTransactionSQL(r.pool, sql, data)
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

func (r *MoneyRepo) getSqlQueryAndData(dto *domain.GetTransactionListRequest) (string, []interface{}, int, bool, error) {
	if dto.SortField == domain.SortByDate {
		return r.getSortFieldSqlAndData(dto)
	} else if dto.SortField == domain.SortByAmount {
		return r.getSortDateSqlAndData(dto)
	}
	return "", nil, 0, false, domain.ErrUnsupportedSortField
}

func (r *MoneyRepo) getSortFieldSqlAndData(dto *domain.GetTransactionListRequest) (string, []interface{}, int, bool, error) {
	i := 1
	var data []interface{}

	cursorSortByDate, currentPage, isNext, err := r.setupFieldCursor(dto)

	if err != nil {
		return "", nil, 0, false, err
	}
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
	data = append(data, dto.AccountId.String())
	i++
	pageCondition := ` AND datetimestamp` + r.getCompareChar(dto.SortDirection, isNext) + ` $` + strconv.Itoa(i)
	if len(dto.Cursor) > 0 {
		sql += pageCondition
		data = append(data, cursorSortByDate)
	}
	sql += `UNION ALL
         SELECT id, datetimestamp, amount AS value, COALESCE(from_id, uuid_nil()) AS otherAccount, false AS isDebit `
	if dto.SortField == domain.SortByAmount {
		sql += `, abs(amount) AS amounts `
	}
	sql += `FROM financial_transaction
         WHERE to_id = $1`
	if len(dto.Cursor) > 0 {
		sql += pageCondition
		i++
	}
	sortDir := dto.SortDirection
	if !isNext {
		if sortDir == domain.SortDesc {
			sortDir = domain.SortAsc
		} else {
			sortDir = domain.SortDesc
		}
	}

	sql += r.getOrderBy(dto.SortField, sortDir) + `
         LIMIT $` + strconv.Itoa(i) + `
     ) AS tmp
         JOIN financial_transaction ft ON tmp.id = ft.id
` + r.getOrderBy(dto.SortField, dto.SortDirection) + `
LIMIT $` + strconv.Itoa(i) + `;`
	data = append(data, dto.Limit)

	return sql, data, currentPage, isNext, nil
}

func (r *MoneyRepo) getSortDateSqlAndData(dto *domain.GetTransactionListRequest) (string, []interface{}, int, bool, error) {
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
	if dto.SortDirection == domain.SortDesc {
		sql += " DESC "
	}
	sql += `
         OFFSET $2 LIMIT $3
     ) AS tmp
         JOIN financial_transaction ft ON tmp.id = ft.id`
	if dto.SortDirection == domain.SortDesc {
		sql += " ORDER BY ABS(tmp.amount) DESC ;"
	} else {
		sql += ";"
	}
	var data []interface{}
	var currentPage int
	var err error
	isNext := true

	if len(dto.Cursor) > 0 {
		_, currentPage, isNext, err = domain.GetVal(dto.Cursor)
		if err != nil {
			return "", nil, 0, false, domain.ErrInvalidCursor
		}
		if currentPage < 0 {
			currentPage = 0
		}
	}
	data = append(data, dto.AccountId, currentPage*dto.Limit, dto.Limit)
	return sql, data, currentPage, isNext, nil
}

func (r *MoneyRepo) getCountAllTransaction(dto *domain.GetTransactionListRequest) (int, error) {
	sql := r.getCountTransactionSQL()
	data := make([]interface{}, 0)
	data = append(data, dto.AccountId)
	countItem, err := Query(r.pool, sql, data, func(rows *pgx.Rows) (interface{}, error) {
		var countItem int
		if !(*rows).Next() {
			err := (*rows).Scan(&countItem)
			if err != nil {
				return 0, err
			}
		}
		return countItem, nil
	})
	if err != nil {
		return 0, err
	}
	return countItem.(int), nil
}
