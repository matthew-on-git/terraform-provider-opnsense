// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &reservationResource{}
	_ resource.ResourceWithImportState = &reservationResource{}
)

var reservationReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/kea/dhcpv4/add_reservation",
	GetEndpoint:         "/api/kea/dhcpv4/get_reservation",
	UpdateEndpoint:      "/api/kea/dhcpv4/set_reservation",
	DeleteEndpoint:      "/api/kea/dhcpv4/del_reservation",
	SearchEndpoint:      "/api/kea/dhcpv4/search_reservation",
	ReconfigureEndpoint: "/api/kea/service/reconfigure",
	Monad:               "reservation",
}

type reservationResource struct{ client *opnsense.Client }

func newReservationResource() resource.Resource { return &reservationResource{} }

func (r *reservationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcpv4_reservation"
}

func (r *reservationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	r.client = client
}

func (r *reservationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ReservationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, reservationReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating DHCPv4 reservation", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[reservationAPIResponse](ctx, r.client, reservationReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DHCPv4 reservation after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reservationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ReservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[reservationAPIResponse](ctx, r.client, reservationReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading DHCPv4 reservation", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *reservationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ReservationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, reservationReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating DHCPv4 reservation", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[reservationAPIResponse](ctx, r.client, reservationReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DHCPv4 reservation after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reservationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ReservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, reservationReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting DHCPv4 reservation", fmt.Sprintf("%s", err))
	}
}

func (r *reservationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
