package chclient

import (
	"context"
	"fmt"
)

func (client *ClickHouseClient) CreateRole(ctx context.Context, name string) error {
	query := fmt.Sprintf("CREATE ROLE %s", QuoteID(name))
	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetRole(ctx context.Context, roleName string) (string, error) {
	query := `select "name" from "system"."roles" where "name" = ` + QuoteValue(roleName)
	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return "", err
	}

	if !rows.Next() {
		return "", &NotFoundError{Entity: "role", Name: roleName, Query: query}
	}

	receivedName := ""
	err = rows.Scan(&receivedName)
	if err != nil {
		return "", err
	}

	return receivedName, nil
}

func (client *ClickHouseClient) RenameRole(ctx context.Context, from, to string) error {
	return nil
}

func (client *ClickHouseClient) DropRole(ctx context.Context, name string) error {
	return nil
}
