---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "clickhouse_view Resource - terraform-provider-clickhouse"
subcategory: ""
description: |-
  ClickHouse view. See: https://clickhouse.com/docs/en/sql-reference/statements/create/view#normal-view
---

# clickhouse_view (Resource)

ClickHouse view. See: https://clickhouse.com/docs/en/sql-reference/statements/create/view#normal-view



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `database` (String) ClickHouse database name
- `name` (String) View name in ClickHouse database
- `query` (String) View definition query. It should be a valid SELECT statement.