package openstack

import (
	"fmt"
	"testing"

	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/extensions/lbaas/Members"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
)

func TestAccOpenstackLBaaSMember(t *testing.T) {
	var MemberId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenstackLBaaSMemberDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testMemberConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSMemberExists(
						"openstack_lbaas_member.tf_accept_test", &MemberId),
				),
			},
			resource.TestStep{
				Config: testMemberForceNewConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSMemberExists(
						"openstack_lbaas_member.tf_accept_test", &MemberId),
				),
			},
		},
	})
}

func testAccCheckOpenstackLBaaSMemberDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_lbaas_member" {
			continue
		}

		networkClient, err := config.getNetworkClient()
		if err != nil {
			return err
		}

		_, err = members.Get(networkClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Member (%s) still exists.", rs.Primary.ID)
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 404 {
			return httpError
		}
	}

	return nil
}

func testAccCheckOpenstackLBaaSMemberExists(n string, memberId *string) resource.TestCheckFunc {

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

		found, err := members.Get(networkClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Member not found")
		}

		memberId = &found.ID

		return nil
	}
}

const testMemberConfig = `
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

resource "openstack_lbaas_member" "tf_accept_test" {
    port = 80
    instance_id = "cb42048e-32c9-495d-ac9a-a6f04ad0ec6b"
    pool_id = "${openstack_lbaas.lbaas.id}"
}
`

const testMemberForceNewConfig = `
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
    name = "tf_accept_test_http_lb"
    subnet_id = "${openstack_subnet.subnet_test.id}"
    protocol = "HTTP"
    lb_method = "ROUND_ROBIN"
}

resource "openstack_lbaas_member" "tf_accept_test" {
    port = 800
    instance_id = "cb42048e-32c9-495d-ac9a-a6f04ad0ec6b"
    pool_id = "${openstack_lbaas.lbaas.id}"
}
`
