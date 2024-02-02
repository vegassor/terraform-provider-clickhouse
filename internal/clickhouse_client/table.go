package clickhouse_client

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

type ClickHouseColumn struct {
	Name    string
	Type    string
	Comment string
}

type ClickHouseTable struct {
	Name    string
	Columns []ClickHouseColumn
	Engine  string
	Comment string
}

func (col ClickHouseColumn) String() string {
	result := fmt.Sprintf(
		"%s %s",
		QuoteWithTicks(col.Name),
		QuoteID(col.Type),
	)

	if col.Comment != "" {
		result += " COMMENT " + QuoteValue(col.Comment)
	}

	return result
}

func (client *ClickHouseClient) CreateTable(ctx context.Context, database string, table ClickHouseTable) error {
	columnsStr := make([]string, 0, len(table.Columns))
	for _, col := range table.Columns {
		columnsStr = append(columnsStr, col.String())
	}
	columns := strings.Join(columnsStr, ",\n")

	query := fmt.Sprintf(
		`CREATE TABLE %s.%s
(
%s
) ENGINE = %s()`,
		QuoteID(database),
		QuoteID(table.Name),
		columns,
		QuoteID(table.Engine),
	)
	if table.Comment != "" {
		query += " COMMENT " + QuoteValue(table.Comment)
	}

	tflog.Info(ctx, "Creating a table", map[string]interface{}{"query": query})

	return client.Conn.Exec(ctx, query)
}
