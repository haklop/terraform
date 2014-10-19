package openstack

import (
	"log"

	"github.com/haklop/gophercloud-extensions/network"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud"
)

func resourceFirewall() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallCreate,
		Read:   resourceFirewallRead,
		Delete: resourceFirewallDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"admin_state_up": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"shared": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceFirewallCreate(d *schema.ResourceData, meta interface{}) error {

	p := meta.(*Config)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return err
	}

	access := p.AccessProvider.(*gophercloud.Access)

	firewallConfiguration := network.NewFirewall{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		PolicyId:     d.Get("policy_id").(string),
		AdminStateUp: d.Get("admin_state_up").(bool),
		Shared:       d.Get("shared").(bool),
		TenantId:     access.Token.Tenant.Id,
	}

	log.Printf("[DEBUG] Create firewall: %#v", firewallConfiguration)

	firewall, err := networksApi.CreateFirewall(firewallConfiguration)
	if err != nil {
		return err
	}

	d.SetId(firewall.Id)

	return nil
}

func resourceFirewallRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Retrieve information about firewall: %s", d.Id())

	p := meta.(*Config)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return err
	}

	n, err := networksApi.GetFirewall(d.Id())
	if err != nil {
		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok {
			return err
		}

		if httpError.Actual == 404 {
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", n.Name)
	return nil
}

func resourceFirewallDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Destroy firewall: %s", d.Id())

	p := meta.(*Config)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return err
	}

	return networksApi.DeleteFirewall(d.Id())
}
