// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package openvpn implements Terraform resources for OPNsense OpenVPN management.
package openvpn

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of OpenVPN resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newInstanceResource,
		newStaticKeyResource,
		newClientOverwriteResource,
	}
}

// DataSources returns the list of OpenVPN data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
