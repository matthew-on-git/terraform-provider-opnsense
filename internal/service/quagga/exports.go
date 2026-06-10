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
		newGeneralResource,
		newBGPGlobalResource,
		newBGPNeighborResource,
		newPrefixListResource,
		newRouteMapResource,
		newRIPResource,
		newOSPFGeneralResource,
		newOSPF6GeneralResource,
		newStaticGeneralResource,
		newStaticRouteResource,
		newBGPASPathResource,
		newBGPCommunityListResource,
		newBGPPeerGroupResource,
		newBGPRedistributionResource,
		// Generated OSPF item resources (internal/generate).
		newOSPFAreaResource,
		newOSPFNetworkResource,
		newOSPFInterfaceResource,
		newOSPFNeighborResource,
		newOSPFPrefixListResource,
		newOSPFRouteMapResource,
		newOSPFRedistributionResource,
		// Generated OSPFv3 item resources (internal/generate).
		newOSPF6NetworkResource,
		newOSPF6InterfaceResource,
		newOSPF6PrefixListResource,
		newOSPF6RouteMapResource,
		newOSPF6RedistributionResource,
	}
}

// DataSources returns the list of Quagga/FRR data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newBGPASPathDataSource,
		newBGPCommunityListDataSource,
		newBGPNeighborDataSource,
		newBGPPeerGroupDataSource,
		newBGPRedistributionDataSource,
		newOSPF6InterfaceDataSource,
		newOSPF6NetworkDataSource,
		newOSPF6PrefixListDataSource,
		newOSPF6RedistributionDataSource,
		newOSPF6RouteMapDataSource,
		newOSPFAreaDataSource,
		newOSPFInterfaceDataSource,
		newOSPFNeighborDataSource,
		newOSPFNetworkDataSource,
		newOSPFPrefixListDataSource,
		newOSPFRedistributionDataSource,
		newOSPFRouteMapDataSource,
		newPrefixListDataSource,
		newRouteMapDataSource,
		newStaticRouteDataSource,
	}
}
