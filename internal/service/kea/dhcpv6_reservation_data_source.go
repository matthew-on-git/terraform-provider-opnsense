// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &dhcpv6ReservationDataSource{}

type dhcpv6ReservationDataSource struct{ client *opnsense.Client }

func newDHCPv6ReservationDataSource() datasource.DataSource { return &dhcpv6ReservationDataSource{} }

func (d *dhcpv6ReservationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_dhcpv6_reservation"
}

func (d *dhcpv6ReservationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing Kea DHCPv6 reservation on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":            dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"subnet_id":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Parent DHCPv6 subnet UUID."},
			"ip_address":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Reserved IPv6 address."},
			"duid":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Client DUID."},
			"hostname":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Reservation hostname."},
			"domain_search": dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Domain search list."},
			"description":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the reservation."},
		},
	}
}

func (d *dhcpv6ReservationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	d.client = client
}

func (d *dhcpv6ReservationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DHCPv6ReservationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[dhcpv6ReservationAPIResponse](ctx, d.client, dhcpv6ReservationReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Kea DHCPv6 reservation", fmt.Sprintf("Could not read Kea DHCPv6 reservation %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
