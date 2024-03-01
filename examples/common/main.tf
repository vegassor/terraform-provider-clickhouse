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
      name     = "col2"
      type     = "Float64"
      nullable = true
    },
    {
      name     = "col3"
      type     = "Float64"
      nullable = true
    },
  ]
}