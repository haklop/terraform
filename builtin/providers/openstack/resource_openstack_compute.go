package openstack

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/ggiamarchi/gophercloud"
	"github.com/ggiamarchi/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/ggiamarchi/gophercloud/openstack/compute/v2/servers"
	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/networks"
	"github.com/ggiamarchi/gophercloud/openstack/networking/v2/ports"
	"github.com/ggiamarchi/gophercloud/pagination"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/racker/perigee"
)

func resourceCompute() *schema.Resource {
	return &schema.Resource{
		Create: resourceComputeCreate,
		Read:   resourceComputeRead,
		Update: resourceComputeUpdate,
		Delete: resourceComputeDelete,

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
				ForceNew: true,
			},

			"flavor_ref": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"availability_zone": &schema.Schema{
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
				Optional: true,
			},

			"user_data": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// just stash the hash for state & diff comparisons
				StateFunc: func(v interface{}) string {
					switch v.(type) {
					case string:
						hash := sha1.Sum([]byte(v.(string)))
						return hex.EncodeToString(hash[:])
					default:
						return ""
					}
				},
			},
		},
	}
}

func resourceComputeCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	computeClient, err := c.getComputeClient()
	if err != nil {
		return err
	}

	v := d.Get("security_groups").(*schema.Set)
	securityGroups := make([]string, v.Len())
	for _, v := range v.List() {
		securityGroups = append(securityGroups, v.(string))
	}

	v = d.Get("networks").(*schema.Set)
	networks := []servers.Network{}
	for _, v := range v.List() {
		if len(v.(string)) > 0 {
			networks = append(networks, servers.Network{
				UUID: v.(string),
			})
		}
	}

	opts := servers.CreateOpts{
		Name:             d.Get("name").(string),
		ImageRef:         d.Get("image_ref").(string),
		FlavorRef:        d.Get("flavor_ref").(string),
		SecurityGroups:   securityGroups,
		AvailabilityZone: d.Get("availability_zone").(string),
		Networks:         networks,
		UserData:         []byte(d.Get("user_data").(string)),
	}

	instance, err := servers.Create(computeClient, keypairs.CreateOptsExt{
		opts,
		d.Get("key_pair_name").(string),
	}).Extract()
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

	floatingip := d.Get("floating_ip").(string)
	pool := d.Get("floating_ip_pool").(string)
	log.Printf("[DEBUG] 0. Looking for an IP in the pool %s\n", pool)
	allFloatingIPs, err := getFloatingIPs(c)
	if err != nil {
		return err
	}
	if len(floatingip) > 0 {
		err = assignFloatingIP(c, extractFloatingIPFromIP(allFloatingIPs, floatingip), instance.ID)
		if err != nil {
			return err
		}
	} else if len(pool) > 0 {
		log.Printf("[DEBUG] Looking for an IP in the pool %s\n", pool)
		networkClient, err := c.getNetworkClient()
		if err != nil {
			return err
		}
		poolID, err := getNetworkID(c, pool)
		if err != nil {
			return err
		}
		opts := floatingips.ListOpts{
			FloatingNetworkID: poolID,
		}
		pager := floatingips.List(networkClient, opts)
		ipAssigned := false
		err = pager.EachPage(func(page pagination.Page) (bool, error) {
			floatingIPList, err := floatingips.ExtractFloatingIPs(page)
			if err != nil {
				return false, err
			}

			for _, f := range floatingIPList {
				if f.FloatingNetworkID == poolID && len(f.PortID) == 0 {
					floatingip = f.FloatingIP
					err = assignFloatingIP(c, &f, instance.ID)
					if err != nil {
						return false, err
					}
					ipAssigned = true
					return false, nil
				}
			}
			return true, nil
		})

		if !ipAssigned {
			allocatedIP, err := allocateAndAssignFloatingIP(c, poolID, instance.ID)
			if err != nil {
				return err
			}
			ipAssigned = true
			floatingip = allocatedIP.FloatingIP
		}
	}
	var sshIP string
	if len(floatingip) > 0 {
		d.Set("floating_ip", floatingip)
		sshIP = floatingip
	} else {
		sshIP, err = getIP(c, instance.ID)
		if err != nil {
			return err
		}
	}

	// Initialize the connection info
	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": sshIP,
	})

	return nil
}

