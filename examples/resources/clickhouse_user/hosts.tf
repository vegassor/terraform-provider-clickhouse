# CREATE USER user1 IDENTIFIED WITH sha256_password
# BY 'qwerty12345'
# HOST IP '192.168.0.0/24', IP '192.168.1.1/32', NAME 'localhost'
resource "clickhouse_user" "user_with_configured_hosts" {
  name = "user1"

  identified_with = {
    sha256_password = "qwerty12345"
  }

  hosts = {
    ip     = ["192.168.0.0/24", "192.168.1.1/32"]
    name   = ["localhost"]
    regexp = []
    like   = []
  }
}


# CREATE USER user3 IDENTIFIED WITH sha256_password
# BY 'qwerty12345'
# HOST ANY
resource "clickhouse_user" "user_with_any_allowed_host" {
  name = "user2"

  identified_with = {
    sha256_password = "qwerty12345"
  }
}


# CREATE USER user3 IDENTIFIED WITH sha256_password
# BY 'qwerty12345'
# HOST NONE
resource "clickhouse_user" "user_with_no_allowed_host" {
  name = "user3"

  identified_with = {
    sha256_password = "qwerty12345"
  }

  hosts = {}
}

