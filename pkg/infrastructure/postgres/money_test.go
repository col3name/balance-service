package postgres

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"money-transfer/pkg/domain"
	"net"
	"testing"
	"time"
)

type Config struct {
	DbUser         string
	DbPassword     string
	DbAddress      string
	DbName         string
	MaxConnections int
	AcquireTimeout int
}

func getConnector(config *Config) (pgx.ConnPoolConfig, error) {
	databaseUri := "postgres://" + config.DbUser + ":" + config.DbPassword + "@" + config.DbAddress + "/" + config.DbName
	pgxConnConfig, err := pgx.ParseURI(databaseUri)
	if err != nil {
		return pgx.ConnPoolConfig{}, errors.Wrap(err, "failed to parse database URI from environment variable")
	}
	pgxConnConfig.Dial = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 5 * time.Minute}).Dial
	pgxConnConfig.RuntimeParams = map[string]string{
		"standard_conforming_strings": "on",
	}
	pgxConnConfig.PreferSimpleProtocol = true

	return pgx.ConnPoolConfig{
		ConnConfig:     pgxConnConfig,
		MaxConnections: config.MaxConnections,
		AcquireTimeout: time.Duration(config.AcquireTimeout) * time.Second,
	}, nil
}

func newConnectionPool(config pgx.ConnPoolConfig) (*pgx.ConnPool, error) {
	return pgx.NewConnPool(config)
}

func TestName(t *testing.T) {
	parse, err := time.Parse("2006-01-02 15:04:05.000000", "2022-02-23 16:57:50.209150")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(parse.String())
	}
}

func TestGetTransactionList(t *testing.T) {
	request := domain.NewGetTransactionListRequest("67f9ff8c-79ea-4f39-a86e-39fb1d9dfb92", "", domain.SortByDate, domain.SortDesc, 2)

	connector, err := getConnector(&Config{
		DbUser:         "payment",
		DbPassword:     "1234",
		DbName:         "payment",
		DbAddress:      "localhost:5432",
		MaxConnections: 1,
		AcquireTimeout: 2000,
	})
	if err != nil {
		t.Error(err)
		return
	}
	pool, err := newConnectionPool(connector)
	if err != nil {
		t.Error(err)
		return
	}
	repo := NewMoneyRepo(pool)
	var queue []domain.GetTransactionListRequest
	queue = append(queue, *request)
	var req domain.GetTransactionListRequest
	var res *domain.GetTransactionListReturn
	i := 0
	t.Run("can next iterate", func(t *testing.T) {
		var responseList []string
		for len(queue) > 0 {
			i++
			req = queue[0]
			queue = queue[1:]
			res, err = repo.GetTransactionListRequest(&req)
			if err != nil {
				t.Error(err)
				return
			}
			for _, item := range res.Data {
				responseList = append(responseList, item.Description)

				//fmt.Println(item.Description)
			}
			next := res.Page.Next
			if next != "" {
				req.SetCursor(next)
				fmt.Println("prev", res.Page.Prev)
				queue = append(queue, req)
			}
			if i > 3 {
				t.Fatal()
			}
		}

		assert.Equal(t, []string{"buy macbook pro 13 m1, ram 16gb, ssd, 512gb", "cash to travel", "buy iphone 13", "buy iphone 12", "buy pixel 6", "buy macbook pro m1 max"}, responseList)
	})
	t.Run("can reverse iterate", func(t *testing.T) {
		fmt.Println("req", req)
		//req.SetCursor("MjAyMi0wMi0yMyAxNjo1OTo1Mi4xMjE1NzMgKzAwMDAgVVRDITIhdHJ1ZQ==")
		queue = append(queue, req)
		var responseList []string
		i = 0
		for len(queue) > 0 {
			i++
			req = queue[0]
			queue = queue[1:]
			res, err = repo.GetTransactionListRequest(&req)
			if err != nil {
				t.Error(err)
				return
			}
			for _, item := range res.Data {
				responseList = append(responseList, item.Description)
				fmt.Println(item.Description)
			}
			if res.Page.Prev != "" {
				req.SetCursor(res.Page.Prev)
				queue = append(queue, req)
			}
			if i > 3 {
				//t.Fatal()
			}
		}
		expected := []string{"buy pixel 6", "buy macbook pro m1 max", "buy iphone 13", "buy iphone 12", "buy macbook pro 13 m1, ram 16gb, ssd, 512gb", "cash to travel"}
		//sort.Slice(expected, func(i, j int) bool {
		//    return i > j
		//})

		assert.Equal(t, expected, responseList)
	})
}
