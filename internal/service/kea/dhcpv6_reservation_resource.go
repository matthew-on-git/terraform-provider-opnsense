// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure dhcpv6ReservationResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &dhcpv6ReservationResource{}
	_ resource.ResourceWithImportState = &dhcpv6ReservationResource{}
)

// dhcpv6ReservationReqOpts configures the OPNsense API endpoints for IPsec local auth entries.
var dhcpv6ReservationReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/kea/dhcpv6/add_reservation",
	GetEndpoint:         "/api/kea/dhcpv6/get_reservation",
	UpdateEndpoint:      "/api/kea/dhcpv6/set_reservation",
	DeleteEndpoint:      "/api/kea/dhcpv6/del_reservation",
	SearchEndpoint:      "/api/kea/dhcpv6/search_reservation",
	ReconfigureEndpoint: "/api/kea/service/reconfigure",
	Monad:               "reservation",
}

// dhcpv6ReservationResource implements the opnsense_kea_dhcpv6_reservation resource.
type dhcpv6ReservationResource struct {
	client *opnsense.Client
}

func newDHCPv6ReservationResource() resource.Resource {
	return &dhcpv6ReservationResource{}
}

// Metadata sets the resource type name.
func (r *dhcpv6ReservationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_dhcpv6_reservation"
}

// Configure extracts the OPNsense API client from provider data.
func (r *dhcpv6ReservationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			"Expected *opnsense.Client, got something else.",
		)
		return
	}
	r.client = client
}

// Create creates a new Kea DHCPv6 reservation via the OPNsense API.
func (r *dhcpv6ReservationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DHCPv6ReservationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, dhcpv6ReservationReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Kea DHCPv6 reservation",
			fmt.Sprintf("Could not create Kea DHCPv6 reservation: %s", err),
		)
		return
	}

	result, err := opnsense.Get[dhcpv6ReservationAPIResponse](ctx, r.client, dhcpv6ReservationReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Kea DHCPv6 reservation after create",
			fmt.Sprintf("Created reservation %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *dhcpv6ReservationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DHCPv6ReservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[dhcpv6ReservationAPIResponse](ctx, r.client, dhcpv6ReservationReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Kea DHCPv6 reservation",
			fmt.Sprintf("Could not read Kea DHCPv6 reservation %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing Kea DHCPv6 reservation via the OPNsense API.
func (r *dhcpv6ReservationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DHCPv6ReservationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DHCPv6ReservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, dhcpv6ReservationReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Kea DHCPv6 reservation",
			fmt.Sprintf("Could not update Kea DHCPv6 reservation %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[dhcpv6ReservationAPIResponse](ctx, r.client, dhcpv6ReservationReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Kea DHCPv6 reservation after update",
			fmt.Sprintf("Updated reservation %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an Kea DHCPv6 reservation from the OPNsense API.
func (r *dhcpv6ReservationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DHCPv6ReservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, dhcpv6ReservationReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Kea DHCPv6 reservation",
			fmt.Sprintf("Could not delete Kea DHCPv6 reservation %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing Kea DHCPv6 reservation by UUID.
func (r *dhcpv6ReservationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
