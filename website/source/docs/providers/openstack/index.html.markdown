---
layout: "openstack"
page_title: "Provider: OpenStack"
sidebar_current: "docs-openstack-index"
---

# OpenStack Provider

The OpenStack provider is used to interact with the
many resources supported by OpenStack. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```
# Configure the OpenStack Provider
provider "openstack" {
    auth_url = "${var.os_auth_url}"
    tenant_id = "${var.os_tenant_id}"
    username = "${var.os_username}"
    password = "${var.os_password}"
}

# Create a web server
resource "openstack_compute" "web" {
    ...
}
```

## Argument Reference

The following arguments are supported:

* `auth_url` - (Required) This is the AWS access key. It must be provided, but
  it can also be sourced from the `OS_AUTH_URL` environment variable.

* `tenant_id` - (Required) This is the AWS secret key. It must be provided, but
  it can also be sourced from the `OS_TENANT_ID` environment variable.

* `username` - (Required) This is the AWS region. It must be provided, but
  it can also be sourced from the `OS_USERNAME` environment variables.

* `password` - (Required) This is the AWS region. It must be provided, but
  it can also be sourced from the `OS_PASSWORD` environment variables.
