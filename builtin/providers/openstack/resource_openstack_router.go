package openstack

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
)

func resourceRouter() *schema.Resource {
	return &schema.Resource{
		Create: resourceRouterCreate,
		Read:   resourceRouterRead,
		Update: resourceRouterUpdate,
		Delete: resourceRouterDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"external_network_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRouterCreate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)

	networkClient, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	gwInfo := routers.GatewayInfo{
		NetworkID: d.Get("name").(string),
	}

	opts := routers.CreateOpts{
		Name:         d.Get("external_network_id").(string),
		AdminStateUp: networks.Up,
		TenantID:     p.TenantId,
		GatewayInfo:  &gwInfo,
	}

	router, err := routers.Create(networkClient, opts).Extract()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] New router created: %#v", router)

	d.SetId(router.ID)

	// TODO wait for router status

	return nil
}

func resourceRouterDelete(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	networkClient, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Delete router: %s", d.Id())

	deleteResult := routers.Delete(networkClient, d.Id())
	return deleteResult.Err
}

func resourceRouterRead(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	networkClient, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	router, err := routers.Get(networkClient, d.Id()).Extract()
	if err != nil {
		return err
	}

	// TODO Handler 404
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

	d.Set("name", router.Name)
	d.Set("external_network_id", router.GatewayInfo.NetworkID)

	return nil
}

func resourceRouterUpdate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	networkClient, err := p.getNetworkClient()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Updating router: %#v", d.Id())

	if d.HasChange("name") || d.HasChange("external_network_id") {

		gwInfo := routers.GatewayInfo{
			NetworkID: d.Get("external_network_id").(string),
		}

		opts := routers.UpdateOpts{
			Name:         d.Get("name").(string),
			AdminStateUp: networks.Up,
			GatewayInfo:  &gwInfo,
		}

		_, err := routers.Update(networkClient, d.Id(), opts).Extract()
		return err
	}

	return nil
}
