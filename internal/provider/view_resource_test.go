package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccViewResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: chViewResource("myview"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_view.test", "id", "default.myview"),
					resource.TestCheckResourceAttr("clickhouse_view.test", "database", "default"),
					resource.TestCheckResourceAttr("clickhouse_view.test", "name", "myview"),
					resource.TestCheckResourceAttr("clickhouse_view.test", "query", "SELECT * FROM system.tables"),
				),
			},
			{
				Config:            chViewResource("myview"),
				ResourceName:      "clickhouse_view.test",
				ImportState:       true,
				ImportStateId:     "default.myview",
				ImportStateVerify: true,
			},
		},
	})
}

func chViewResource(name string) string {
	providerConfig := chProviderConfig()
	resources := fmt.Sprintf(`
resource "clickhouse_view" "test" {
  database = "default"
  name     = %[1]q

  query = "SELECT * FROM system.tables"
}
`, name)
	return providerConfig + resources
}
