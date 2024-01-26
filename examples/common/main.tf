resource "clickhouse_user" "my_user" {
  name = "my_user"

  identified_with = {
    sha256_hash = {
      hash = "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
      salt = "asdasa"
    }
  }

  hosts = {
    ip     = ["192.168.0.1/32", "10.0.0.0/8"]
    name   = ["localhost"]
    regexp = []
    like   = ["192.%.1.1"]
  }
}

resource "clickhouse_database" "internal_analytics" {
  name    = "internal_analytics_2"
  engine  = "Atomic"
  comment = "Example database"
}

resource "clickhouse_user" "nikolay" {
  name = "nikolay"

  identified_with = {
    sha256_password = random_password.nikolay_ch_password.result
  }

  default_database = clickhouse_database.internal_analytics.name
}

resource "random_password" "nikolay_ch_password" {
  length  = 32
  special = true
}