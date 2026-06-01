// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package kea implements Terraform resources for OPNsense Kea DHCP.
package kea

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of Kea resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newCtrlAgentResource,
	}
}

// DataSources returns the list of Kea data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
