package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
)

func TestAccOpenstackNetwork(t *testing.T) {
	var networkId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenstackNetworkDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testNetworkConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackNetworkExists(
						"openstack_network.accept_test", &networkId),
				),
			},
		},
	})
}

func testAccCheckOpenstackNetworkDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_network" {
			continue
		}

		networkClient, err := config.getNetworkClient()
		if err != nil {
			return err
		}

		_, err = networks.Get(networkClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Network (%s) still exists.", rs.Primary.ID)
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 404 {
			return httpError
		}
	}

	return nil
}

func testAccCheckOpenstackNetworkExists(n string, networkId *string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		networkClient, err := config.getNetworkClient()
		if err != nil {
			return err
		}

		found, err := networks.Get(networkClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Network not found")
		}

		networkId = &found.ID

		return nil
	}
}

const testNetworkConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test Network"
}
`
