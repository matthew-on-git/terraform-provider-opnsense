// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import "github.com/hashicorp/terraform-plugin-framework/datasource"

func newInterfacesOverviewAllDataSource() datasource.DataSource {
	return &liveInterfaceDataSource{name: "interfaces_overview_all", overview: true, all: true}
}
