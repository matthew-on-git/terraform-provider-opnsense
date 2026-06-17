// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &actionResource{}
	_ resource.ResourceWithImportState = &actionResource{}
)

var actionReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/haproxy/settings/addAction",
	GetEndpoint:         "/api/haproxy/settings/getAction",
	UpdateEndpoint:      "/api/haproxy/settings/setAction",
	DeleteEndpoint:      "/api/haproxy/settings/delAction",
	SearchEndpoint:      "/api/haproxy/settings/searchActions",
	ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
	Monad:               "action",
}

type actionResource struct{ client *opnsense.Client }

func newActionResource() resource.Resource { return &actionResource{} }

func (r *actionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_action"
}

func (r *actionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *actionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, actionReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating HAProxy action", fmt.Sprintf("Could not create HAProxy action: %s", err))
		return
	}
	result, err := opnsense.Get[actionAPIResponse](ctx, r.client, actionReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy action after create", fmt.Sprintf("Created action %s but could not read it back: %s", uuid, err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *actionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[actionAPIResponse](ctx, r.client, actionReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading HAProxy action", fmt.Sprintf("Could not read HAProxy action %s: %s", state.ID.ValueString(), err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *actionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, actionReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating HAProxy action", fmt.Sprintf("Could not update HAProxy action %s: %s", id, err))
		return
	}
	result, err := opnsense.Get[actionAPIResponse](ctx, r.client, actionReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy action after update", fmt.Sprintf("Updated action %s but could not read it back: %s", id, err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *actionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, actionReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting HAProxy action", fmt.Sprintf("Could not delete HAProxy action %s: %s", state.ID.ValueString(), err))
	}
}

func (r *actionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
