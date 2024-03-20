resource "clickhouse_user" "my_user" {
  name = "my_user"
  identified_with = {
    sha256_password = "qwerty12345"
  }
}

resource "clickhouse_role" "my_role" {
  name = "my_role"
}

resource "clickhouse_role_grant" "my_role_grant" {
  role    = clickhouse_role.my_role.name
  grantee = clickhouse_user.my_user.name
}