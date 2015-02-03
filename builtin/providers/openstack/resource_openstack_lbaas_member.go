package openstack

import (
	"fmt"
	"log"

	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/extensions/lbaas/members"
	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/ports"
	"github.com/ggiamarchi/gophercloud/pagination"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
)

func resourceLBaaSMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceLBaaSMemberCreate,
		Read:   resourceLBaaSMemberRead,
		Delete: resourceLBaaSMemberDelete,

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLBaaSMemberCreate(d *schema.ResourceData, meta interface{}) error {

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	opts := ports.ListOpts{
		DeviceID: d.Get("instance_id").(string),
	}
	pager := ports.List(networkClient, opts)

	var address string
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		extractedPorts, err := ports.ExtractPorts(page)
		if err != nil {
			return false, err
		}

		for _, port := range extractedPorts {
			for _, ips := range port.FixedIPs {
				address = ips.IPAddress
			}
		}

		return len(address) == 0, nil
	})

	if len(address) == 0 {
		return fmt.Errorf("Cannot find a subnet")
	}

	membersOpts := members.CreateOpts{
		Address:      address,
		ProtocolPort: d.Get("port").(int),
		PoolID:       d.Get("pool_id").(string),
	}

	log.Printf("[DEBUG] Add member %#v", membersOpts)
	result, err := members.Create(networkClient, membersOpts).Extract()
	if err != nil {
		return err
	}

	d.SetId(result.ID)

	return nil
}

func resourceLBaaSMemberDelete(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] Destroy member: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	res := members.Delete(networkClient, d.Id())
	return res.Err
}

func resourceLBaaSMemberRead(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] Retrieve member: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	_, err = members.Get(networkClient, d.Id()).Extract()
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
	return nil
}
