// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package dhcp implements Terraform resources for OPNsense Kea DHCP management.
package dhcp

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of DHCP resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newSubnetResource,
		newReservationResource,
	}
}

// DataSources returns the list of DHCP data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
