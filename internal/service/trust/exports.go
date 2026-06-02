// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package trust implements Terraform resources for OPNsense PKI (CAs, certificates).
package trust

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of trust (PKI) resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newCAResource,
		newCertResource,
	}
}

// DataSources returns the list of trust data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
