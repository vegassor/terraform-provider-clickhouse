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
