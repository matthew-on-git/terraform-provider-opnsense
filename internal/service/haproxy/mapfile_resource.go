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
	_ resource.Resource                = &mapfileResource{}
	_ resource.ResourceWithImportState = &mapfileResource{}
)

var mapfileReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/haproxy/settings/addMapFile",
	GetEndpoint:         "/api/haproxy/settings/getMapFile",
	UpdateEndpoint:      "/api/haproxy/settings/setMapFile",
	DeleteEndpoint:      "/api/haproxy/settings/delMapFile",
	SearchEndpoint:      "/api/haproxy/settings/searchMapFiles",
	ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
	Monad:               "mapfile",
}

type mapfileResource struct{ client *opnsense.Client }

func newMapfileResource() resource.Resource { return &mapfileResource{} }

func (r *mapfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_mapfile"
}

func (r *mapfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client, got something else.")
		return
	}
	r.client = client
}

func (r *mapfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MapfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid, err := opnsense.Add(ctx, r.client, mapfileReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating HAProxy map file", fmt.Sprintf("Could not create map file: %s", err))
		return
	}

	result, err := opnsense.Get[mapfileAPIResponse](ctx, r.client, mapfileReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy map file after create", fmt.Sprintf("Created map file %s but could not read it back: %s", uuid, err))
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mapfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MapfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[mapfileAPIResponse](ctx, r.client, mapfileReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading HAProxy map file", fmt.Sprintf("Could not read map file %s: %s", state.ID.ValueString(), err))
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mapfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MapfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MapfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, mapfileReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating HAProxy map file", fmt.Sprintf("Could not update map file %s: %s", id, err))
		return
	}

	result, err := opnsense.Get[mapfileAPIResponse](ctx, r.client, mapfileReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy map file after update", fmt.Sprintf("Updated map file %s but could not read it back: %s", id, err))
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mapfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MapfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := opnsense.Delete(ctx, r.client, mapfileReqOpts, state.ID.ValueString()); err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError("Error deleting HAProxy map file", fmt.Sprintf("Could not delete map file %s: %s", state.ID.ValueString(), err))
	}
}

func (r *mapfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
