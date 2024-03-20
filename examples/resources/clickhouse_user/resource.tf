# CREATE USER my_user IDENTIFIED WITH sha256_hash
# BY '87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7'
# SALT 'aaabbb'
# HOST IP '192.168.0.0/24', IP '192.168.1.1/32'
resource "clickhouse_user" "my_user" {
  name = "my_user"

  identified_with = {
    sha256_hash = {
      hash = "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
      salt = "aaabbb"
    }
  }

  hosts = {
    ip = ["192.168.0.0/24", "192.168.1.1/32"]
  }
}
