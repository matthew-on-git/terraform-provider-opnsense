// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package quagga implements Terraform resources for OPNsense FRR/Quagga plugin management.
package quagga

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of Quagga/FRR resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newBGPNeighborResource,
		newPrefixListResource,
		newRouteMapResource,
	}
}

// DataSources returns the list of Quagga/FRR data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
