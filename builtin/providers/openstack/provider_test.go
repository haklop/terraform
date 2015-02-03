package openstack

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"openstack": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("OS_AUTH_URL"); v == "" {
		t.Fatal("OS_AUTH_URL must be set for acceptance tests")
	}

	if v := os.Getenv("OS_USERNAME"); v == "" {
		t.Fatal("OS_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("OS_PASSWORD"); v == "" {
		t.Fatal("OS_PASSWORD must be set for acceptance tests.")
	}

	tenantID := os.Getenv("OS_TENANT_ID")
	tenantName := os.Getenv("OS_TENANT_NAME")

	if tenantID == "" && tenantName == "" {
		t.Fatal("OS_TENANT_ID or OS_TENANT_NAME must be set for acceptance tests.")
	}

	//if v := os.Getenv("OS_AT_DEFAULT_IMAGE_REF"); v == "" {
	//	t.Fatal("OS_AT_DEFAULT_IMAGE_REF must be set for acceptance tests.")
	//}

	//if v := os.Getenv("OS_AT_DEFAULT_FLAVOR_REF"); v == "" {
	//	t.Fatal("OS_AT_DEFAULT_IMAGE_REF must be set for acceptance tests.")
	//}
}
