package openstack

import (
	"fmt"
	"testing"

	"github.com/ggiamarchi/gophercloud/openstack/compute/v2/extensions/secgroups"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
)

func TestAccOpenstackSecurityGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenstackSecurityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testSecGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackSecurityGroupExists(
						"openstack_security_group.http_rules", &secgroups.SecurityGroup{
							Name:        "http_rules",
							Description: "Access rules for http servers",
						}),
				),
			},
			resource.TestStep{
				Config: testSecGroupUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackSecurityGroupExists(
						"openstack_security_group.http_rules", &secgroups.SecurityGroup{
							Name:        "http_rules2",
							Description: "Access rules for http servers2",
						}),
				),
			},
			resource.TestStep{
				Config: testSecGroupForceNewConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackSecurityGroupExists(
						"openstack_security_group.http_rules", &secgroups.SecurityGroup{
							Name:        "http_rules2",
							Description: "Access rules for http servers2",
						}),
				),
			},
		},
	})
}

func testAccCheckOpenstackSecurityGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_security_group" {
			continue
		}

		computeClient, err := config.getComputeClient()
		if err != nil {
			return err
		}
		_, err = secgroups.Get(computeClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Security Group (%s) still exists.", rs.Primary.ID)
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 404 {
			return httpError
		}
	}

	return nil
}

func testAccCheckOpenstackSecurityGroupExists(n string, expected *secgroups.SecurityGroup) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		computeClient, err := config.getComputeClient()
		if err != nil {
			return err
		}
		found, err := secgroups.Get(computeClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		expected.ID = found.ID
		if expected.Name != found.Name {
			return fmt.Errorf("Expected:\n%#v\nFound:\n%#v", expected.Name, found.Name)
		}
		if expected.Description != found.Description {
			return fmt.Errorf("Expected:\n%#v\nFound:\n%#v", expected.Description, found.Description)
		}
		return nil
	}
}

const testSecGroupConfig = `
resource "openstack_security_group" "http_rules" {
	name = "http_rules"
	description = "Access rules for http servers"
	rule {
		from_port = 80
		to_port = 80
		ip_protocol = "tcp"
		cidr = "0.0.0.0/0"
	}
	rule {
		from_port = 443
		to_port = 443
		ip_protocol = "tcp"
		cidr = "0.0.0.0/0"
	}
}
`

const testSecGroupUpdateConfig = `
resource "openstack_security_group" "http_rules" {
	name = "http_rules2"
	description = "Access rules for http servers2"
	rule {
		from_port = 80
		to_port = 80
		ip_protocol = "tcp"
		cidr = "0.0.0.0/0"
	}
	rule {
		from_port = 443
		to_port = 443
		ip_protocol = "tcp"
		cidr = "0.0.0.0/0"
	}
}
`

const testSecGroupForceNewConfig = `
resource "openstack_security_group" "http_rules" {
	name = "http_rules2"
	description = "Access rules for http servers2"
	rule {
		from_port = 80
		to_port = 81
		ip_protocol = "tcp"
		cidr = "0.0.0.0/0"
	}
	rule {
		from_port = 443
		to_port = 443
		ip_protocol = "tcp"
		cidr = "0.0.0.0/0"
	}
}
`
