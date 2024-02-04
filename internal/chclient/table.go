package chclient

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
	Database string
	Name     string
	Columns  []ClickHouseColumn
	Engine   string
	Comment  string
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

func (client *ClickHouseClient) CreateTable(ctx context.Context, table ClickHouseTable) error {
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
		QuoteID(table.Database),
		QuoteID(table.Name),
		columns,
		QuoteID(table.Engine),
	)
	if table.Comment != "" {
		query += " COMMENT " + QuoteValue(table.Comment)
	}

	tflog.Info(ctx, "Creating a table", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetTable(ctx context.Context, database string, table string) (ClickHouseTable, error) {
	query := fmt.Sprintf(
		`SELECT "database", "name", "engine", "comment"
FROM "system"."tables"
WHERE "database" = %s AND "name" = %s`,
		QuoteValue(database),
		QuoteValue(table),
	)

	tflog.Info(ctx, "Looking for a table", dict{"query": query})

	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return ClickHouseTable{}, err
	}

	if !rows.Next() {
		return ClickHouseTable{}, &NotFoundError{
			Entity: "table",
			Name:   fmt.Sprintf("%s.%s", database, table),
			Query:  query,
		}
	}

	var dbReceived string
	var nameReceived string
	var engineReceived string
	var commentReceived string
	err = rows.Scan(
		&dbReceived,
		&nameReceived,
		&engineReceived,
		&commentReceived,
	)
	if err != nil {
		return ClickHouseTable{}, err
	}

	query = fmt.Sprintf(
		`SELECT "name", "type", "comment" from "system"."columns"
where database = %s and table = %s`,
		QuoteValue(database),
		QuoteValue(table),
	)
	rows, err = client.Conn.Query(ctx, query)
	if err != nil {
		return ClickHouseTable{}, err
	}
	var cols []ClickHouseColumn
	for rows.Next() {
		var col ClickHouseColumn
		err := rows.Scan(&col.Name, &col.Type, &col.Comment)
		if err != nil {
			return ClickHouseTable{}, err
		}
		cols = append(cols, col)
	}

	return ClickHouseTable{
		Database: dbReceived,
		Name:     nameReceived,
		Engine:   engineReceived,
		Comment:  commentReceived,
		Columns:  cols,
	}, nil
}

func (client *ClickHouseClient) AlterTable(ctx context.Context, currentTableName string, desiredTable ClickHouseTable) error {
	if currentTableName != desiredTable.Name {
		return fmt.Errorf(
			"table name mismatch: %s != %s. Renaming tables is not yet supported",
			desiredTable.Name,
			currentTableName,
		)
	}

	currentTable, err := client.GetTable(ctx, desiredTable.Database, currentTableName)
	if err != nil {
		return err
	}

	for i := range desiredTable.Columns {
		desiredCol := desiredTable.Columns[i]
		currentCol := currentTable.Columns[i]
		if desiredCol.Name != currentCol.Name {
			return fmt.Errorf(
				"column name mismatch: %s != %s",
				desiredCol.Name,
				currentCol.Name,
			)
		}

		query := fmt.Sprintf(`ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s COMMENT %s`,
			QuoteID(desiredTable.Database),
			QuoteID(currentTableName),
			QuoteID(desiredCol.Name),
			QuoteWithTicks(desiredCol.Type),
			QuoteValue(desiredCol.Comment),
		)
		tflog.Info(ctx, "Changing a column", dict{"query": query})
		err := client.Conn.Exec(ctx, query)
		if err != nil {
			return err
		}
	}
	return nil
}
