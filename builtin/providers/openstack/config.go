package openstack

import (
	"log"

	"github.com/ggiamarchi/gophercloud"
	"github.com/ggiamarchi/gophercloud/openstack"
)

type Config struct {
	AuthUrl    string `mapstructure:"auth_url"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	TenantId   string `mapstructure:"tenant_id"`
	TenantName string `mapstructure:"tenant_name"`
	Region     string `mapstructure:"region"`

	Provider *gophercloud.ProviderClient
}

// Client() returns a new client for accessing openstack.
//
func (c *Config) NewClient() error {

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: c.AuthUrl,
		Username:         c.Username,
		Password:         c.Password,
		TenantName:       c.TenantName,
		TenantID:         c.TenantId,
		AllowReauth:      true,
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return err
	}
	c.Provider = provider

	log.Printf("[INFO] Openstack Client configured for user %s", c.Username)

	return nil
}

func (c *Config) getComputeClient() (*gophercloud.ServiceClient, error) {
	return openstack.NewComputeV2(c.Provider, gophercloud.EndpointOpts{
		Name:         "nova",
		Availability: gophercloud.AvailabilityPublic,
		Region:       c.Region,
	})
}

func (c *Config) getNetworkClient() (*gophercloud.ServiceClient, error) {
	return openstack.NewNetworkV2(c.Provider, gophercloud.EndpointOpts{
		Name:         "neutron",
		Availability: gophercloud.AvailabilityPublic,
		Region:       c.Region,
	})
}
