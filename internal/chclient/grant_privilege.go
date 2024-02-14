package chclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

type PrivilegeGrant struct {
	Grantee   string
	Database  string
	Entity    string
	Privilege string
	Columns   []string
}

func (client *ClickHouseClient) GrantPrivilege(ctx context.Context, grant PrivilegeGrant) error {
	what := grant.Privilege
	if len(grant.Columns) > 0 {
		what = fmt.Sprintf(
			"%s(%s)",
			grant.Privilege,
			strings.Join(QuoteList(grant.Columns, `"`), ", "),
		)
	}

	if grant.Database == "" {
		return errors.New("database is required")
	}
	on := QuoteID(grant.Database)
	if grant.Entity != "" {
		on = on + "." + QuoteID(grant.Entity)
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

func (client *ClickHouseClient) GetPrivilegeGrant(ctx context.Context, grantee, accessType string) (PrivilegeGrant, error) {
	query := fmt.Sprintf(`SELECT
    grant_option,
    groupArray(database) as databases,
    groupArray(table) as tables,
    groupArray(column) as columns
FROM system.grants
WHERE
    (role_name = %s OR user_name = %s)
    AND access_type = %s
    AND is_partial_revoke = 0
GROUP BY
    access_type,
    role_name,
    grant_option`,
		QuoteValue(grantee),
		QuoteValue(grantee),
		QuoteValue(accessType),
	)

	tflog.Info(ctx, "Querying privileges: %s", dict{"query": query})
	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return PrivilegeGrant{}, err
	}
	defer rows.Close()

	var r []struct {
		GrantOption int8     `db:"grant_option"`
		Databases   []string `db:"databases"`
		Tables      []string `db:"tables"`
		Columns     []string `db:"columns"`
	}
	err = rows.Scan(&r)
	if err != nil {
		return PrivilegeGrant{}, err
	}
	//for rows.Next() {
	//	var dbs, tables, columns []string
	//	var grantOption int8
	//
	//}

	tflog.Info(ctx, "What I got: %s", dict{"query": query})

	return PrivilegeGrant{}, nil
}
