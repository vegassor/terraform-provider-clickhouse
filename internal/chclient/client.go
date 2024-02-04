package chclient

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type dict map[string]interface{}
type ClickHouseClient struct {
	Conn driver.Conn
}

func NewClickHouseClient(connOpts *clickhouse.Options) (*ClickHouseClient, error) {
	conn, err := clickhouse.Open(connOpts)

	if err != nil {
		return nil, err
	}

	err = conn.Ping(context.TODO())
	if err != nil {
		return nil, err
	}

	return &ClickHouseClient{conn}, nil
}

type ClickHouseClientError interface {
	query() string
}

type NotFoundError struct {
	Entity string
	Name   string
	Query  string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("could not find %s %s: query: %s", e.Entity, e.Name, e.Query)
}

func (e *NotFoundError) query() string {
	return e.Query
}

var _ error = &NotFoundError{}
var _ ClickHouseClientError = &NotFoundError{}
