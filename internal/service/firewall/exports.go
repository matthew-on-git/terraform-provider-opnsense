// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package firewall implements Terraform resources for OPNsense firewall management.
package firewall

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of firewall resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newAliasResource,
		newCategoryResource,
		newFilterRuleResource,
		newNatPortForwardResource,
		newNatOutboundResource,
	}
}

// DataSources returns the list of firewall data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newAliasDataSource,
	}
}
