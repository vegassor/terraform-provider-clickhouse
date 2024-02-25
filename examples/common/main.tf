resource "clickhouse_table" "my_table" {
  database = "default"
  name     = "my_table"

  engine   = "ReplacingMergeTree"
  order_by = ["id", "id2"]

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
      name = "value"
      type = "Float64"
    }
  ]

  settings = {
    index_granularity           = "8192"
    max_part_loading_threads    = "8"
    max_suspicious_broken_parts = "500"
  }
}