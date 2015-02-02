package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/subnets"
)

func TestAccOpenstackSubnet(t *testing.T) {
	var subnetId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenstackSubnetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackSubnetExists(
						"openstack_subnet.subnet_test", "subnet_accept_test", &subnetId),
				),
			},
			resource.TestStep{
				Config: testSubnetUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackSubnetExists(
						"openstack_subnet.subnet_test", "subnet_accept_test_updated", &subnetId),
				),
			},
			resource.TestStep{
				Config: testSubnetForceNewConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackSubnetExists(
						"openstack_subnet.subnet_test", "subnet_accept_test_updated", &subnetId),
				),
			},
		},
	})
}

func testAccCheckOpenstackSubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_subnet" {
			continue
		}

		networkClient, err := config.getNetworkClient()
		if err != nil {
			return err
		}

		_, err = subnets.Get(networkClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Subnet (%s) still exists.", rs.Primary.ID)
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 404 {
			return httpError
		}
	}

	return nil
}

func testAccCheckOpenstackSubnetExists(n string, subnetName string, subnetId *string) resource.TestCheckFunc {

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

		found, err := subnets.Get(networkClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Subnet not found")
		}

		if found.Name != subnetName {
			return fmt.Errorf("Wrong Subnet name %s expected %s", found.Name, subnetName)
		}

		subnetId = &found.ID

		return nil
	}
}

const testSubnetConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
	name = "subnet_accept_test"
	cidr = "10.20.30.0/24"
	ip_version = 4
	network_id = "${openstack_network.accept_test.id}"
}
`

const testSubnetUpdateConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
	name = "subnet_accept_test_updated"
	cidr = "10.20.30.0/24"
	ip_version = 4
	network_id = "${openstack_network.accept_test.id}"
}
`

const testSubnetForceNewConfig = `
resource "openstack_network" "accept_test" {
    name = "accept_test_network"
}

resource "openstack_subnet" "subnet_test" {
	name = "subnet_accept_test_updated"
	cidr = "10.20.20.0/24"
	ip_version = 4
	network_id = "${openstack_network.accept_test.id}"
	allocation_pool = {
        start = "10.20.20.200"
        end = "10.20.20.250"
    }
}
`
