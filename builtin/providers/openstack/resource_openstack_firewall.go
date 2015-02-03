package openstack

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
	"github.com/ggiamarchi/gophercloud"
	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/extensions/fwaas/firewalls"
)

func resourceFirewall() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallCreate,
		Read:   resourceFirewallRead,
		Update:   resourceFirewallUpdate,
		Delete: resourceFirewallDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"admin_state_up": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceFirewallCreate(d *schema.ResourceData, meta interface{}) error {

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	firewallConfiguration := firewalls.CreateOpts{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		PolicyID:     d.Get("policy_id").(string),
		AdminStateUp: d.Get("admin_state_up").(bool),
	}

	log.Printf("[DEBUG] Create firewall: %#v", firewallConfiguration)

	firewall, err := firewalls.Create(networkClient, firewallConfiguration).Extract()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Firewall created: %#v", firewall)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"PENDING_CREATE"},
		Target:     "ACTIVE",
		Refresh:    WaitForFirewallActive(networkClient, firewall.ID),
		Timeout:    30 * time.Second,
		Delay:      0,
		MinTimeout: 2 * time.Second,
	}

	d.SetId(firewall.ID)

	d.Set("name", firewall.Name)
	d.Set("description", firewall.Description)
	d.Set("policy_id", firewall.PolicyID)
	d.Set("admin_state_up", firewall.AdminStateUp)

	_, err = stateConf.WaitForState()

	return nil
}

func resourceFirewallRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Retrieve information about firewall: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	firewall, err := firewalls.Get(networkClient, d.Id()).Extract()
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

	d.Set("name", firewall.Name)
	d.Set("description", firewall.Description)
	d.Set("policy_id", firewall.PolicyID)
	d.Set("admin_state_up", firewall.AdminStateUp)

	return nil
}

func resourceFirewallUpdate(d *schema.ResourceData, meta interface{}) error {

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	opts := firewalls.UpdateOpts{}

	if d.HasChange("name") {
		opts.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		opts.Description = d.Get("description").(string)
	}

	if d.HasChange("policy_id") {
		opts.PolicyID = d.Get("policy_id").(string)
	}

	log.Printf("[DEBUG] Updating firewall with id %s: %#v", d.Id(), opts)

	if err := firewalls.Update(networkClient, d.Id(), opts).Err; err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"PENDING_CREATE"},
		Target:     "ACTIVE",
		Refresh:    WaitForFirewallActive(networkClient, d.Id()),
		Timeout:    30 * time.Second,
		Delay:      0,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForState()

	return err
}

func resourceFirewallDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Destroy firewall: %s", d.Id())

	c := meta.(*Config)
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}

	err = firewalls.Delete(networkClient, d.Id()).Err

	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     "DELETED",
		Refresh:    WaitForFirewallDeletion(networkClient, d.Id()),
		Timeout:    2 * time.Minute,
		Delay:      0,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForState()

	return err
}

func WaitForFirewallActive(networkClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		fw, err := firewalls.Get(networkClient, id).Extract()
		log.Printf("[DEBUG] Get firewall %s => %#v", id, fw)

		if err != nil {
			return nil, "", err
		}
		return fw, fw.Status, nil
	}
}

func WaitForFirewallDeletion(networkClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		fw, err := firewalls.Get(networkClient, id).Extract()
		log.Printf("[DEBUG] Get firewall %s => %#v", id, fw)

		if err != nil {
			httpStatus := err.(*perigee.UnexpectedResponseCodeError)
			log.Printf("[DEBUG] Get firewall %s status is %d", id, httpStatus.Actual)

			if httpStatus.Actual == 404 {
				log.Printf("[DEBUG] Firewall %s is actually deleted", id)
				return "", "DELETED", nil
			}
			return nil, "", errors.New(fmt.Sprintf("Unexpected status code %d", httpStatus.Actual))
		}

		log.Printf("[DEBUG] Firewall %s deletion is pending", id)
		return fw, "DELETING", nil
	}
}
