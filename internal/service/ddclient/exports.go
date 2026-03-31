// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package ddclient implements Terraform resources for OPNsense dynamic DNS client management.
package ddclient

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of ddclient resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newAccountResource,
	}
}

// DataSources returns the list of ddclient data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
