terraform {
  required_providers {
    clickhouse = {
      source = "vegassor/clickhouse"
    }
  }
}

provider "clickhouse" {
  username = "default"
  password = "default"
  host     = "localhost"
  port     = 9000
}
