resource "clickhouse_table" "my_table" {
  database = "default"
  name     = "my_table"
  engine   = "Memory"

  columns = [
    {
      name    = "col1"
      type    = "String"
      comment = "col1 comment"
    },
    {
      name = "col2"
      type = "Float64"
    },
  ]
}

resource "clickhouse_view" "my_view" {
  database = "default"
  name     = "my_view"
  query    = "SELECT col1 AS x FROM ${clickhouse_table.my_table.full_name}"
}
