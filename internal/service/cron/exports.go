// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package cron implements Terraform resources for OPNsense cron jobs.
package cron

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of cron resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newJobResource,
	}
}

// DataSources returns the list of cron data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
