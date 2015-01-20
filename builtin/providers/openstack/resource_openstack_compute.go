package openstack

import (
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

func resourceCompute() *schema.Resource {
	return &schema.Resource{
		Create: resourceComputeCreate,
		Read:   resourceComputeRead,
		Update: resourceComputeUpdate,
		Delete: resourceComputeDelete,

		//TODO Handle UserData
		//TODO Handle Networks
		//TODO Handle Metadata
		//TODO Handle Personnality
		//TODO Handle ConfigDrive
		//TODO Handle AdminPass
		Schema: map[string]*schema.Schema{
			"image_ref": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"key_pair_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // TODO handle update
			},

			"flavor_ref": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"availabilty_zone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"networks": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true, // TODO handle update
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},

			"security_groups": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true, // TODO handle update
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},

			"floating_ip_pool": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"floating_ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			// "user_data": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// 	ForceNew: true,
			// 	// just stash the hash for state & diff comparisons
			// 	StateFunc: func(v interface{}) string {
			// 		switch v.(type) {
			// 		case string:
			// 			hash := sha1.Sum([]byte(v.(string)))
			// 			return hex.EncodeToString(hash[:])
			// 		default:
			// 			return ""
			// 		}
			// 	},
			// },
		},
	}
}

func resourceComputeCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	// v := d.Get("networks").(*schema.Set)
	// networks := make([]gophercloud.NetworkConfig, v.Len())
	// for i, v := range v.List() {
	// 	networks[i] = gophercloud.NetworkConfig{v.(string)}
	// }

	v := d.Get("security_groups").(*schema.Set)
	securityGroups := make([]string, v.Len())
	for _, v := range v.List() {
		securityGroups = append(securityGroups, v.(string))
	}

	v = d.Get("networks").(*schema.Set)
	networks := []servers.Network{}
	for _, v := range v.List() {
		if len(v.(string)) > 0 {
			networks = append(networks, servers.Network {
				UUID: v.(string),
			})
		}
	}

	opts := servers.CreateOpts {
		Name: d.Get("name").(string),
		ImageRef: d.Get("image_ref").(string),
		FlavorRef: d.Get("flavor_ref").(string),
		SecurityGroups: securityGroups,
		AvailabilityZone: d.Get("availabilty_zone").(string),
		Networks: networks,
	}

	instance, err := servers.Create(computeClient, opts).Extract()
	if err != nil {
		return err
	}

	d.SetId(instance.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     "ACTIVE",
		Refresh:    WaitForServerState(computeClient, d.Id()),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()

	if err != nil {
		return err
	}

	// pool := d.Get("floating_ip_pool").(string)
	// if len(pool) > 0 {
	// 	var newIp gophercloud.FloatingIp
	// 	hasFloatingIps := false
	//
	// 	floatingIps, err := serversApi.ListFloatingIps()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	for _, element := range floatingIps {
	// 		// use first floating ip available on the pool
	// 		if element.Pool == pool && element.InstanceId == "" {
	// 			newIp = element
	// 			hasFloatingIps = true
	// 		}
	// 	}
	//
	// 	// if there is no available floating ips, try to create a new one
	// 	if !hasFloatingIps {
	// 		newIp, err = serversApi.CreateFloatingIp(pool)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	//
	// 	err = serversApi.AssociateFloatingIp(newServer.Id, newIp)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	d.Set("floating_ip", newIp.Ip)
	//
	// 	// Initialize the connection info
	// 	d.SetConnInfo(map[string]string{
	// 		"type": "ssh",
	// 		"host": newIp.Ip,
	// 	})
	// }

	return nil
}

func resourceComputeDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	res := servers.Delete(computeClient, d.Id())
	if res.Err != nil {
		return res.Err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE", "ERROR"},
		Target:     "DELETED",
		Refresh:    WaitForServerState(computeClient, d.Id()),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()

	return err
}

func resourceComputeUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	d.Partial(true)

	if d.HasChange("name") {
		opts:= servers.UpdateOpts {
			Name: d.Get("name").(string),
		}
		_, err := servers.Update(computeClient, d.Id(), opts).Extract()
		if err != nil {
			return err
		}

		d.SetPartial("name")
	}

	if d.HasChange("flavor_ref") {
		opts := servers.ResizeOpts {
			FlavorRef: d.Get("flavor_ref").(string),
		}
		res := servers.Resize(computeClient, d.Id(), opts)
		if res.Err != nil {
			return res.Err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"ACTIVE", "RESIZE"},
			Target:     "VERIFY_RESIZE",
			Refresh:    WaitForServerState(computeClient, d.Id()),
			Timeout:    30 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()

		if err != nil {
			return err
		}

		res = servers.ConfirmResize(computeClient, d.Id())
		if res.Err != nil {
			return res.Err
		}

		d.SetPartial("flavor_ref")
	}

	d.Partial(false)

	return nil
}

func resourceComputeRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	server, err := servers.Get(computeClient, d.Id()).Extract()
	// TODO Handle error and not found
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

	// TODO check networks, seucrity groups and floating ip

	d.Set("name", server.Name)
	d.Set("flavor_ref", server.Flavor["id"])

	return nil
}

func WaitForServerState(client *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		s, err := servers.Get(client, id).Extract()
		//TODO Handle err and 404
		if err != nil {
			return nil, "", err
		}
		// if err != nil {
		// 	httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		// 	if !ok {
		// 		return nil, "", err
		// 	}
		//
		// 	if httpError.Actual == 404 {
		// 		return s, "DELETED", nil
		// 	}
		//
		// 	return nil, "", err
		// }

		return s, s.Status, nil

	}
}
