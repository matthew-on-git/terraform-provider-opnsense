// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package trafficshaper implements Terraform resources for OPNsense traffic shaping.
package trafficshaper

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of traffic shaper resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newPipeResource,
		newQueueResource,
		newRuleResource,
	}
}

// DataSources returns the list of traffic shaper data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
