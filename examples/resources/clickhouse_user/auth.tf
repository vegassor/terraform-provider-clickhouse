# CREATE USER user1 IDENTIFIED WITH sha256_hash
# BY '87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7'
# SALT 'aaabbb'
# HOST ANY
resource "clickhouse_user" "user_sha256_hash_and_salt" {
  name = "user3"

  identified_with = {
    sha256_hash = {
      hash = "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
      salt = "aaabbb"
    }
  }
}

# CREATE USER user2 IDENTIFIED WITH sha256_hash
# BY '87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7'
# HOST ANY
resource "clickhouse_user" "user_no_salt" {
  name = "user2"

  identified_with = {
    sha256_hash = {
      hash = "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
    }
  }
}

# CREATE USER user3 IDENTIFIED WITH sha256_password
# BY 'qwerty12345'
# HOST ANY
resource "clickhouse_user" "user_sha256_password" {
  name = "user3"

  identified_with = {
    sha256_password = "qwerty12345"
  }
}
