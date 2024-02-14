package chclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

type PrivilegeGrant struct {
	Grantee     string
	AccessType  string
	Database    string
	Table       string
	Columns     []string
	GrantOption bool
}

func (client *ClickHouseClient) GrantPrivilege(ctx context.Context, grant PrivilegeGrant) error {
	what := grant.AccessType
	if len(grant.Columns) > 0 {
		what = fmt.Sprintf(
			"%s(%s)",
			grant.AccessType,
			strings.Join(QuoteList(grant.Columns, `"`), ", "),
		)
	}

	if grant.Database == "" {
		return errors.New("database is required")
	}
	on := QuoteID(grant.Database)
	if grant.Table != "" {
		on = on + "." + QuoteID(grant.Table)
	}

	query := fmt.Sprintf(
		"GRANT %s ON %s TO %s",
		what,
		on,
		grant.Grantee,
	)
	tflog.Info(ctx, "Granting privilege: %s", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetPrivilegeGrants(ctx context.Context, grantee, accessType string) ([]PrivilegeGrant, error) {
	query := fmt.Sprintf(`SELECT
    database,
    table,
    groupArray(column) as columns,
    grant_option
FROM "system"."grants"
WHERE
    (role_name = %s OR user_name = %s)
    AND access_type = %s
    AND is_partial_revoke = 0
GROUP BY
    access_type,
    role_name,
    database,
    table,
    grant_option`,
		QuoteValue(grantee),
		QuoteValue(grantee),
		QuoteValue(accessType),
	)

	tflog.Info(ctx, "Querying privileges", dict{"query": query})
	rows, err := client.Conn.Query(ctx, query)
	tflog.Info(ctx, "Aaaa", dict{"ccc": err})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []PrivilegeGrant
	for rows.Next() {
		var db, table string
		var columns []string
		var grantOption uint8
		err = rows.Scan(&db, &table, &columns, &grantOption)
		if err != nil {
			return nil, err
		}
		tflog.Info(ctx, "ITERATION", dict{"ctx": rows, "ccc": err})
		grants = append(grants, PrivilegeGrant{
			Grantee:     grantee,
			AccessType:  accessType,
			Database:    db,
			Table:       table,
			Columns:     columns,
			GrantOption: grantOption == 1,
		})
	}

	tflog.Info(ctx, "What I got: %s", dict{"query": query})

	return grants, nil
}