func extractFloatingIPFromIP(ips []floatingips.FloatingIP, IP string) *floatingips.FloatingIP {
	for _, floatingIP := range ips {
		if floatingIP.FloatingIP == IP {
			return &floatingIP
		}
	}
	return nil
}

func getFloatingIPs(c *Config) ([]floatingips.FloatingIP, error) {
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return nil, err
	}
	pager := floatingips.List(networkClient, floatingips.ListOpts{})

	ips := []floatingips.FloatingIP{}
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		floatingipList, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return false, err
		}
		for _, f := range floatingipList {
			ips = append(ips, f)
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}
	return ips, nil
}

func assignFloatingIP(c *Config, floatingIP *floatingips.FloatingIP, instanceID string) error {
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return err
	}
	networkID, err := getFirstNetworkID(c, instanceID)
	if err != nil {
		return err
	}
	portID, err := getInstancePortID(c, instanceID, networkID)
	_, err = floatingips.Update(networkClient, floatingIP.ID, floatingips.UpdateOpts{
		PortID: portID,
	}).Extract()
	return err
}

func allocateAndAssignFloatingIP(c *Config, floatingNetworkID, instanceID string) (*floatingips.FloatingIP, error) {
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return nil, err
	}
	networkID, err := getFirstNetworkID(c, instanceID)
	if err != nil {
		return nil, err
	}
	portID, err := getInstancePortID(c, instanceID, networkID)

	return floatingips.Create(networkClient, floatingips.CreateOpts{
		FloatingNetworkID: floatingNetworkID,
		PortID:            portID,
	}).Extract()
}

func getInstancePortID(c *Config, instanceID, networkID string) (string, error) {
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return "", err
	}
	pager := ports.List(networkClient, ports.ListOpts{
		DeviceID:  instanceID,
		NetworkID: networkID,
	})

	var portID string
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		portList, err := ports.ExtractPorts(page)
		if err != nil {
			return false, err
		}
		for _, port := range portList {
			portID = port.ID
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return "", err
	}
	return portID, nil
}

func getFirstNetworkID(c *Config, instanceID string) (string, error) {
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return "", err
	}
	pager := networks.List(networkClient, networks.ListOpts{})

	var networkdID string
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		if len(networkList) > 0 {
			networkdID = networkList[0].ID
			return false, nil
		}
		return false, fmt.Errorf("No network found for the instance %s", instanceID)
	})
	if err != nil {
		return "", err
	}
	return networkdID, nil

}

func getNetworkID(c *Config, networkName string) (string, error) {
	networkClient, err := c.getNetworkClient()
	if err != nil {
		return "", err
	}

	opts := networks.ListOpts{Name: networkName}
	pager := networks.List(networkClient, opts)
	networkID := ""

	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		for _, n := range networkList {
			if n.Name == networkName {
				networkID = n.ID
				return false, nil
			}
		}

		return true, nil
	})

	return networkID, err
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
		opts := servers.UpdateOpts{
			Name: d.Get("name").(string),
		}
		_, err := servers.Update(computeClient, d.Id(), opts).Extract()
		if err != nil {
			return err
		}

		d.SetPartial("name")
	}

	if d.HasChange("flavor_ref") {
		opts := servers.ResizeOpts{
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

	// TODO check networks, seucrity groups and floating ip

	d.Set("name", server.Name)
	d.Set("flavor_ref", server.Flavor["id"])

	return nil
}

func getIP(c *Config, instanceID string) (string, error) {
	computeClient, err := c.getComputeClient()
	if err != nil {
		return "", err
	}

	server, err := servers.Get(computeClient, instanceID).Extract()
	if err != nil {
		return "", err
	}

	for _, networkAddresses := range server.Addresses {
		for _, element := range networkAddresses.([]interface{}) {
			address := element.(map[string]interface{})
			return address["addr"].(string), nil
		}
	}
	return "", fmt.Errorf("No IP found for the machine")
}

func WaitForServerState(client *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		s, err := servers.Get(client, id).Extract()

		if err != nil {
			httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
			if !ok {
				return nil, "", err
			}

			if httpError.Actual == 404 {
				return s, "DELETED", nil
			}
			return nil, "", err
		}
		return s, s.Status, nil
	}
}
