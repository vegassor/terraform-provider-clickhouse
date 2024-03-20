# Role grant can be imported by specifying id in format granted_role/grantee_name
terraform import clickhouse_role_grant.my_role_grant my_role/my_user
# or
terraform import clickhouse_role_grant.my_role_grant my_role/my_other_role
