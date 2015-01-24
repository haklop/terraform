package openstack

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
)

func resourceNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkCreate,
		Read:   resourceNetworkRead,
		Update: resourceNetworkUpdate,
		Delete: resourceNetworkDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
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

func resourceNetworkCreate(d *schema.ResourceData, meta interface{}) error {

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	adminStateUp := d.Get("admin_state_up").(bool)
	shared := d.Get("shared").(bool)

	opts := networks.CreateOpts{
		Name:         d.Get("name").(string),
		AdminStateUp: &adminStateUp,
		Shared:       &shared,
	}

	log.Printf("[DEBUG] Create network: %#v", opts)

	network, err := networks.Create(networkClient, opts).Extract()
	if err != nil {
		return err
	}

	d.SetId(network.ID)

	return nil
}

func resourceNetworkDelete(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] Destroy network: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	res := networks.Delete(networkClient, d.Id())
	return res.Err
}

func resourceNetworkRead(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] Retrieve information about network: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	network, err := networks.Get(networkClient, d.Id()).Extract()
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

	d.Set("name", network.Name)
	d.Set("shared", network.Shared)
	d.Set("admin_state_up", network.AdminStateUp)
	return nil
}

func resourceNetworkUpdate(d *schema.ResourceData, meta interface{}) error {

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	opts := networks.UpdateOpts{}

	if d.HasChange("name") {
		opts.Name = d.Get("name").(string)
	}

	if d.HasChange("shared") {
		shared := d.Get("shared").(bool)
		opts.Shared = &shared
	}

	if d.HasChange("admin_state_up") {
		adminStateUp := d.Get("admin_state_up").(bool)
		opts.AdminStateUp = &adminStateUp
	}

	log.Printf("[DEBUG] Updating network: %#v", opts)

	_, err = networks.Update(networkClient, d.Id(), opts).Extract()
	return err
}
