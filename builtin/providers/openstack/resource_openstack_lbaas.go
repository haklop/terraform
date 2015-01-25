package openstack

import (
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas/pools"
)

func resourceLBaaS() *schema.Resource {
	return &schema.Resource{
		Create: resourceLBaaSCreate,
		Read:   resourceLBaaSRead,
		Update: resourceLBaaSUpdate,
		Delete: resourceLBaaSDelete,

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
			"lb_method": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"health_monitors": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},
		},
	}
}

func resourceLBaaSCreate(d *schema.ResourceData, meta interface{}) error {

	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	opts := pools.CreateOpts{
		Name:     d.Get("name").(string),
		SubnetID: d.Get("subnet_id").(string),
		Protocol: d.Get("protocol").(string),
		LBMethod: d.Get("lb_method").(string),
	}

	pool, err := pools.Create(client, opts).Extract()
	if err != nil {
		return err
	}

	d.SetId(pool.ID)

	healthMonitor := d.Get("health_monitors").(*schema.Set)
	monitors := expandStringList(healthMonitor.List())
	for _, id := range monitors {
		_, err = pools.AssociateMonitor(client, d.Id(), id).Extract()
		if err != nil {
			return err
		}
	}

	d.Set("health_monitors", monitors)

	return nil

}

func resourceLBaaSDelete(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	healthMonitor := d.Get("health_monitors").(*schema.Set)
	monitors := expandStringList(healthMonitor.List())
	for _, id := range monitors {
		_, err = pools.AssociateMonitor(client, d.Id(), id).Extract()
		if err != nil {
			return err
		}
	}

	result := pools.Delete(client, d.Id())
	return result.Err
}

func resourceLBaaSRead(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	pool, err := pools.Get(client, d.Id()).Extract()
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

	readPool(pool, d)
	return nil

}

func resourceLBaaSUpdate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	d.Partial(true)
	if d.HasChange("name") || d.HasChange("lb_method") {
		opts := pools.UpdateOpts{}

		if d.HasChange("name") {
			opts.Name = d.Get("name").(string)
		}

		if d.HasChange("lb_method") {
			opts.LBMethod = d.Get("lb_method").(string)
		}

		_, err = pools.Update(client, d.Id(), opts).Extract()
		if err != nil {
			return err
		}

		d.SetPartial("name")
		d.SetPartial("lb_method")
	}

	if d.HasChange("health_monitors") {
		o, n := d.GetChange("health_monitors")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := expandStringList(os.Difference(ns).List())
		add := expandStringList(ns.Difference(os).List())

		for _, id := range remove {
			_, err = pools.DisassociateMonitor(client, d.Id(), id).Extract()
			if err != nil {
				return err
			}
		}

		for _, id := range add {
			_, err = pools.AssociateMonitor(client, d.Id(), id).Extract()
			if err != nil {
				return err
			}
		}

		d.SetPartial("health_monitors")
	}

	d.Partial(false)
	return nil

}

func readPool(pool *pools.Pool, d *schema.ResourceData) {
	d.Set("name", pool.Name)
	d.Set("subnet_id", pool.SubnetID)
	d.Set("protocol", pool.Protocol)
	d.Set("lb_method", pool.LBMethod)
	d.Set("health_monitors", pool.MonitorIDs)
}

// Takes the result of flatmap.Expand for an array of strings
// and returns a []string
func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, v.(string))
	}
	return vs
}
