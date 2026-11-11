// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import "github.com/hashicorp/terraform-plugin-framework/datasource"

func newInterfaceDataSource() datasource.DataSource {
	return &liveInterfaceDataSource{name: "interface", overview: false, all: false}
}
