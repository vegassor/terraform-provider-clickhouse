package chclient

import (
	"context"
	"fmt"
	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"regexp"
	"strings"
	"time"
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
	Database     string
	Name         string
	Comment      string
	Engine       string
	EngineParams []string
	OrderBy      []string
	Columns      ClickHouseColumns
}

type ClickHouseTableFullInfo struct {
	Database                    string
	Name                        string
	UUID                        uuid.UUID
	Comment                     string
	Engine                      string
	EngineFull                  string
	IsTemporary                 bool
	DataPaths                   []string
	MetadataPath                string
	MetadataModificationTime    time.Time
	DependenciesDatabase        []string
	DependenciesTable           []string
	CreateTableQuery            string
	AsSelect                    string
	PartitionKey                string
	SortingKey                  string
	PrimaryKey                  string
	SamplingKey                 string
	StoragePolicy               string
	TotalRows                   *uint64
	TotalBytes                  *uint64
	TotalBytesUncompressed      *uint64
	LifetimeRows                *uint64
	LifetimeBytes               *uint64
	HasOwnData                  bool
	LoadingDependenciesDatabase []string
	LoadingDependenciesTable    []string
	LoadingDependentDatabase    []string
	LoadingDependentTable       []string

	Columns      ClickHouseColumns
	EngineParams []string
	OrderBy      []string
}

