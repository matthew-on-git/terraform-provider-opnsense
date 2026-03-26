// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package acme implements Terraform resources for OPNsense ACME client plugin management.
package acme

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resources returns the list of ACME resource types.
func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newAccountResource,
		newChallengeResource,
		newCertificateResource,
	}
}

// DataSources returns the list of ACME data source types.
func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
