// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package wireguard implements Terraform resources for OPNsense WireGuard VPN management.
package wireguard

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of wireguard resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newServerResource,
		newPeerResource,
	}
}

// DataSources returns the list of wireguard data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