func (info ClickHouseTableFullInfo) ToTable() ClickHouseTable {
	return ClickHouseTable{
		Database:     info.Database,
		Name:         info.Name,
		Comment:      info.Comment,
		Engine:       info.Engine,
		EngineParams: info.EngineParams,
		OrderBy:      info.OrderBy,
		Columns:      info.Columns,
	}
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

	query := fmt.Sprintf(
		`CREATE TABLE %s.%s
(
%s
) ENGINE = %s(%s)`,
		QuoteID(table.Database),
		QuoteID(table.Name),
		strings.Join(columnsStr, ",\n"),
		QuoteID(table.Engine),
		strings.Join(QuoteList(table.EngineParams, "`"), " "),
	)
	if len(table.OrderBy) > 0 {
		query += " ORDER BY (" + strings.Join(QuoteList(table.OrderBy, "`"), ", ") + ")"
	}

	if table.Comment != "" {
		query += " COMMENT " + QuoteValue(table.Comment)
	}

	tflog.Info(ctx, "Creating a table", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetTable(ctx context.Context, database string, table string) (ClickHouseTableFullInfo, error) {
	query := fmt.Sprintf(
		`SELECT
	"database",
    "name",
    "uuid",
    "comment",
    "engine",
    "engine_full",
    "is_temporary",
    "data_paths",
    "metadata_path",
    "metadata_modification_time",
    "dependencies_database",
    "dependencies_table",
    "create_table_query",
    "as_select",
    "partition_key",
    "sorting_key",
    "primary_key",
    "sampling_key",
    "storage_policy",
    "total_rows",
    "total_bytes",
    "total_bytes_uncompressed",
    "lifetime_rows",
    "lifetime_bytes",
    "has_own_data",
    "loading_dependencies_database",
    "loading_dependencies_table",
    "loading_dependent_database",
    "loading_dependent_table"
FROM "system"."tables"
WHERE "database" = %s AND "name" = %s`,
		QuoteValue(database),
		QuoteValue(table),
	)

	tflog.Info(ctx, "Looking for a table", dict{"query": query})

	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return ClickHouseTableFullInfo{}, err
	}

	if !rows.Next() {
		return ClickHouseTableFullInfo{}, &NotFoundError{
			Entity: "table",
			Name:   fmt.Sprintf("%s.%s", database, table),
			Query:  query,
		}
	}

	var tableInfo ClickHouseTableFullInfo
	var hasOwnData, isTemporary uint8
	err = rows.Scan(
		&tableInfo.Database,
		&tableInfo.Name,
		&tableInfo.UUID,
		&tableInfo.Comment,
		&tableInfo.Engine,
		&tableInfo.EngineFull,
		&isTemporary,
		&tableInfo.DataPaths,
		&tableInfo.MetadataPath,
		&tableInfo.MetadataModificationTime,
		&tableInfo.DependenciesDatabase,
		&tableInfo.DependenciesTable,
		&tableInfo.CreateTableQuery,
		&tableInfo.AsSelect,
		&tableInfo.PartitionKey,
		&tableInfo.SortingKey,
		&tableInfo.PrimaryKey,
		&tableInfo.SamplingKey,
		&tableInfo.StoragePolicy,
		&tableInfo.TotalRows,
		&tableInfo.TotalBytes,
		&tableInfo.TotalBytesUncompressed,
		&tableInfo.LifetimeRows,
		&tableInfo.LifetimeBytes,
		&hasOwnData,
		&tableInfo.LoadingDependenciesDatabase,
		&tableInfo.LoadingDependenciesTable,
		&tableInfo.LoadingDependentDatabase,
		&tableInfo.LoadingDependentTable,
	)
	tableInfo.HasOwnData = hasOwnData == 1
	tableInfo.IsTemporary = isTemporary == 1

	if err != nil {
		return ClickHouseTableFullInfo{}, err
	}

	re := regexp.MustCompile(`\S\((.*?)\)`)
	matches := re.FindStringSubmatch(tableInfo.EngineFull)
	if len(matches) > 1 {
		tableInfo.EngineParams = strings.Split(matches[1], ", ")
	} else {
		tableInfo.EngineParams = make([]string, 0)
	}

	if tableInfo.SortingKey != "" {
		tableInfo.OrderBy = strings.Split(tableInfo.SortingKey, ", ")
	} else {
		tableInfo.OrderBy = make([]string, 0)
	}

	cols, err := client.GetColumns(ctx, database, table)
	if err != nil {
		return ClickHouseTableFullInfo{}, err
	}
	tableInfo.Columns = cols

	return tableInfo, nil
}

func (client *ClickHouseClient) GetColumns(ctx context.Context, database, table string) (ClickHouseColumns, error) {
	query := fmt.Sprintf(
		`SELECT "name", "type", "comment" from "system"."columns"
where database = %s and table = %s
order by "position"`,
		QuoteValue(database),
		QuoteValue(table),
	)
	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var cols ClickHouseColumns

	for rows.Next() {
		var col ClickHouseColumn
		err := rows.Scan(&col.Name, &col.Type, &col.Comment)
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}

	return cols, nil
}

func (client *ClickHouseClient) AlterTable(ctx context.Context, currentTableName string, desiredTable ClickHouseTable) error {
	err := client.RenameTable(ctx, desiredTable.Database, currentTableName, desiredTable.Name)
	if err != nil {
		return err
	}

	currentTableInfo, err := client.GetTable(ctx, desiredTable.Database, desiredTable.Name)
	if err != nil {
		return err
	}
	currentTable := currentTableInfo.ToTable()

	err = client.AlterColumns(ctx, currentTable, desiredTable)
	if err != nil {
		return err
	}

	if len(desiredTable.OrderBy) > 1 {
		err = client.ModifyOrderBy(ctx, currentTable.Database, currentTable.Name, desiredTable.OrderBy)
		if err != nil {
			return err
		}
	}

	return nil
}

func (client *ClickHouseClient) RenameTable(ctx context.Context, db, from, to string) error {
	if from == to {
		return nil
	}
	query := fmt.Sprintf(
		"RENAME TABLE %s.%s TO %s.%s",
		QuoteID(db),
		QuoteID(from),
		QuoteID(db),
		QuoteID(to),
	)
	tflog.Info(ctx, "Renaming a table", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) ModifyOrderBy(ctx context.Context, db, table string, orderBy []string) error {
	query := fmt.Sprintf(
		"ALTER TABLE %s.%s MODIFY ORDER BY (%s)",
		QuoteID(db),
		QuoteID(table),
		QuoteListWithTicksAndJoin(orderBy),
	)
	tflog.Info(ctx, "Renaming a table", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) AlterColumns(ctx context.Context, currentTable, desiredTable ClickHouseTable) error {
	desiredColsSet := hashset.New[string](desiredTable.Columns.Names()...)
	currentColsSet := hashset.New[string](currentTable.Columns.Names()...)
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
			`ALTER TABLE %s.%s ADD COLUMN %s %s COMMENT %s`,
			QuoteID(desiredTable.Database),
			QuoteID(desiredTable.Name),
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
		if newCols.Contains(col.Name) {
			continue
		}

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
		if i < len(currentTable.Columns) {
			currentCol := currentTable.Columns[i]
			if desiredCol.Name == currentCol.Name {
				continue
			}
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

func (client *ClickHouseClient) DropTable(ctx context.Context, table ClickHouseTable, checkEmpty bool) error {
	if checkEmpty {
		empty, err := client.IsTableEmpty(ctx, table)
		if err != nil {
			return err
		}
		if !empty {
			return &TableIsNotEmptyError{table}
		}
	}

	query := fmt.Sprintf(
		`DROP TABLE %s.%s`,
		QuoteID(table.Database),
		QuoteID(table.Name),
	)
	tflog.Info(ctx, "Dropping a table", dict{"query": query})
	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) IsTableEmpty(ctx context.Context, table ClickHouseTable) (bool, error) {
	query := fmt.Sprintf(
		"select 1 from %s.%s limit 1",
		QuoteID(table.Database),
		QuoteID(table.Name),
	)
	tflog.Info(ctx, "Checking if table is empty", dict{"query": query})
	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return false, err
	}
	if rows.Next() {
		return false, nil
	}

	return true, nil
}
