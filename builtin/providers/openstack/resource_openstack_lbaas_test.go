package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas/pools"
)

func TestAccOpenstackLBaaS(t *testing.T) {
	var poolId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenstackLBaaSDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testLBaaSConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSExists(
						"openstack_lbaas.lbaas", &poolId),
				),
			},
			resource.TestStep{
				Config: testLBaaSUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSExists(
						"openstack_lbaas.lbaas", &poolId),
				),
			},
			resource.TestStep{
				Config: testLBaaSForceNewConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSExists(
						"openstack_lbaas.lbaas", &poolId),
				),
			},
		},
	})
}

func testAccCheckOpenstackLBaaSDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_lbaas" {
			continue
		}

		networkClient, err := config.getNetworkClient()
		if err != nil {
			return err
		}

		_, err = pools.Get(networkClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Pool (%s) still exists.", rs.Primary.ID)
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 404 {
			return httpError
		}
	}

	return nil
}

func testAccCheckOpenstackLBaaSExists(n string, poolId *string) resource.TestCheckFunc {

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

		found, err := pools.Get(networkClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Pool not found")
		}

		poolId = &found.ID

		return nil
	}
}

const testLBaaSConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
  name = "subnet_accept_test"
  cidr = "10.20.30.0/24"
  ip_version = 4
  network_id = "${openstack_network.accept_test.id}"
}

resource "openstack_lbaas" "lbaas" {
    name = "accept_test_http_lb"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    protocol = "HTTP"
    lb_method = "ROUND_ROBIN"
}
`

const testLBaaSUpdateConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
  name = "subnet_accept_test"
  cidr = "10.20.30.0/24"
  ip_version = 4
  network_id = "${openstack_network.accept_test.id}"
}

resource "openstack_lbaas" "lbaas" {
    name = "accept_test_http_lb_updated"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    protocol = "HTTP"
    lb_method = "LEAST_CONNECTIONS"
}
`

const testLBaaSForceNewConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
  name = "subnet_accept_test"
  cidr = "10.20.30.0/24"
  ip_version = 4
  network_id = "${openstack_network.accept_test.id}"
}

resource "openstack_lbaas" "lbaas" {
    name = "accept_test_http_lb_updated"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    protocol = "HTTPS"
    lb_method = "LEAST_CONNECTIONS"
}
`
