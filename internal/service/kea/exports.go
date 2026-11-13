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
		newDDNSSettingsResource,
		newDHCPv4SettingsResource,
		newDHCPv4SubnetResource,
		newDHCPv4OptionResource,
		newDHCPv4ReservationResource,
		newDHCPv6SettingsResource,
		newDHCPv6SubnetResource,
		newDHCPv6ReservationResource,
		newHAPeerResource,
		newDHCPv6PeerResource,
		newDHCPv6PDPoolResource,
	}
}

// DataSources returns the list of Kea data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newCtrlAgentDataSource,
		newDDNSSettingsDataSource,
		newDHCPv4SettingsDataSource,
		newDHCPv4SubnetDataSource,
		newDHCPv4OptionDataSource,
		newDHCPv4ReservationDataSource,
		newDHCPv6SubnetDataSource,
		newDHCPv6ReservationDataSource,
		newHAPeerDataSource,
		newDHCPv6PeerDataSource,
		newDHCPv6PDPoolDataSource,
	}
}
