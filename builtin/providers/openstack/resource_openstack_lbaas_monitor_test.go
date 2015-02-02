package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/extensions/lbaas/monitors"
)

func TestAccOpenstackLBaaSMonitor(t *testing.T) {
	var monitorId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenstackLBaaSMonitorDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testMonitorConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSMonitorExists(
						"openstack_lbaas_monitor.tf_http_monitor", &monitorId),
				),
			},
			resource.TestStep{
				Config: testMonitorUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSMonitorExists(
						"openstack_lbaas_monitor.tf_http_monitor", &monitorId),
				),
			},
			resource.TestStep{
				Config: testMonitorForceNewConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackLBaaSMonitorExists(
						"openstack_lbaas_monitor.tf_http_monitor", &monitorId),
				),
			},
		},
	})
}

func testAccCheckOpenstackLBaaSMonitorDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_lbaas_monitor" {
			continue
		}

		networkClient, err := config.getNetworkClient()
		if err != nil {
			return err
		}

		_, err = monitors.Get(networkClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Monitor (%s) still exists.", rs.Primary.ID)
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 404 {
			return httpError
		}
	}

	return nil
}

func testAccCheckOpenstackLBaaSMonitorExists(n string, monitorId *string) resource.TestCheckFunc {

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

		found, err := monitors.Get(networkClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Subnet not found")
		}

		monitorId = &found.ID

		return nil
	}
}

const testMonitorConfig = `
resource "openstack_lbaas_monitor" "tf_http_monitor" {
    type = "HTTP"
    delay = 25
    timeout = 2
    max_retries = 2
    expected_codes = "200"
    url_path = "/"
    http_method = "GET"
}
`

const testMonitorUpdateConfig = `
resource "openstack_lbaas_monitor" "tf_http_monitor" {
    type = "HTTP"
    delay = 10
    timeout = 2
    max_retries = 2
    expected_codes = "204"
    url_path = "/healthcheck"
    http_method = "GET"
}
`

const testMonitorForceNewConfig = `
resource "openstack_lbaas_monitor" "tf_http_monitor" {
    type = "PING"
    delay = 25
    timeout = 5
    max_retries = 3
}
`
