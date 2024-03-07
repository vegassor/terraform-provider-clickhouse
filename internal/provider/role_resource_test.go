package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: chRoleResource("myrole"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_role.test", "id", "myrole"),
					resource.TestCheckResourceAttr("clickhouse_role.test", "name", "myrole"),
				),
			},
			{
				Config:            chRoleResource("myrole"),
				ResourceName:      "clickhouse_role.test",
				ImportState:       true,
				ImportStateId:     "myrole",
				ImportStateVerify: true,
			},
		},
	})
}

func chRoleResource(name string) string {
	providerConfig := chProviderConfig()
	resources := fmt.Sprintf(`
resource "clickhouse_role" "test" {
  name = %[1]q
}
`, name)
	return providerConfig + resources
}
