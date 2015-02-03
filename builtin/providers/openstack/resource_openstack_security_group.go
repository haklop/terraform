package openstack

import (
	"fmt"

	"github.com/ggiamarchi/gophercloud/openstack/compute/v2/extensions/secgroups"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
)

func resourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceSecurityGroupCreate,
		Read:   resourceSecurityGroupRead,
		Delete: resourceSecurityGroupDelete,
		Update: resourceSecurityGroupUpdate,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"rule": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"to_port": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"ip_protocol": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"cidr": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"from_group_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	opts := secgroups.CreateOpts{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	newSecurityGroup, err := secgroups.Create(computeClient, opts).Extract()
	if err != nil {
		return err
	}

	d.SetId(newSecurityGroup.ID)

	rulesCount := d.Get("rule.#").(int)
	for i := 0; i < rulesCount; i++ {
		prefix := fmt.Sprintf("rule.%d", i)

		var ruleOpts secgroups.CreateRuleOpts
		ruleOpts.ParentGroupID = newSecurityGroup.ID
		ruleOpts.FromPort = d.Get(prefix + ".from_port").(int)
		ruleOpts.ToPort = d.Get(prefix + ".to_port").(int)
		ruleOpts.IPProtocol = d.Get(prefix + ".ip_protocol").(string)
		ruleOpts.CIDR = d.Get(prefix + ".cidr").(string)
		ruleOpts.FromGroupID = d.Get(prefix + ".from_group_id").(string)
		_, err = secgroups.CreateRule(computeClient, ruleOpts).Extract()
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	res := secgroups.Delete(computeClient, d.Id())
	return res.Err
}

func resourceSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	d.Partial(true)

	if d.HasChange("name") || d.HasChange("description") {
		opts := secgroups.UpdateOpts{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}
		_, err := secgroups.Update(computeClient, d.Id(), opts).Extract()
		if err != nil {
			return err
		}

		d.SetPartial("name")
		d.SetPartial("description")
	}

	d.Partial(false)

	return nil
}

func resourceSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	_, err = secgroups.Get(computeClient, d.Id()).Extract()
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
