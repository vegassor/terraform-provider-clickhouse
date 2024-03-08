package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: chUserSha256HashPasswordResource("myuser"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clickhouse_user.test", "id", "myuser"),
					resource.TestCheckResourceAttr("clickhouse_user.test", "name", "myuser"),
					resource.TestCheckResourceAttr("clickhouse_user.test", "identified_with.sha256_hash.hash", "2b915881367d1bd1ed3ab58b9fccc69fe4e3ee5492ab654ebd56c989ea6bd571"),
					resource.TestCheckResourceAttr("clickhouse_user.test", "hosts.ip.#", "2"),
					resource.TestCheckTypeSetElemAttr("clickhouse_user.test", "hosts.ip.*", "192.168.0.0/24"),
					resource.TestCheckTypeSetElemAttr("clickhouse_user.test", "hosts.ip.*", "192.168.1.1/32"),
					resource.TestCheckResourceAttr("clickhouse_user.test", "hosts.name.#", "1"),
					resource.TestCheckTypeSetElemAttr("clickhouse_user.test", "hosts.name.*", "localhost"),
				),
			},
			{
				Config:                  chUserSha256HashPasswordResource("myuser"),
				ResourceName:            "clickhouse_user.test",
				ImportState:             true,
				ImportStateId:           "myuser",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"identified_with"},
			},
		},
	})
}

func chUserSha256HashPasswordResource(name string) string {
	providerConfig := chProviderConfig()
	resources := fmt.Sprintf(`
resource "clickhouse_user" "test" {
  name = %[1]q

  identified_with = {
    sha256_hash = {
      hash = "2b915881367d1bd1ed3ab58b9fccc69fe4e3ee5492ab654ebd56c989ea6bd571"
    }
  }

  hosts = {
    ip     = ["192.168.0.0/24", "192.168.1.1/32"]
    name   = ["localhost"]
    regexp = []
    like   = []
  }
}
`, name)
	return providerConfig + resources
}
