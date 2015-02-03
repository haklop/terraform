package openstack

import (
	"fmt"
	"testing"

	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/extensions/lbaas/vips"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
)

func TestAccOpenstackLBaaSVip(t *testing.T) {
	var vipId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenstackLBaaSVipDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testVipConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSVipExists(
						"openstack_lbaas_vip.tf_vip", &vipId),
				),
			},
			resource.TestStep{
				Config: testVipUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSVipExists(
						"openstack_lbaas_vip.tf_vip", &vipId),
				),
			},
			resource.TestStep{
				Config: testVipForceNewConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSVipExists(
						"openstack_lbaas_vip.tf_vip", &vipId),
				),
			},
		},
	})
}

func testAccCheckOpenstackLBaaSVipDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_lbaas_vip" {
			continue
		}

		networkClient, err := config.getNetworkClient()
		if err != nil {
			return err
		}

		_, err = vips.Get(networkClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Vip (%s) still exists.", rs.Primary.ID)
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 404 {
			return httpError
		}
	}

	return nil
}

func testAccCheckOpenstackLBaaSVipExists(n string, vipId *string) resource.TestCheckFunc {

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

		found, err := vips.Get(networkClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Vip not found")
		}

		vipId = &found.ID

		return nil
	}
}

const testVipConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
  name = "subnet_accept_test_updated"
  cidr = "10.20.30.0/24"
  ip_version = 4
  network_id = "${openstack_network.accept_test.id}"
}

resource "openstack_lbaas" "lbaas" {
    name = "http_lb"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    protocol = "HTTP"
    lb_method = "ROUND_ROBIN"
}

resource "openstack_lbaas_vip" "tf_vip" {
    pool_id = "${openstack_lbaas.lbaas.id}"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    name = "test vip"
    description = "tf vip"
    protocol = "HTTP"
    protocol_port = 80
}
`

const testVipUpdateConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
  name = "subnet_accept_test_updated"
  cidr = "10.20.30.0/24"
  ip_version = 4
  network_id = "${openstack_network.accept_test.id}"
}

resource "openstack_lbaas" "lbaas" {
    name = "http_lb"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    protocol = "HTTP"
    lb_method = "ROUND_ROBIN"
}

resource "openstack_lbaas_vip" "tf_vip" {
    pool_id = "${openstack_lbaas.lbaas.id}"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    name = "test vip2"
    description = "tf vip"
    protocol = "HTTP"
    protocol_port = 80
}
`

const testVipForceNewConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
  name = "subnet_accept_test_updated"
  cidr = "10.20.30.0/24"
  ip_version = 4
  network_id = "${openstack_network.accept_test.id}"
}

resource "openstack_lbaas" "lbaas" {
    name = "http_lb"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    protocol = "HTTP"
    lb_method = "ROUND_ROBIN"
}

resource "openstack_lbaas_vip" "tf_vip" {
    pool_id = "${openstack_lbaas.lbaas.id}"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    name = "test vip"
    description = "tf vip"
    protocol = "HTTP"
    protocol_port = 8080
}
`
