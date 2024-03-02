package chclient

import (
	"context"
	"fmt"
	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
	"time"
)

type ClickHouseColumn struct {
	Name     string
	Type     string
	Comment  string
	Nullable bool
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
	Database      string
	Name          string
	Comment       string
	Engine        string
	EngineParams  []string
	PartitionBy   string
	OrderBy       []string
	PrimaryKeyArr []string
	Settings      map[string]string
	Columns       ClickHouseColumns
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

	Columns       ClickHouseColumns
	EngineParams  []string
	PartitionBy   string
	OrderBy       []string
	PrimaryKeyArr []string
	Settings      map[string]string
}

func (info ClickHouseTableFullInfo) ToTable() ClickHouseTable {
	return ClickHouseTable{
		Database:     info.Database,
		Name:         info.Name,
		Comment:      info.Comment,
		Engine:       info.Engine,
		EngineParams: info.EngineParams,
		OrderBy:      info.OrderBy,
		Settings:     info.Settings,
		Columns:      info.Columns,
	}
}

func (col ClickHouseColumn) String() string {
	result := fmt.Sprintf(
		"%s %s",
		QuoteWithTicks(col.Name),
		QuoteID(col.Type),
	)

	if col.Nullable {
		result += " NULL"
	}

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
	if table.PartitionBy != "" {
		query += " PARTITION BY " + table.PartitionBy + " "
	}

	if len(table.OrderBy) > 0 {
		query += " ORDER BY (" + QuoteListWithTicksAndJoin(table.OrderBy) + ")"
	}

	if len(table.PrimaryKeyArr) > 0 {
		query += " PRIMARY KEY (" + QuoteListWithTicksAndJoin(table.PrimaryKeyArr) + ")"
	}

	if len(table.Settings) > 0 {
		query += " SETTINGS " + QuoteMapAndJoin(table.Settings)
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

	tableInfo.PartitionBy = tableInfo.PartitionKey

	if tableInfo.SortingKey != "" {
		tableInfo.OrderBy = strings.Split(tableInfo.SortingKey, ", ")
	} else {
		tableInfo.OrderBy = make([]string, 0)
	}

	if tableInfo.PrimaryKey != "" {
		tableInfo.PrimaryKeyArr = strings.Split(tableInfo.PrimaryKey, ", ")
	} else {
		tableInfo.PrimaryKeyArr = make([]string, 0)
	}

	tableInfo.Settings = MustParseSettings(tableInfo.EngineFull)
	tableInfo.EngineParams = MustParseEngineParams(tableInfo.EngineFull)

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

		if strings.Contains(col.Type, "Nullable") {
			col.Nullable = true
			col.Type = strings.ReplaceAll(col.Type, "Nullable(", "")
			col.Type = strings.ReplaceAll(col.Type, ")", "")
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

	err = client.AlterTableSettings(ctx, currentTable, desiredTable)
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

func (client *ClickHouseClient) AlterTableSettings(ctx context.Context, currentTable, desiredTable ClickHouseTable) error {
	desiredSettingsSet := hashset.New[string]()
	for k := range desiredTable.Settings {
		desiredSettingsSet.Add(k)
	}

	currentSettingsSet := hashset.New[string]()
	for k := range currentTable.Settings {
		currentSettingsSet.Add(k)
	}

	err := client.ModifyTableSettings(ctx, desiredTable.Database, desiredTable.Name, desiredTable.Settings)
	if err != nil {
		return err
	}

	resetSettings := currentSettingsSet.Difference(desiredSettingsSet)
	return client.ResetTableSettings(ctx, desiredTable.Database, desiredTable.Name, resetSettings.Values()...)
}

func (client *ClickHouseClient) ModifyTableSettings(ctx context.Context, db, table string, settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"ALTER TABLE %s.%s MODIFY SETTING %s",
		QuoteID(db),
		QuoteID(table),
		QuoteMapAndJoin(settings),
	)
	tflog.Info(ctx, "Modifying table settings", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) ResetTableSettings(ctx context.Context, db, table string, settingsNames ...string) error {
	if len(settingsNames) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"ALTER TABLE %s.%s RESET SETTING %s",
		QuoteID(db),
		QuoteID(table),
		QuoteListWithTicksAndJoin(settingsNames),
	)
	tflog.Info(ctx, "Resetting table settings", dict{"query": query})

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
			Detail: fmt.Sprintf("cannot update columns of table %s.%s: "+
				"desired config does not contain columns from previous config. "+
				"If you did not try to rename or delete columns, it is a bug in the Client",
				desiredTable.Database,
				desiredTable.Name,
			),
		}
	}

	for _, colName := range newCols.Values() {
		colType := desiredColsMap[colName].Type
		if desiredColsMap[colName].Nullable {
			colType = "Nullable(" + colType + ")"
		}
		query := fmt.Sprintf(
			`ALTER TABLE %s.%s ADD COLUMN %s %s COMMENT %s`,
			QuoteID(desiredTable.Database),
			QuoteID(desiredTable.Name),
			QuoteID(colName),
			colType,
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

		colType := col.Type
		if col.Nullable {
			colType = "Nullable(" + colType + ")"
		}
		query := fmt.Sprintf(`ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s COMMENT %s`,
			QuoteID(desiredTable.Database),
			QuoteID(desiredTable.Name),
			QuoteID(col.Name),
			colType,
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
		colType := desiredCol.Type
		if desiredCol.Nullable {
			colType = "Nullable(" + colType + ")"
		}
		desiredIdx := i
		var query string

		if desiredIdx == 0 {
			query = fmt.Sprintf(
				`ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s FIRST`,
				QuoteID(desiredTable.Database),
				QuoteID(desiredTable.Name),
				QuoteID(desiredCol.Name),
				colType,
			)
		} else {
			prevColName := desiredTable.Columns[desiredIdx-1].Name
			query = fmt.Sprintf(
				`ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s AFTER %s`,
				QuoteID(desiredTable.Database),
				QuoteID(desiredTable.Name),
				QuoteID(desiredCol.Name),
				colType,
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
