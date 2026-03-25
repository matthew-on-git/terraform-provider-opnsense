// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &gatewayResource{}
	_ resource.ResourceWithImportState = &gatewayResource{}
)

var gatewayReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/routing/settings/add_gateway",
	GetEndpoint:         "/api/routing/settings/get_gateway",
	UpdateEndpoint:      "/api/routing/settings/set_gateway",
	DeleteEndpoint:      "/api/routing/settings/del_gateway",
	SearchEndpoint:      "/api/routing/settings/search_gateway",
	ReconfigureEndpoint: "/api/routing/settings/reconfigure",
	Monad:               "gateway",
}

type gatewayResource struct{ client *opnsense.Client }

func newGatewayResource() resource.Resource { return &gatewayResource{} }

func (r *gatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_gateway"
}

func (r *gatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GatewayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, gatewayReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating gateway", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[gatewayAPIResponse](ctx, r.client, gatewayReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading gateway after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GatewayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[gatewayAPIResponse](ctx, r.client, gatewayReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading gateway", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GatewayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, gatewayReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating gateway", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[gatewayAPIResponse](ctx, r.client, gatewayReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading gateway after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GatewayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, gatewayReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting gateway", fmt.Sprintf("%s", err))
	}
}

func (r *gatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
