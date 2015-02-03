package openstack

import (
	"log"

	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/extensions/lbaas/vips"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
)

func resourceLBaaSVips() *schema.Resource {
	return &schema.Resource{
		Create: resourceLBaaSVipsCreate,
		Read:   resourceLBaaSVipsRead,
		Update: resourceLBaaSVipsUpdate,
		Delete: resourceLBaaSVipsDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol_port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"conn_limit": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"persistance": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceLBaaSVipsCreate(d *schema.ResourceData, meta interface{}) error {

	c := meta.(*Config)
	client, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	opts := vips.CreateOpts{
		Name:         d.Get("name").(string),
		SubnetID:     d.Get("subnet_id").(string),
		Protocol:     d.Get("protocol").(string),
		ProtocolPort: d.Get("protocol_port").(int),
		PoolID:       d.Get("pool_id").(string),
	}

	address := d.Get("ip_address").(string)
	if len(address) > 0 {
		opts.Address = address
	}

	description := d.Get("description").(string)
	if len(description) > 0 {
		opts.Description = description
	}

	connLimit := d.Get("conn_limit").(int)
	opts.ConnLimit = &connLimit

	// TODO persistance

	log.Printf("[DEBUG] Create vip %#v", opts)
	result, err := vips.Create(client, opts).Extract()
	if err != nil {
		return err
	}

	d.SetId(result.ID)

	return nil
}

func resourceLBaaSVipsUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] Update vip: %s", d.Id())

	opts := vips.UpdateOpts{}

	if d.HasChange("name") {
		opts.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		opts.Description = d.Get("description").(string)
	}

	if d.HasChange("conn_limit") {
		connLimit := d.Get("conn_limit").(int)
		opts.ConnLimit = &connLimit
	}

	// TODO persistence

	c := meta.(*Config)
	client, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	_, err = vips.Update(client, d.Id(), opts).Extract()
	return err
}

func resourceLBaaSVipsDelete(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] Destroy vips: %s", d.Id())

	c := meta.(*Config)
	client, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	res := vips.Delete(client, d.Id())
	return res.Err
}

func resourceLBaaSVipsRead(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] Retrieve vips: %s", d.Id())

	c := meta.(*Config)
	client, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	vip, err := vips.Get(client, d.Id()).Extract()
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

	d.Set("name", vip.Name)
	d.Set("description", vip.Description)
	d.Set("conn_limit", vip.ConnLimit)

	return nil
}
