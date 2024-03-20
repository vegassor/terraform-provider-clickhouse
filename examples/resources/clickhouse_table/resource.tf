resource "clickhouse_database" "my_db" {
  name    = "my_db"
  engine  = "Atomic"
  comment = "Example DB"
}

resource "clickhouse_table" "my_table" {
  database = "default"
  name     = "my_table"

  engine       = "ReplacingMergeTree"
  order_by     = ["id", "id2"]
  partition_by = "toYYYYMM(time)"

  columns = [
    {
      name = "time"
      type = "DateTime"
    },
    {
      name = "id"
      type = "Int64"
    },
    {
      name = "id2"
      type = "Int64"
    },
    {
      name     = "value"
      type     = "Float64"
      nullable = true
    }
  ]
}
