package chclient

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type HttpConn struct {
	DB *sql.DB
}

type HttpRows struct {
	*sql.Rows
}

func (c *HttpConn) Ping(ctx context.Context) error {
	return c.DB.Ping()
}

func (c *HttpConn) Exec(ctx context.Context, query string, args ...any) error {
	_, err := c.DB.ExecContext(ctx, query, args)
	return err
}

func (c *HttpConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	rows, err := c.DB.QueryContext(ctx, query, args)
	if err != nil {
		return nil, err
	}

	return &HttpRows{rows}, nil
}

func (c *HttpConn) Contributors() []string {
	return []string{}
}

func (c *HttpConn) ServerVersion() (*driver.ServerVersion, error) {
	return nil, errors.New("not implemented")
}

func (c *HttpConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return errors.New("not implemented")
}

func (c *HttpConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	return nil
}

func (c *HttpConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return nil, errors.New("not implemented")
}

func (c *HttpConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return errors.New("not implemented")
}

func (c *HttpConn) Stats() driver.Stats {
	return driver.Stats{}
}

func (c *HttpConn) Close() error {
	return c.DB.Close()
}

func (r *HttpRows) ScanStruct(dest any) error {
	return errors.New("not implemented")
}

func (r *HttpRows) Totals(dest ...any) error {
	return errors.New("not implemented")
}

func (r *HttpRows) ColumnTypes() []driver.ColumnType {
	return nil
}

func (r *HttpRows) Columns() []string {
	cols, err := r.Rows.Columns()
	if err != nil {
		panic(err)
	}

	return cols
}
