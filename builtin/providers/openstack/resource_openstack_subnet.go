package openstack

import (
	"github.com/haklop/gophercloud-extensions/network"
	"github.com/hashicorp/terraform/helper/diff"
	"github.com/hashicorp/terraform/terraform"
	"github.com/racker/perigee"
	"log"
	"strconv"
)

func resource_openstack_subnet_create(
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

	newSubnet := network.NewSubnet{
		NetworkId: rs.Attributes["network_id"],
		Name:      rs.Attributes["name"],
		Cidr:      rs.Attributes["cidr"],
	}

	enableDhcp, err := strconv.ParseBool(rs.Attributes["enable_dhcp"])
	if err != nil {
		return nil, err
	}
	newSubnet.EnableDhcp = enableDhcp

	ipVersion, err := strconv.Atoi(rs.Attributes["ip_version"])
	if err != nil {
		return nil, err
	}
	newSubnet.IPVersion = ipVersion

	createdSubnet, err := networksApi.CreateSubnet(newSubnet)

	log.Printf("[DEBUG] Create subnet: %s", createdSubnet.Id)

	rs.ID = createdSubnet.Id
	rs.Attributes["id"] = createdSubnet.Id

	return rs, err
}

func resource_openstack_subnet_destroy(
	s *terraform.ResourceState,
	meta interface{}) error {

	p := meta.(*ResourceProvider)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return err
	}

	err = networksApi.DeleteSubnet(s.ID)

	log.Printf("[DEBUG] Destroy subnet: %s", s.ID)

	return err
}

func resource_openstack_subnet_refresh(
	s *terraform.ResourceState,
	meta interface{}) (*terraform.ResourceState, error) {

	log.Printf("[DEBUG] Retrieve information about subnet: %s", s.ID)

	p := meta.(*ResourceProvider)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return nil, err
	}

	subnet, err := networksApi.GetSubnet(s.ID)
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

	s.Attributes["name"] = subnet.Name
	s.Attributes["enable_dhcp"] = strconv.FormatBool(subnet.EnableDhcp)

	return s, nil
}

func resource_openstack_subnet_update(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {

	p := meta.(*ResourceProvider)
	networksApi, err := p.getNetworkApi()
	if err != nil {
		return nil, err
	}

	rs := s.MergeDiff(d)

	desiredSubnet := network.UpdatedSubnet{}

	if attr, ok := d.Attributes["name"]; ok {
		desiredSubnet.Name = attr.New
	}

	if attr, ok := d.Attributes["enable_dhcp"]; ok {
		desiredSubnet.EnableDhcp, err = strconv.ParseBool(attr.New)
		if err != nil {
			return nil, err
		}
	}

	newSubnet, err := networksApi.UpdateSubnet(rs.ID, desiredSubnet)
	if err != nil {
		return nil, err
	}

	rs.Attributes["name"] = newSubnet.Name
	rs.Attributes["enable_dhcp"] = strconv.FormatBool(newSubnet.EnableDhcp)

	return rs, nil
}

func resource_openstack_subnet_diff(
	s *terraform.ResourceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.ResourceDiff, error) {

	b := &diff.ResourceBuilder{
		Attrs: map[string]diff.AttrType{
			"name":        diff.AttrTypeUpdate,
			"cidr":        diff.AttrTypeCreate,
			"ip_version":  diff.AttrTypeCreate,
			"enable_dhcp": diff.AttrTypeUpdate,
			"network_id":  diff.AttrTypeCreate,
		},

		ComputedAttrs: []string{
			"id",
		},
	}

	return b.Diff(s, c)
}
