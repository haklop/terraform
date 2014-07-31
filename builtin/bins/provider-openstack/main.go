package main

import (
	"github.com/hashicorp/terraform/builtin/providers/openstack"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(new(openstack.ResourceProvider))
}
