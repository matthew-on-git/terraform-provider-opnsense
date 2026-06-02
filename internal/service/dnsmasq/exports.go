// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package dnsmasq implements Terraform resources for OPNsense Dnsmasq DNS/DHCP.
package dnsmasq

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of Dnsmasq resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newSettingsResource,
	}
}

// DataSources returns the list of Dnsmasq data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
