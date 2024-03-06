package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: chDatabaseResource("mydb"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_database.test", "name", "mydb"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "engine", "Atomic"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "comment", ""),
				),
			},
			// Update and Read testing
			{
				Config: chDatabaseResource("yourdb"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"clickhouse_database.test",
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_database.test", "name", "yourdb"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "engine", "Atomic"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "comment", ""),
				),
			},
			// Check if resource can be imported
			{
				Config:                  chDatabaseResource("yourdb"),
				ResourceName:            "clickhouse_database.test",
				ImportState:             true,
				ImportStateId:           "yourdb",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"engine", "comment", "id", "name"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func chDatabaseResource(name string) string {
	providerConfig := chProviderConfig()
	resources := fmt.Sprintf(`
resource "clickhouse_database" "test" {
  name = %[1]q
}
`, name)
	return providerConfig + resources
}

func chProviderConfig() string {
	return `terraform {
  required_providers {
    clickhouse = {
      source = "vegassor/clickhouse"
    }
  }
}

provider "clickhouse" {
  username = "default"
  password = "default"
  host     = "localhost"
  port     = 9000
  protocol = "native"
}
`
}
