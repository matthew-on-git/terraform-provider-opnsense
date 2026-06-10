// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func configureQuaggaDataSource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse, setClient func(*opnsense.Client)) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	setClient(client)
}
