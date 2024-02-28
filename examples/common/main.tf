resource "clickhouse_table" "my_table" {
  database = "default"
  name     = "my_table"

  engine            = "ReplacingMergeTree"
#  engine_parameters = ["time"]
  order_by          = ["id", "id2"]

  columns = [
    {
      name     = "time"
      type     = "DateTime"
    },
    {
      name     = "id"
      type     = "Int64"
    },
    {
      name     = "id2"
      type     = "Int64"
    },
    {
      name = "value"
      type = "Float64"
    }
  ]
}
