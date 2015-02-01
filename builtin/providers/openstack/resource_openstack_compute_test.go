package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

func TestAccOpenstackCompute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenstackComputeDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testComputeMinimalConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackComputeExists(
						"openstack_compute.compute_test", &servers.Server{
							Name: "compute_test",
							Image: map[string]interface{}{
								"id": "594f1287-9de3-4f3e-b82a-6ad223943ab2",
							},
							Flavor: map[string]interface{}{
								"id": "100",
							},
						}),
				),
			},
			resource.TestStep{
				Config: testComputeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackComputeExists(
						"openstack_compute.compute_test", &servers.Server{
							Name: "compute_test",
							Image: map[string]interface{}{
								"id": "594f1287-9de3-4f3e-b82a-6ad223943ab2",
							},
							Flavor: map[string]interface{}{
								"id": "100",
							},
						}),
				),
			},
			resource.TestStep{
				Config: testComputeUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackComputeExists(
						"openstack_compute.compute_test", &servers.Server{
							Name: "compute_test2",
							Image: map[string]interface{}{
								"id": "594f1287-9de3-4f3e-b82a-6ad223943ab2",
							},
							Flavor: map[string]interface{}{
								"id": "100",
							},
						}),
				),
			},
			resource.TestStep{
				Config: testComputeResizeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackComputeExists(
						"openstack_compute.compute_test", &servers.Server{
							Name: "compute_test2",
							Image: map[string]interface{}{
								"id": "594f1287-9de3-4f3e-b82a-6ad223943ab2",
							},
							Flavor: map[string]interface{}{
								"id": "101",
							},
						}),
				),
			},
			resource.TestStep{
				Config: testComputeForceNewConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenstackComputeExists(
						"openstack_compute.compute_test", &servers.Server{
							Name: "compute_test2",
							Image: map[string]interface{}{
								"id": "594f1287-9de3-4f3e-b82a-6ad223943ab2",
							},
							Flavor: map[string]interface{}{
								"id": "101",
							},
						}),
				),
			},
		},
	})
}

func testAccCheckOpenstackComputeDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_compute" {
			continue
		}

		computeClient, err := config.getComputeClient()
		if err != nil {
			return err
		}
		_, err = servers.Get(computeClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Compute (%s) still exists.", rs.Primary.ID)
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 404 {
			return httpError
		}
	}

	return nil
}

func testAccCheckOpenstackComputeExists(n string, expected *servers.Server) resource.TestCheckFunc {

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
		found, err := servers.Get(computeClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		expected.ID = found.ID
		// Not using reflect.DeepEqual because of too many fields in servers.Server
		if expected.Name != found.Name {
			return fmt.Errorf("Name field does not match: %s instead of %s", found.Name, expected.Name)
		}
		if expected.Image["id"] != found.Image["id"] {
			return fmt.Errorf("Image id field does not match: %s instead of %s", found.Image["id"], expected.Image["id"])
		}
		if expected.Flavor["id"] != found.Flavor["id"] {
			return fmt.Errorf("Flavor id field does not match: %s instead of %s", found.Flavor["id"], expected.Flavor["id"])
		}
		return nil
	}
}

const testComputeMinimalConfig = `
resource "openstack_compute" "compute_test" {
	name = "compute_test"
	image_ref = "594f1287-9de3-4f3e-b82a-6ad223943ab2"
	flavor_ref = "100"
	networks = ["183a3f67-3415-4bca-a5ab-0d1924d6f0e9"]
}
`

const testComputeConfig = `
resource "openstack_compute" "compute_test" {
	name = "compute_test"
	image_ref = "594f1287-9de3-4f3e-b82a-6ad223943ab2"
	flavor_ref = "100"
	key_pair_name = "jownmac"
	floating_ip_pool = "PublicNetwork-01"
	networks = ["183a3f67-3415-4bca-a5ab-0d1924d6f0e9"]
}
`

const testComputeUpdateConfig = `
resource "openstack_compute" "compute_test" {
	name = "compute_test2"
	image_ref = "594f1287-9de3-4f3e-b82a-6ad223943ab2"
	flavor_ref = "100"
	key_pair_name = "jownmac"
	floating_ip_pool = "PublicNetwork-01"
	networks = ["183a3f67-3415-4bca-a5ab-0d1924d6f0e9"]
}
`

const testComputeResizeConfig = `
resource "openstack_compute" "compute_test" {
	name = "compute_test2"
	image_ref = "594f1287-9de3-4f3e-b82a-6ad223943ab2"
	flavor_ref = "101"
	key_pair_name = "jownmac"
	floating_ip_pool = "PublicNetwork-01"
	networks = ["183a3f67-3415-4bca-a5ab-0d1924d6f0e9"]
}
`

const testComputeForceNewConfig = `
resource "openstack_compute" "compute_test" {
	name = "compute_test2"
	image_ref = "594f1287-9de3-4f3e-b82a-6ad223943ab2"
	flavor_ref = "101"
	key_pair_name = "tf-openstack"
	floating_ip_pool = "PublicNetwork-01"
	networks = ["183a3f67-3415-4bca-a5ab-0d1924d6f0e9"]
}
`
