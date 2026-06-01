// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package auth implements Terraform resources for OPNsense local users and groups.
package auth

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of auth resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newGroupResource,
		newUserResource,
	}
}

// DataSources returns the list of auth data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
