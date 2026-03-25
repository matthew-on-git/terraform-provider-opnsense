// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package haproxy implements Terraform resources for OPNsense HAProxy plugin management.
package haproxy

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of HAProxy resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newServerResource,
		newBackendResource,
		newFrontendResource,
		newACLResource,
		newHealthcheckResource,
	}
}

// DataSources returns the list of HAProxy data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
