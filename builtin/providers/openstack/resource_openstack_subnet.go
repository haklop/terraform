package openstack

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
)

func resourceSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceSubnetDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"cidr": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_version": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"enable_dhcp": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"gateway_ip": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"allocation_pool": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"end": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	enableDHCP := d.Get("enable_dhcp").(bool)

	opts := subnets.CreateOpts{
		NetworkID:  d.Get("network_id").(string),
		Name:       d.Get("name").(string),
		CIDR:       d.Get("cidr").(string),
		EnableDHCP: &enableDHCP,
		IPVersion:  d.Get("ip_version").(int),
	}

	gatewayIP := d.Get("gateway_ip").(string)
	if len(gatewayIP) > 0 {
		opts.GatewayIP = gatewayIP
	}

	poolsCount := d.Get("allocation_pool.#").(int)
	if poolsCount > 0 {
		pools := make([]subnets.AllocationPool, 0, poolsCount)
		for i := 0; i < poolsCount; i++ {
			prefix := fmt.Sprintf("allocation_pool.%d", i)

			var pool subnets.AllocationPool
			pool.Start = d.Get(prefix + ".start").(string)
			pool.End = d.Get(prefix + ".end").(string)

			pools = append(pools, pool)
		}

		opts.AllocationPools = pools
	}

	log.Printf("[DEBUG] Creating subnet: %#v", opts)

	createdSubnet, err := subnets.Create(client, opts).Extract()
	if err != nil {
		return err
	}

	d.SetId(createdSubnet.ID)
	readSubnet(createdSubnet, d)

	return nil
}

func resourceSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Destroy subnet: %s", d.Id())

	res := subnets.Delete(client, d.Id())
	return res.Err
}

func resourceSubnetRead(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Retrieve information about subnet: %s", d.Id())

	subnet, err := subnets.Get(client, d.Id()).Extract()
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

	readSubnet(subnet, d)

	return nil
}

func resourceSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	opts := subnets.UpdateOpts{}

	if d.HasChange("name") {
		opts.Name = d.Get("name").(string)
	}

	if d.HasChange("enable_dhcp") {
		enableDHCP := d.Get("enable_dhcp").(bool)
		opts.EnableDHCP = &enableDHCP
	}

	if d.HasChange("gateway_ip") {
		opts.GatewayIP = d.Get("gateway_ip").(string)
	}

	log.Printf("[DEBUG] Updating subnet: %#v", opts)

	_, err = subnets.Update(client, d.Id(), opts).Extract()
	return err
}

func readSubnet(subnet *subnets.Subnet, d *schema.ResourceData) {
	d.Set("name", subnet.Name)
	d.Set("cidr", subnet.CIDR)
	d.Set("enable_dhcp", subnet.EnableDHCP)
	d.Set("ip_version", subnet.IPVersion)
	d.Set("network_id", subnet.NetworkID)
	d.Set("gateway_ip", subnet.GatewayIP)

	d.Set("allocation_pool.#", len(subnet.AllocationPools))
	for i, pool := range subnet.AllocationPools {
		prefix := fmt.Sprintf("allocation_pool.%d", i)
		d.Set(prefix+".start", pool.Start)
		d.Set(prefix+".end", pool.End)
	}
}
