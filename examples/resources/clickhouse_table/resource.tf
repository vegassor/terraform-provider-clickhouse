resource "clickhouse_database" "my_db" {
  name    = "my_db"
  engine  = "Atomic"
  comment = "Example DB"
}

resource "clickhouse_table" "my_table" {
  database = clickhouse_database.my_db.name
  name     = "my_table"
  engine   = "Memory"
  comment  = "Some comment"

  columns = [
    {
      name     = "col1"
      type     = "String"
      comment  = "col1 comment"
    },
    {
      name = "col2"
      type = "Float64"
    }
  ]
}
