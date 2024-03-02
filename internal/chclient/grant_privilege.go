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
	if grant.Database == "" {
		return errors.New("database is required")
	}

	what := grant.AccessType
	if len(grant.Columns) > 0 {
		what = fmt.Sprintf(
			"%s(%s)",
			grant.AccessType,
			strings.Join(QuoteList(grant.Columns, `"`), ", "),
		)
	}

	db := grant.Database
	table := grant.Table
	if db != "*" {
		db = QuoteID(db)
	}
	if table != "*" {
		table = QuoteID(table)
	}

	query := fmt.Sprintf(
		"GRANT %s ON %s.%s TO %s",
		what,
		db,
		table,
		grant.Grantee,
	)
	if grant.GrantOption {
		query += " WITH GRANT OPTION"
	}

	tflog.Info(ctx, "Granting privilege: %s", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) RevokePrivilege(ctx context.Context, grant PrivilegeGrant) error {
	what := grant.AccessType
	if len(grant.Columns) > 0 {
		what = fmt.Sprintf(
			"%s(%s)",
			grant.AccessType,
			strings.Join(QuoteList(grant.Columns, `"`), ", "),
		)
	}

	db := grant.Database
	table := grant.Table

	if db == "" || db == "*" {
		db = "*"
	} else {
		db = QuoteID(db)
	}

	if table == "" || table == "*" {
		table = "*"
	} else {
		table = QuoteID(table)
	}

	query := fmt.Sprintf(
		"REVOKE %s ON %s.%s FROM %s",
		what,
		db,
		table,
		grant.Grantee,
	)
	tflog.Info(ctx, "Revoking privilege", dict{
		"query":        query,
		"grantee":      grant.Grantee,
		"access_type":  grant.AccessType,
		"database":     grant.Database,
		"table":        grant.Table,
		"columns":      grant.Columns,
		"grant_option": grant.GrantOption,
	})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetPrivilegeGrants(ctx context.Context, grantee, accessType string) ([]PrivilegeGrant, error) {
	query := fmt.Sprintf(`SELECT
    coalesce(database, '*'),
    coalesce(table, '*'),
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
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		name := fmt.Sprintf("(grantee=%s, access_type=%s)")
		return nil, &NotFoundError{Entity: "privilege grant", Name: name, Query: query}
	}

	var grants []PrivilegeGrant
	for rows.Next() {
		var db, table string
		var columns []string
		var grantOption uint8
		err = rows.Scan(&db, &table, &columns, &grantOption)
		if err != nil {
			return nil, err
		}
		grants = append(grants, PrivilegeGrant{
			Grantee:     grantee,
			AccessType:  accessType,
			Database:    db,
			Table:       table,
			Columns:     columns,
			GrantOption: grantOption == 1,
		})
	}

	return grants, nil
}
