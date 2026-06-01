// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure remoteResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &remoteResource{}
	_ resource.ResourceWithImportState = &remoteResource{}
)

// remoteReqOpts configures the OPNsense API endpoints for IPsec local auth entries.
var remoteReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/ipsec/connections/add_remote",
	GetEndpoint:         "/api/ipsec/connections/get_remote",
	UpdateEndpoint:      "/api/ipsec/connections/set_remote",
	DeleteEndpoint:      "/api/ipsec/connections/del_remote",
	SearchEndpoint:      "/api/ipsec/connections/search_remote",
	ReconfigureEndpoint: "/api/ipsec/service/reconfigure",
	Monad:               "remote",
}

// remoteResource implements the opnsense_ipsec_remote resource.
type remoteResource struct {
	client *opnsense.Client
}

func newRemoteResource() resource.Resource {
	return &remoteResource{}
}

// Metadata sets the resource type name.
func (r *remoteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_remote"
}

// Configure extracts the OPNsense API client from provider data.
func (r *remoteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new IPsec remote auth entry via the OPNsense API.
func (r *remoteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RemoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, remoteReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IPsec remote auth entry",
			fmt.Sprintf("Could not create IPsec remote auth entry: %s", err),
		)
		return
	}

	result, err := opnsense.Get[remoteAPIResponse](ctx, r.client, remoteReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec remote auth entry after create",
			fmt.Sprintf("Created remote entry %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *remoteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RemoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[remoteAPIResponse](ctx, r.client, remoteReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading IPsec remote auth entry",
			fmt.Sprintf("Could not read IPsec remote auth entry %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing IPsec remote auth entry via the OPNsense API.
func (r *remoteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RemoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RemoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, remoteReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IPsec remote auth entry",
			fmt.Sprintf("Could not update IPsec remote auth entry %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[remoteAPIResponse](ctx, r.client, remoteReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec remote auth entry after update",
			fmt.Sprintf("Updated remote entry %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an IPsec remote auth entry from the OPNsense API.
func (r *remoteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RemoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, remoteReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting IPsec remote auth entry",
			fmt.Sprintf("Could not delete IPsec remote auth entry %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing IPsec remote auth entry by UUID.
func (r *remoteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
