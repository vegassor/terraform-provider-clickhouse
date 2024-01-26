package clickhouse_client

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseClient struct {
	Conn driver.Conn
}

func NewClickHouseClient(connOpts *clickhouse.Options) (*ClickHouseClient, error) {
	conn, err := clickhouse.Open(connOpts)

	if err != nil {
		return nil, err
	}

	return &ClickHouseClient{conn}, nil
}
