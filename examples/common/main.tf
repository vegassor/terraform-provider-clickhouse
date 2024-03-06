import {
  to = clickhouse_table.my_table
  id = "default.my_table"
}

resource "clickhouse_table" "my_table" {
  name     = "my_table"
  database = "default"
  engine   = "Memory"

  columns = [
    {
      name = "id"
      type = "Int32"
    },
    {
      name = "name"
      type = "String"
    }
  ]
}
