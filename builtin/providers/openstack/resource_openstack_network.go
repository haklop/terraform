package openstack

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
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
		},
	}
}

func resourceNetworkCreate(d *schema.ResourceData, meta interface{}) error {

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	opts := networks.CreateOpts{
		Name:         d.Get("name").(string),
		AdminStateUp: networks.Up,
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

	network, err := networks.Get(networkClient, "id").Extract()
	if err != nil {
		return err
	}
	// TODO Handle 404 as before
	// n, err := networksApi.GetNetwork(d.Id())
	// if err != nil {
	// 	httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
	// 	if !ok {
	// 		return err
	// 	}
	//
	// 	if httpError.Actual == 404 {
	// 		d.SetId("")
	// 		return nil
	// 	}
	//
	// 	return err
	// }

	d.Set("name", network.Name)
	return nil
}

func resourceNetworkUpdate(d *schema.ResourceData, meta interface{}) error {

	d.Partial(true)

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}


	if d.HasChange("name") {
		opts := networks.UpdateOpts{Name: d.Get("name").(string)}
		_, err := networks.Update(networkClient, d.Id(), opts).Extract()
		if err != nil {
			return err
		}

		d.SetPartial("name")
	}

	d.Partial(false)

	return nil
}
