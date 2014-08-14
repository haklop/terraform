package openstack

import (
	"github.com/haklop/gophercloud-extensions/network"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/hashicorp/terraform/helper/diff"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
	"log"
	"strconv"
)

func resource_openstack_lbaas_create(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {

	p := meta.(*ResourceProvider)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return nil, err
	}

	// Merge the diff into the state so that we have all the attributes
	// properly.
	rs := s.MergeDiff(d)

	pool, err := networksApi.CreatePool(network.NewPool{
		Name:        rs.Attributes["name"],
		SubnetId:    rs.Attributes["subnet_id"],
		LoadMethod:  rs.Attributes["lb_method"],
		Protocol:    rs.Attributes["protocol"],
		Description: rs.Attributes["description"],
	})
	if err != nil {
		return nil, err
	}

	rs.ID = pool.Id

	v, ok := flatmap.Expand(rs.Attributes, "member").([]interface{})
	if ok {
		members, err := expandMembers(v)
		if err != nil {
			return rs, err
		}

		for _, member := range members {

			// TODO order ports
			ports, err := networksApi.GetPorts()
			if err != nil {
				return nil, err
			}

			var address string
			for _, port := range ports {
				if port.DeviceId == member.InstanceId {
					for _, ips := range port.FixedIps {
						// ff possible, select a port on pool subnet
						if ips.SubnetId == rs.Attributes["subnet_id"] || address == "" {
							address = ips.IpAddress
						}
					}
				}
			}

			newMember := network.NewMember{
				ProtocolPort: member.ProtocolPort,
				PoolId:       rs.ID,
				AdminStateUp: true,
				Address:      address,
			}

			result, err := networksApi.CreateMember(newMember)
			if err != nil {
				return nil, err
			}
			member.MemberId = result.Id

		}

	}

	return rs, err
}

func resource_openstack_lbaas_destroy(
	s *terraform.ResourceState,
	meta interface{}) error {

	p := meta.(*ResourceProvider)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return err
	}

	// TODO destroy member

	err = networksApi.DeletePool(s.ID)

	return err
}

func resource_openstack_lbaas_refresh(
	s *terraform.ResourceState,
	meta interface{}) (*terraform.ResourceState, error) {

	log.Printf("[DEBUG] Retrieve information about pool: %s", s.ID)

	p := meta.(*ResourceProvider)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return nil, err
	}

	pool, err := networksApi.GetPool(s.ID)
	if err != nil {
		httpError, ok := err.(*perigee.UnexpectedResponseCodeError)
		if !ok {
			return nil, err
		}

		if httpError.Actual == 404 {
			return nil, nil
		}

		return nil, err
	}

	s.Attributes["name"] = pool.Name
	s.Attributes["description"] = pool.Description
	s.Attributes["lb_method"] = pool.LoadMethod

	// TODO compare pool.Members with Id on s.Extra['members'] ?

	return s, nil
}

func resource_openstack_lbaas_diff(
	s *terraform.ResourceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.ResourceDiff, error) {

	b := &diff.ResourceBuilder{
		Attrs: map[string]diff.AttrType{
			"name":        diff.AttrTypeUpdate,
			"subnet_id":   diff.AttrTypeCreate,
			"protocol":    diff.AttrTypeCreate,
			"lb_method":   diff.AttrTypeUpdate,
			"description": diff.AttrTypeUpdate,
			"member":      diff.AttrTypeUpdate,
		},

		ComputedAttrs: []string{
			"id",
		},
	}

	return b.Diff(s, c)
}

func resource_openstack_lbaas_update(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {

	p := meta.(*ResourceProvider)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return nil, err
	}

	// Merge the diff into the state so that we have all the attributes
	// properly.
	rs := s.MergeDiff(d)

	updatedPool := network.Pool{
		Id: rs.ID,
	}

	if attr, ok := d.Attributes["name"]; ok {
		updatedPool.Name = attr.New
		rs.Attributes["name"] = attr.New
	}

	if attr, ok := d.Attributes["lb_method"]; ok {
		updatedPool.LoadMethod = attr.New
		rs.Attributes["lb_method"] = attr.New
	}

	if attr, ok := d.Attributes["description"]; ok {
		updatedPool.Description = attr.New
		rs.Attributes["description"] = attr.New
	}

	_, err = networksApi.UpdatePool(updatedPool)

	// TODO update members

	return rs, err
}

func expandMembers(configured []interface{}) ([]poolMember, error) {
	members := make([]poolMember, 0, len(configured))

	for _, member := range configured {
		raw := member.(map[string]interface{})

		newMember := poolMember{}

		if attr, ok := raw["port"].(string); ok {
			port, err := strconv.Atoi(attr)
			if err != nil {
				return nil, err
			}
			newMember.ProtocolPort = port
		}

		if attr, ok := raw["instance_id"].(string); ok {
			newMember.InstanceId = attr
		}

		members = append(members, newMember)
	}

	return members, nil
}

type poolMember struct {
	ProtocolPort int
	InstanceId   string
	MemberId     string
}
