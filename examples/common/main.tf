import {
  to = clickhouse_database.my_db
  id = "my_db"
}

resource "clickhouse_database" "my_db" {
  name = "my_db"
}
