package openstack

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_url": &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: envDefaultFunc("OS_AUTH_URL", false),
				Required:    true,
			},

			"username": &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: envDefaultFunc("OS_USERNAME", false),
				Required:    true,
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: envDefaultFunc("OS_PASSWORD", false),
				Required:    true,
			},

			"tenant_id": &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: envDefaultFunc("OS_TENANT_ID", true),
				Required:    true,
			},

			"tenant_name": &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: envDefaultFunc("OS_TENANT_NAME", true),
				Required:    true,
			},

			"region": &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: envDefaultFunc("OS_REGION_NAME", true),
				Required:    true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"openstack_network":         resourceNetwork(),
			// "openstack_subnet":          resourceSubnet(),
			"openstack_router":          resourceRouter(),
			// "openstack_security_group":  resourceSecurityGroup(),
			"openstack_compute":         resourceCompute(),
			// "openstack_lbaas":           resourceLBaaS(),
			// "openstack_firewall":        resourceFirewall(),
			// "openstack_firewall_policy": resourceFirewallPolicy(),
			// "openstack_firewall_rule":   resourceFirewallRule(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func envDefaultFunc(k string, emptyStringifNotSpecified bool) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		if v := os.Getenv(k); v != "" {
			return v, nil
		}
		if emptyStringifNotSpecified {
			return "", nil
		}
		return nil, nil
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	tenantID := d.Get("tenant_id").(string)
	tenantName := d.Get("tenant_name").(string)

	if tenantID == "" && tenantName == "" {
		return nil, fmt.Errorf("tenant_id or tenant_name must be provided")
	}

	config := Config{
		AuthUrl:    d.Get("auth_url").(string),
		Username:   d.Get("username").(string),
		Password:   d.Get("password").(string),
		TenantId:   d.Get("tenant_id").(string),
		TenantName: d.Get("tenant_name").(string),
		Region:     d.Get("region").(string),
	}

	if err := config.NewClient(); err != nil {
		return nil, err
	}

	return &config, nil
}
