// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package syslog implements Terraform resources for OPNsense syslog.
package syslog

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of syslog resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newDestinationResource,
	}
}

// DataSources returns the list of syslog data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newDestinationDataSource,
	}
}
