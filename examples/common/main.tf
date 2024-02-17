resource "clickhouse_database" "my_db" {
  name    = "my_db"
  engine  = "Atomic"
  comment = "Example DB"
}

resource "clickhouse_table" "my_table" {
  database = clickhouse_database.my_db.name
  name     = "my_table"
  engine   = "Memory"
  comment  = "Some comment"

  columns = [
    {
      name    = "col1"
      type    = "String"
      comment = "col1 comment"
    },
    {
      name = "col2"
      type = "Float64"
    }
  ]
}

resource "clickhouse_user" "my_user" {
  name = "my_user"

  identified_with = {
    sha256_password = "qwerty12345"
  }
}

resource "clickhouse_role" "my_role" {
  name = "my_role"
}

resource "clickhouse_role_grant" "my_role" {
  role    = clickhouse_role.my_role.name
  grantee = clickhouse_user.my_user.name
}

resource "clickhouse_privilege_grant" "to_user" {
  grantee     = clickhouse_user.my_user.name
  access_type = "SELECT"

  grants = [
    {
      database = clickhouse_database.my_db.name
      table    = clickhouse_table.my_table.name
      columns  = ["col1", "col2"]
    },
    {
      database = "default"
      table    = "*"
    },
  ]

  # TODO: implement partial revoke
  # revoke = [
  #   {
  #     database = clickhouse_database.my_another_db.name
  #     table    = "some_table"
  #   },
  # ]
}
