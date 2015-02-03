package openstack

import (
	"log"
	"time"

	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/extensions/fwaas/policies"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
)

func resourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallPolicyCreate,
		Read:   resourceFirewallPolicyRead,
		Update: resourceFirewallPolicyUpdate,
		Delete: resourceFirewallPolicyDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"audited": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"shared": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"rules": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},
		},
	}
}

func resourceFirewallPolicyCreate(d *schema.ResourceData, meta interface{}) error {

	time.Sleep(time.Second * 5)

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	v := d.Get("rules").(*schema.Set)

	log.Printf("[DEBUG] Rules found : %#v", v)
	log.Printf("[DEBUG] Rules count : %i", v.Len())

	rules := make([]string, v.Len())
	for i, v := range v.List() {
		rules[i] = v.(string)
	}

	opts := policies.CreateOpts{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		//		Audited:     d.Get("audited").(bool),
		//		Shared:      d.Get("shared").(bool),
		Rules: rules,
	}

	log.Printf("[DEBUG] Create firewall policy: %#v", opts)

	policy, err := policies.Create(networkClient, opts).Extract()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Firewall policy craeted: %#v", policy)

	d.SetId(policy.ID)

	return nil
}

func resourceFirewallPolicyRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Retrieve information about firewall policy: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	policy, err := policies.Get(networkClient, d.Id()).Extract()

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

	d.Set("name", policy.Name)

	return nil
}

func resourceFirewallPolicyUpdate(d *schema.ResourceData, meta interface{}) error {

	p := meta.(*Config)
	client, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	opts := policies.UpdateOpts{}

	if d.HasChange("name") {
		opts.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		opts.Description = d.Get("description").(string)
	}

	if d.HasChange("rules") {
		v := d.Get("rules").(*schema.Set)

		log.Printf("[DEBUG] Rules found : %#v", v)
		log.Printf("[DEBUG] Rules count : %i", v.Len())

		rules := make([]string, v.Len())
		for i, v := range v.List() {
			rules[i] = v.(string)
		}
		opts.Rules = rules
	}

	log.Printf("[DEBUG] Updating firewall policy with id %s: %#v", d.Id(), opts)

	return policies.Update(client, d.Id(), opts).Err
}

func resourceFirewallPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Destroy firewall policy: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	for i := 0; i < 15; i++ {

		err = policies.Delete(networkClient, d.Id()).Err
		if err == nil {
			break
		}

		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok || httpError.Actual != 409 {
			return err
		}

		// This error usualy means that the policy is attached
		// to a firewall. At this point, the firewall is probably
		// being delete. So, we retry a few times.

		time.Sleep(time.Second * 2)
	}

	return err
}
