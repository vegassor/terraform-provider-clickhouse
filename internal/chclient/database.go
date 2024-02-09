package chclient

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"strings"
)

type DatabaseEngine int

func (v DatabaseEngine) String() string {
	switch v {
	case ATOMIC:
		return "Atomic"
	case MEMORY:
		return "Memory"
	default:
		return strconv.Itoa(int(v))
	}
}

func DatabaseEngineFromString(name string) DatabaseEngine {
	name = strings.ToLower(name)
	switch name {
	case "atomic":
		return ATOMIC
	case "memory":
		return MEMORY
	default:
		return -1
	}
}

const (
	MEMORY DatabaseEngine = iota
	ATOMIC
)

type ClickHouseDatabase struct {
	Name    string
	Engine  DatabaseEngine
	Comment string
}

func (client *ClickHouseClient) CreateDatabase(ctx context.Context, database ClickHouseDatabase) error {
	query := fmt.Sprintf(
		"CREATE DATABASE %s ENGINE = %s",
		QuoteID(database.Name),
		QuoteID(database.Engine.String()),
	)

	if database.Comment != "" {
		query += " COMMENT " + QuoteValue(database.Comment)
	}

	tflog.Info(ctx, "Creating a database", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) DropDatabase(ctx context.Context, database string) error {
	query := fmt.Sprintf(
		"DROP DATABASE %s SYNC",
		QuoteID(database),
	)

	tflog.Info(ctx, "Dropping a database", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetDatabase(ctx context.Context, name string) (ClickHouseDatabase, error) {
	query := fmt.Sprintf(
		`SELECT "name", "engine", "comment" 
		FROM "system"."databases"
		WHERE "name" = %s`,
		QuoteValue(name),
	)

	tflog.Debug(ctx, "Getting a database", dict{"query": query})

	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return ClickHouseDatabase{}, err
	}

	if !rows.Next() {
		return ClickHouseDatabase{}, fmt.Errorf(
			"could not find database %s: no records returned. Query: `%s`",
			name,
			query,
		)
	}

	var nameReceived string
	var engine string
	var comment string

	err = rows.Scan(&nameReceived, &engine, &comment)
	if err != nil {
		return ClickHouseDatabase{}, err
	}

	return ClickHouseDatabase{
		Name:    nameReceived,
		Engine:  DatabaseEngineFromString(engine),
		Comment: comment,
	}, nil
}
