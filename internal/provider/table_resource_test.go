package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTableResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: chMergeTreeTableResource("mytable"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_table.test", "id", "default.mytable"),
					resource.TestCheckResourceAttr("clickhouse_table.test", "database", "default"),
					resource.TestCheckResourceAttr("clickhouse_table.test", "name", "mytable"),
					resource.TestCheckResourceAttr("clickhouse_table.test", "settings.%", "1"),
					resource.TestCheckResourceAttr("clickhouse_table.test", "settings.index_granularity", "8192"),
					resource.TestCheckResourceAttr("clickhouse_table.test", "order_by.#", "1"),
					resource.TestCheckResourceAttr("clickhouse_table.test", "order_by.0", "date"),
				),
			},
			{
				Config:            chMergeTreeTableResource("mytable"),
				ResourceName:      "clickhouse_table.test",
				ImportState:       true,
				ImportStateId:     "default.mytable",
				ImportStateVerify: true,
			},
		},
	})
}

func chMergeTreeTableResource(name string) string {
	providerConfig := chProviderConfig()
	resources := fmt.Sprintf(`
resource "clickhouse_table" "test" {
  database = "default"
  name     = %[1]q
  engine   = "MergeTree"
  order_by = ["date"]

  columns = [
    {name = "date", type = "Date"},
    {name = "data", type = "Float64", nullable = true},
  ]
}
`, name)
	return providerConfig + resources
}
