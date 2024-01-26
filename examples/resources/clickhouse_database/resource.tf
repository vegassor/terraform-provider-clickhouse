resource "clickhouse_database" "my_db" {
  name    = "my_db"
  engine  = "Atomic"
  comment = "Example DB"
}

resource "clickhouse_database" "in_mem_db" {
  name    = "in_memory_db"
  engine  = "Memory"
  comment = "Some description..."
}
