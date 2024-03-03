resource "clickhouse_database" "my_db" {
  name = "my_db"
}

resource "clickhouse_table" "my_table" {
  database = clickhouse_database.my_db.name
  name     = "my_table"
  engine   = "Memory"

  columns = [
    {
      name = "col1"
      type = "String"
    },
    {
      name     = "col2"
      type     = "Float64"
      nullable = true
    },
  ]
}

resource "clickhouse_view" "my_view" {
  database = clickhouse_database.my_db.name
  name     = "my_view"
  query    = "SELECT * FROM ${clickhouse_table.my_table.full_name}"
}

resource "clickhouse_role" "my_role" {
  name = "my_role"
}

resource "clickhouse_privilege_grant" "my_grant" {
  access_type = "SELECT"
  grantee     = clickhouse_role.my_role.name
  grants = [
    {
      database = clickhouse_table.my_table.database
      table    = clickhouse_table.my_table.name
    },
    {
      database = clickhouse_view.my_view.database
      table    = clickhouse_view.my_view.name
    },
  ]
}

resource "clickhouse_user" "my_user" {
  name = "my_user"

  identified_with = {
    sha256_password = "qwerty12345"
  }
}

resource "clickhouse_role_grant" "my_grant" {
  role    = clickhouse_role.my_role.name
  grantee = clickhouse_user.my_user.name
}
