package provider

import (
	"fmt"
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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_database.test", "name", "yourdb"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "engine", "Atomic"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "comment", ""),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccImportDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        chDatabaseResource("my_db"),
				ResourceName:  "clickhouse_database.test",
				ImportState:   true,
				ImportStateId: "my_db",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_database.test", "id", "my_db"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "name", "my_db"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "engine", "Atomic"),
					resource.TestCheckResourceAttr("clickhouse_database.test", "comment", ""),
				),
			},
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
