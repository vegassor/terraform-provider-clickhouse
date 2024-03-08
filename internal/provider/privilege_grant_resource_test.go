package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrivilegeGrantResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: chPrivilegeGrantResource(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_role.grantee", "name", "grantee_role"),

					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "id", "ALTER MODIFY COLUMN/grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "access_type", "ALTER MODIFY COLUMN"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grantee", "grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.#", "1"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.database", "default"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.table", "mytable"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.columns.#", "2"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.columns.0", "col1"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.columns.1", "col2"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.with_grant_option", "false"),
				),
			},
			{
				Config:            chPrivilegeGrantResource(false),
				ResourceName:      "clickhouse_privilege_grant.test",
				ImportState:       true,
				ImportStateId:     "ALTER MODIFY COLUMN/grantee_role",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPrivilegeGrantWithGrantOptionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: chPrivilegeGrantResource(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_role.grantee", "name", "grantee_role"),

					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "id", "ALTER MODIFY COLUMN/grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "access_type", "ALTER MODIFY COLUMN"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grantee", "grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.#", "1"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.database", "default"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.table", "mytable"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.columns.#", "2"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.columns.0", "col1"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.columns.1", "col2"),
					resource.TestCheckResourceAttr("clickhouse_privilege_grant.test", "grants.0.with_grant_option", "true"),
				),
			},
			{
				Config:            chPrivilegeGrantResource(true),
				ResourceName:      "clickhouse_privilege_grant.test",
				ImportState:       true,
				ImportStateId:     "ALTER MODIFY COLUMN/grantee_role",
				ImportStateVerify: true,
			},
		},
	})
}

func chPrivilegeGrantResource(withGrantOption bool) string {
	providerConfig := chProviderConfig()
	resources := fmt.Sprintf(`
resource "clickhouse_role" "grantee" {
  name = "grantee_role"
}

resource "clickhouse_privilege_grant" "test" {
  access_type = "ALTER MODIFY COLUMN"
  grantee = clickhouse_role.grantee.name

  grants = [
    {
      database = "default"
      table    = "mytable"
      columns  = ["col1", "col2"]

      with_grant_option = %[1]t
    },
  ]
}
`, withGrantOption)
	return providerConfig + resources
}
