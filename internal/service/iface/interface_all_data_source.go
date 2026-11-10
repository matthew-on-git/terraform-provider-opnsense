// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import "github.com/hashicorp/terraform-plugin-framework/datasource"

func newInterfaceAllDataSource() datasource.DataSource {
	return &liveInterfaceDataSource{name: "interface_all", overview: false, all: true}
}
