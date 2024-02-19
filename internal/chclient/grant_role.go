package chclient

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type roleGrantType string

const (
	roleGrantTypeUser roleGrantType = "user"
	roleGrantTypeRole roleGrantType = "role"
)

type RoleGrant struct {
	Role            string
	Grantee         string
	WithAdminOption bool
	RoleGrantType   roleGrantType
}

func (client *ClickHouseClient) GrantRole(ctx context.Context, roleName, grantee string, withAdminOption bool) error {
	query := fmt.Sprintf("GRANT %s TO %s", QuoteID(roleName), QuoteID(grantee))
	if withAdminOption {
		query = query + " WITH ADMIN OPTION"
	}
	tflog.Info(ctx, "Granting role", dict{"query": query})
	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetRoleGrant(ctx context.Context, roleName, grantee string) (RoleGrant, error) {
	var roleGrant RoleGrant
	query := fmt.Sprintf(
		`SELECT coalesce("user_name", ''), coalesce("role_name", ''),
"granted_role_name", "with_admin_option"
FROM system.role_grants
WHERE "user_name" = %s OR "role_name" = %s`,
		QuoteValue(grantee),
		QuoteValue(grantee),
	)

	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return roleGrant, err
	}

	if !rows.Next() {
		return roleGrant, &NotFoundError{Entity: "role grant", Name: roleName, Query: query}
	}

	var user, role string
	err = rows.Scan(&user, &role, &roleGrant.Grantee, &roleGrant.WithAdminOption)
	if err != nil {
		return roleGrant, err
	}

	if user == "" && role == "" {
		return roleGrant, &NotFoundError{Entity: "role grant", Name: roleName, Query: query}
	}
	if user != "" {
		roleGrant.Grantee = user
		roleGrant.RoleGrantType = roleGrantTypeUser
	} else {
		roleGrant.Grantee = role
		roleGrant.RoleGrantType = roleGrantTypeRole
	}

	roleGrant.Role = roleName
	return roleGrant, nil
}

func (client *ClickHouseClient) RevokeRole(ctx context.Context, roleName, grantee string) error {
	grant, err := client.GetRoleGrant(ctx, roleName, grantee)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		"REVOKE %s FROM %s",
		QuoteWithTicks(grant.Role),
		QuoteWithTicks(grant.Grantee),
	)

	tflog.Info(ctx, "Revoking role grant", dict{"query": query})
	return client.Conn.Exec(ctx, query)
}
