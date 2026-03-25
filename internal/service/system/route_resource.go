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
	_ resource.Resource                = &routeResource{}
	_ resource.ResourceWithImportState = &routeResource{}
)

var routeReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/routes/routes/addroute",
	GetEndpoint:         "/api/routes/routes/getroute",
	UpdateEndpoint:      "/api/routes/routes/setroute",
	DeleteEndpoint:      "/api/routes/routes/delroute",
	SearchEndpoint:      "/api/routes/routes/searchroute",
	ReconfigureEndpoint: "/api/routes/routes/reconfigure",
	Monad:               "route",
}

type routeResource struct{ client *opnsense.Client }

func newRouteResource() resource.Resource { return &routeResource{} }

func (r *routeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_route"
}

func (r *routeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *routeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, routeReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating static route", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[routeAPIResponse](ctx, r.client, routeReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading static route after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[routeAPIResponse](ctx, r.client, routeReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading static route", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, routeReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating static route", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[routeAPIResponse](ctx, r.client, routeReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading static route after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, routeReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting static route", fmt.Sprintf("%s", err))
	}
}

func (r *routeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
