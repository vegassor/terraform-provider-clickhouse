package chclient

import (
	"context"
	"fmt"
)

func (client *ClickHouseClient) GrantRole(ctx context.Context, name string) error {
	query := fmt.Sprintf("GRANT [ON CLUSTER cluster_name] role [,...] TO {user | another_role | CURRENT_USER} [,...] [WITH ADMIN OPTION] [WITH REPLACE OPTION]", QuoteID(name))
	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetRoleGrant(ctx context.Context, roleName string) (string, error) {
	return "", nil
}

func (client *ClickHouseClient) RevokeRoleGrant(ctx context.Context, roleName string) error {
	return nil
}
