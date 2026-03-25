// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package system implements Terraform resources for OPNsense core infrastructure.
package system

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of system resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newVlanResource,
		newVipResource,
		newRouteResource,
		newGatewayResource,
	}
}

// DataSources returns the list of system data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
