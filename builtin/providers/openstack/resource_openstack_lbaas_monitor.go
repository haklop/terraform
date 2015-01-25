package openstack

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas/monitors"
)

func resourceLBaaSMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceLBaaSMonitorCreate,
		Read:   resourceLBaaSMonitorRead,
		Update: resourceLBaaSMonitorUpdate,
		Delete: resourceLBaaSMonitorDelete,

		Schema: map[string]*schema.Schema{
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"delay": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"max_retries": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"expected_codes": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"http_method": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"url_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"admin_state_up": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceLBaaSMonitorCreate(d *schema.ResourceData, meta interface{}) error {

	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	adminStateUp := d.Get("admin_state_up").(bool)
	opts := monitors.CreateOpts{
		Type:         d.Get("type").(string),
		Delay:        d.Get("delay").(int),
		Timeout:      d.Get("timeout").(int),
		MaxRetries:   d.Get("max_retries").(int),
		AdminStateUp: &adminStateUp,
	}

	expectedCodes := d.Get("expected_codes").(string)
	if len(expectedCodes) > 0 {
		opts.ExpectedCodes = expectedCodes
	}

	httpMethod := d.Get("http_method").(string)
	if len(httpMethod) > 0 {
		opts.HTTPMethod = httpMethod
	}

	urlPath := d.Get("url_path").(string)
	if len(urlPath) > 0 {
		opts.URLPath = urlPath
	}

	monitor, err := monitors.Create(client, opts).Extract()
	if err != nil {
		return err
	}

	d.SetId(monitor.ID)
	readMonitor(monitor, d)

	return nil

}

func resourceLBaaSMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	res := monitors.Delete(client, d.Id())
	return res.Err

}

func resourceLBaaSMonitorRead(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	monitor, err := monitors.Get(client, d.Id()).Extract()
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

	readMonitor(monitor, d)

	return nil

}

func resourceLBaaSMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	opts := monitors.UpdateOpts{
		Delay:      d.Get("delay").(int),
		Timeout:    d.Get("timeout").(int),
		MaxRetries: d.Get("max_retries").(int),
	}

	expectedCodes := d.Get("expected_codes").(string)
	if len(expectedCodes) > 0 {
		opts.ExpectedCodes = expectedCodes
	}

	httpMethod := d.Get("http_method").(string)
	if len(httpMethod) > 0 {
		opts.HTTPMethod = httpMethod
	}

	urlPath := d.Get("url_path").(string)
	if len(urlPath) > 0 {
		opts.URLPath = urlPath
	}

	if d.HasChange("admin_state_up") {
		adminStateUp := d.Get("admin_state_up").(bool)
		opts.AdminStateUp = &adminStateUp
	}

	monitor, err := monitors.Update(client, d.Id(), opts).Extract()
	if err != nil {
		return err
	}

	readMonitor(monitor, d)
	return nil

}

func readMonitor(monitor *monitors.Monitor, d *schema.ResourceData) {
	d.Set("type", monitor.Type)
	d.Set("delay", monitor.Delay)
	d.Set("timeout", monitor.Timeout)
	d.Set("max_retries", monitor.MaxRetries)
	d.Set("admin_state_up", monitor.AdminStateUp)

	if monitor.Type == monitors.TypeHTTP || monitor.Type == monitors.TypeHTTPS {
		d.Set("expected_codes", monitor.ExpectedCodes)
		d.Set("http_method", monitor.HTTPMethod)
		d.Set("url_path", monitor.URLPath)
	} else {
		d.Set("expected_codes", "")
		d.Set("http_method", "")
		d.Set("url_path", "")
	}

}

func flattenMonitor(list []monitors.Monitor) []string {
	result := make([]string, 0, len(list))
	for _, i := range list {
		result = append(result, i.ID)
	}
	return result
}
