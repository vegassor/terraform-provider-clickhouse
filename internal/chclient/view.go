package chclient

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ClickHouseView struct {
	Database string
	Name     string
	Query    string
}

func (client *ClickHouseClient) CreateView(ctx context.Context, view ClickHouseView, replace bool) error {
	replacePart := ""
	if replace {
		replacePart = "OR REPLACE"
	}

	query := fmt.Sprintf(
		"CREATE %s VIEW %s.%s AS %s",
		replacePart,
		QuoteID(view.Database),
		QuoteID(view.Name),
		view.Query,
	)

	tflog.Info(ctx, "Creating a view", dict{"query": query})

	return client.Conn.Exec(ctx, query)
}
