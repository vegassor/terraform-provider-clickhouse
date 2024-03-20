# Privilege grant can be imported by specifying id in format access_type/grantee
terraform import clickhouse_role_grant.my_role_grant SELECT/my_user
# or
terraform import clickhouse_role_grant.my_role_grant "ALTER MODIFY COLUMN/my_role"
