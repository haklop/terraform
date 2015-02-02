package openstack

import (
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/fwaas/rules"
)

func resourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallRuleCreate,
		Read:   resourceFirewallRuleRead,
		Update: resourceFirewallRuleUpdate,
		Delete: resourceFirewallRuleDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"action": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_version": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  4,
			},
			"source_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	ipVersion := d.Get("ip_version").(int)
	enabled := d.Get("enabled").(bool)

	ruleConfiguration := rules.CreateOpts{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Protocol:             d.Get("protocol").(string),
		Action:               d.Get("action").(string),
		IPVersion:            &ipVersion,
		SourceIPAddress:      d.Get("source_ip_address").(string),
		DestinationIPAddress: d.Get("destination_ip_address").(string),
		SourcePort:           d.Get("source_port").(string),
		DestinationPort:      d.Get("destination_port").(string),
		Enabled:              &enabled,
	}

	log.Printf("[DEBUG] Create firewall rule: %#v", ruleConfiguration)

	rule, err := rules.Create(networkClient, ruleConfiguration).Extract()

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Firewall rule with id %s : %#v", rule.ID, rule)

	d.SetId(rule.ID)

	time.Sleep(time.Second * 5)

	d.Set("name", rule.Name)
	d.Set("description", rule.Description)
	d.Set("protocol", rule.Protocol)
	d.Set("action", rule.Action)
	d.Set("ip_version", rule.IPVersion)
	d.Set("source_ip_address", rule.SourceIPAddress)
	d.Set("destination_ip_address", rule.DestinationIPAddress)
	d.Set("source_port", rule.SourcePort)
	d.Set("destination_port", rule.DestinationPort)
	d.Set("enabled", rule.Enabled)

	return nil
}

func resourceFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Retrieve information about firewall rule: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	rule, err := rules.Get(networkClient, d.Id()).Extract()

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

	d.Set("name", rule.Name)
	d.Set("description", rule.Description)
	d.Set("protocol", rule.Protocol)
	d.Set("action", rule.Action)
	d.Set("ip_version", rule.IPVersion)
	d.Set("source_ip_address", rule.SourceIPAddress)
	d.Set("destination_ip_address", rule.DestinationIPAddress)
	d.Set("source_port", rule.SourcePort)
	d.Set("destination_port", rule.DestinationPort)
	d.Set("enabled", rule.Enabled)

	return nil
}

func resourceFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	opts := rules.UpdateOpts{}

	if d.HasChange("name") {
		opts.Name = d.Get("name").(string)
	}
	if d.HasChange("description") {
		opts.Description = d.Get("description").(string)
	}
	if d.HasChange("protocol") {
		opts.Protocol = d.Get("protocol").(string)
	}
	if d.HasChange("action") {
		opts.Action = d.Get("action").(string)
	}
	if d.HasChange("ip_version") {
		ipVersion := d.Get("ip_version").(int)
		opts.IPVersion = &ipVersion
	}
	if d.HasChange("source_ip_address") {
		opts.SourceIPAddress = d.Get("source_ip_address").(string)
	}
	if d.HasChange("destination_ip_address") {
		opts.DestinationIPAddress = d.Get("destination_ip_address").(string)
	}
	if d.HasChange("source_port") {
		opts.SourcePort = d.Get("source_port").(string)
	}
	if d.HasChange("destination_port") {
		opts.DestinationPort = d.Get("destination_port").(string)
	}
	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		opts.Enabled = &enabled
	}

	log.Printf("[DEBUG] Updating firewall rules: %#v", opts)

	return rules.Update(networkClient, d.Id(), opts).Err
}

func resourceFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Destroy firewall rule: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}
	return rules.Delete(networkClient, d.Id()).Err
}
