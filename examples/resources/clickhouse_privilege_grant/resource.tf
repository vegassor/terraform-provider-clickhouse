resource "clickhouse_table" "my_table" {
  database = "default"
  name     = "my_table"
  engine   = "Memory"

  columns = [
    {
      name = "col1"
      type = "String"
    },
    {
      name = "col2"
      type = "Float64"
    }
  ]
}

resource "clickhouse_role" "my_role" {
  name = "my_role"
}

# GRANT SELECT(col1) ON default.my_table TO my_role
# GRANT SELECT ON system.tables TO my_role
resource "clickhouse_privilege_grant" "to_role" {
  grantee     = clickhouse_role.my_role.name
  access_type = "SELECT"

  grants = [
    {
      database = clickhouse_table.my_table.database
      table    = clickhouse_table.my_table.name
      columns  = ["col1"]
    },
    {
      database = "system"
      table    = "tables"
    },
  ]
}
