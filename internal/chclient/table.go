package chclient

import (
	"context"
	"fmt"
	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

type ClickHouseColumn struct {
	Name    string
	Type    string
	Comment string
}

type ClickHouseColumns []ClickHouseColumn

func (cols ClickHouseColumns) Names() []string {
	names := make([]string, len(cols))
	for i, col := range cols {
		names[i] = col.Name
	}
	return names
}

type ClickHouseTable struct {
	Database string
	Name     string
	Columns  ClickHouseColumns
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
where database = %s and table = %s
order by "position"`,
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

	desiredColsSet := hashset.New[string](desiredTable.Columns.Names()...)
	currentColsSet := hashset.New[string](desiredTable.Columns.Names()...)
	oldCols := currentColsSet.Intersection(desiredColsSet)
	newCols := desiredColsSet.Difference(currentColsSet)
	desiredColsMap := make(map[string]struct {
		ClickHouseColumn
		int
	}, len(desiredTable.Columns))
	for i, col := range desiredTable.Columns {
		desiredColsMap[col.Name] = struct {
			ClickHouseColumn
			int
		}{col, i}
	}

	if !oldCols.Contains(currentColsSet.Values()...) {
		return &NotSupportedError{
			Operation: "renaming or removing columns",
			Detail: fmt.Sprintf("cannot update columns of table %s.%s:"+
				"desired config does not contain columns from previous config."+
				"If you did not try to rename or delete columns, it is a bug in the Client",
				desiredTable.Database,
				desiredTable.Name,
			),
		}
	}

	for _, colName := range newCols.Values() {
		query := fmt.Sprintf(
			`ALTER TABLE %s.%s ADD COLUMN %s TYPE %s COMMENT %s`,
			QuoteID(desiredTable.Database),
			QuoteID(currentTableName),
			QuoteID(colName),
			QuoteID(desiredColsMap[colName].Type),
			QuoteValue(desiredColsMap[colName].Comment),
		)
		tflog.Info(ctx, "Adding a column", dict{"query": query})
		err := client.Conn.Exec(ctx, query)
		if err != nil {
			return err
		}
	}

	for _, col := range desiredTable.Columns {
		query := fmt.Sprintf(`ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s COMMENT %s`,
			QuoteID(desiredTable.Database),
			QuoteID(desiredTable.Name),
			QuoteID(col.Name),
			QuoteID(col.Type),
			QuoteValue(col.Comment),
		)
		tflog.Info(ctx, "Changing a column", dict{"query": query})
		err := client.Conn.Exec(ctx, query)
		if err != nil {
			return err
		}
	}

	for i := range desiredTable.Columns {
		desiredCol := desiredTable.Columns[i]
		currentCol := currentTable.Columns[i]
		if desiredCol.Name == currentCol.Name {
			continue
		}

		desiredIdx := i
		var query string

		if desiredIdx == 0 {
			query = fmt.Sprintf(
				`ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s FIRST`,
				QuoteID(desiredTable.Database),
				QuoteID(desiredTable.Name),
				QuoteID(desiredCol.Name),
				QuoteID(desiredCol.Type),
			)
		} else {
			prevColName := desiredTable.Columns[desiredIdx-1].Name
			query = fmt.Sprintf(
				`ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s AFTER %s`,
				QuoteID(desiredTable.Database),
				QuoteID(desiredTable.Name),
				QuoteID(desiredCol.Name),
				QuoteID(desiredCol.Type),
				QuoteID(prevColName),
			)
		}

		tflog.Info(ctx, "Changing column's order", dict{"query": query})
		err := client.Conn.Exec(ctx, query)
		if err != nil {
			return err
		}
	}
	return nil
}
