// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package iface implements Terraform resources for OPNsense interface types.
package iface

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of interface resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newLoopbackResource,
		newVxlanResource,
		newNeighborResource,
		newGREResource,
		newGIFResource,
		newBridgeResource,
	}
}

// DataSources returns the list of interface data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newBridgeDataSource,
		newGIFDataSource,
		newGREDataSource,
		newLoopbackDataSource,
		newNeighborDataSource,
		newVxlanDataSource,
	}
}
