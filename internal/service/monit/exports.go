// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package monit implements Terraform resources for OPNsense Monit monitoring.
package monit

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of Monit resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newServiceResource,
		newTestResource,
		newAlertResource,
	}
}

// DataSources returns the list of Monit data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
