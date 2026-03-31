// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package unbound implements Terraform resources for OPNsense Unbound DNS management.
package unbound

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of Unbound DNS resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newHostOverrideResource,
		newDomainOverrideResource,
		newACLResource,
	}
}

// DataSources returns the list of Unbound DNS data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
