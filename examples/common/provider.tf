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
  port     = 8123
  protocol = "http"
}
