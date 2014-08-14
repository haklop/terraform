package openstack

import (
	"github.com/hashicorp/terraform/helper/config"
	"github.com/hashicorp/terraform/helper/resource"
)

// resourceMap is the mapping of resources we support to their basic
// operations. This makes it easy to implement new resource types.
var resourceMap *resource.Map

func init() {
	resourceMap = &resource.Map{
		Mapping: map[string]resource.Resource{
			"openstack_compute": resource.Resource{
				ConfigValidator: &config.Validator{
					Required: []string{
						"image_ref",
						"flavor_ref",
					},
					Optional: []string{
						"floating_ip_pool",
						"name",
						"key_pair_name",
						"networks.*",
						"security_groups.*",
					},
				},
				Create:  resource_openstack_compute_create,
				Destroy: resource_openstack_compute_destroy,
				Diff:    resource_openstack_compute_diff,
				Update:  resource_openstack_compute_update,
				Refresh: resource_openstack_compute_refresh,
			},

			"openstack_network": resource.Resource{
				ConfigValidator: &config.Validator{
					Required: []string{
						"name",
					},
					Optional: []string{},
				},
				Create:  resource_openstack_network_create,
				Destroy: resource_openstack_network_destroy,
				Diff:    resource_openstack_network_diff,
				Update:  resource_openstack_network_update,
				Refresh: resource_openstack_network_refresh,
			},

			"openstack_subnet": resource.Resource{
				ConfigValidator: &config.Validator{
					Required: []string{
						"cidr",
						"ip_version",
						"network_id",
					},
					Optional: []string{
						"name",
						"enable_dhcp",
					},
				},
				Create:  resource_openstack_subnet_create,
				Destroy: resource_openstack_subnet_destroy,
				Diff:    resource_openstack_subnet_diff,
				Update:  resource_openstack_subnet_update,
				Refresh: resource_openstack_subnet_refresh,
			},

			"openstack_security_group": resource.Resource{
				ConfigValidator: &config.Validator{
					Required: []string{
						"name",
						"rule.*.direction",
						"rule.*.remote_ip_prefix",
					},
					Optional: []string{
						"description",
						"rule.*.port_range_min",
						"rule.*.port_range_max",
						"rule.*.protocol",
					},
				},
				Create:  resource_openstack_security_group_create,
				Destroy: resource_openstack_security_group_destroy,
				Diff:    resource_openstack_security_group_diff,
				Refresh: resource_openstack_security_group_refresh,
			},

			"openstack_router": resource.Resource{
				ConfigValidator: &config.Validator{
					Required: []string{
						"name",
					},
					Optional: []string{
						"external_gateway",
						"external_id",
						"subnets.*",
					},
				},
				Create:  resource_openstack_router_create,
				Destroy: resource_openstack_router_destroy,
				Diff:    resource_openstack_router_diff,
				Refresh: resource_openstack_router_refresh,
			},

			"openstack_lbaas": resource.Resource{
				ConfigValidator: &config.Validator{
					Required: []string{
						"name",
						"subnet_id",
						"protocol",
						"lb_method",
						"member.*.port",
						"member.*.instance_id",
					},
					Optional: []string{
						"description",
					},
				},
				Create:  resource_openstack_lbaas_create,
				Destroy: resource_openstack_lbaas_destroy,
				Diff:    resource_openstack_lbaas_diff,
				Update:  resource_openstack_lbaas_update,
				Refresh: resource_openstack_lbaas_refresh,
			},
		},
	}
}