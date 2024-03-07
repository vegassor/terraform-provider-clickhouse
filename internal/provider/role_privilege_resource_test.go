package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleGrantResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: chRolePrivilegeResource(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_role.grantee", "name", "grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_role.granted", "name", "granted_role"),

					resource.TestCheckResourceAttr("clickhouse_role_grant.test", "id", "granted_role/grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_role_grant.test", "role", "granted_role"),
					resource.TestCheckResourceAttr("clickhouse_role_grant.test", "grantee", "grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_role_grant.test", "with_admin_option", "false"),
				),
			},
			{
				Config:            chRolePrivilegeResource(false),
				ResourceName:      "clickhouse_role_grant.test",
				ImportState:       true,
				ImportStateId:     "granted_role/grantee_role",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoleGrantResourceWithAdminOption(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: chRolePrivilegeResource(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_role.grantee", "name", "grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_role.granted", "name", "granted_role"),

					resource.TestCheckResourceAttr("clickhouse_role_grant.test", "id", "granted_role/grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_role_grant.test", "role", "granted_role"),
					resource.TestCheckResourceAttr("clickhouse_role_grant.test", "grantee", "grantee_role"),
					resource.TestCheckResourceAttr("clickhouse_role_grant.test", "with_admin_option", "true"),
				),
			},
			{
				Config:            chRolePrivilegeResource(true),
				ResourceName:      "clickhouse_role_grant.test",
				ImportState:       true,
				ImportStateId:     "granted_role/grantee_role",
				ImportStateVerify: true,
			},
		},
	})
}

func chRolePrivilegeResource(withAdminOption bool) string {
	providerConfig := chProviderConfig()
	resources := fmt.Sprintf(`
resource "clickhouse_role" "grantee" {
  name = "grantee_role"
}

resource "clickhouse_role" "granted" {
  name = "granted_role"
}

resource "clickhouse_role_grant" "test" {
  role    = clickhouse_role.granted.name
  grantee = clickhouse_role.grantee.name

  with_admin_option = %[1]t
}
`, withAdminOption)
	return providerConfig + resources
}
