// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package ipsec implements Terraform resources for OPNsense IPsec VPN management.
package ipsec

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of ipsec resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newConnectionResource,
		newChildResource,
		newPSKResource,
		newPoolResource,
		newVTIResource,
		newManualSPDResource,
		newLocalResource,
		newRemoteResource,
		newKeyPairResource,
	}
}

// DataSources returns the list of ipsec data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newConnectionDataSource,
		newChildDataSource,
		newLocalDataSource,
		newManualSPDDataSource,
		newPoolDataSource,
		newRemoteDataSource,
		newVTIDataSource,
	}
}
