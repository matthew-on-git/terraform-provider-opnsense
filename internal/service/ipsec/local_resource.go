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

// Ensure localResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &localResource{}
	_ resource.ResourceWithImportState = &localResource{}
)

// localReqOpts configures the OPNsense API endpoints for IPsec local auth entries.
var localReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/ipsec/connections/add_local",
	GetEndpoint:         "/api/ipsec/connections/get_local",
	UpdateEndpoint:      "/api/ipsec/connections/set_local",
	DeleteEndpoint:      "/api/ipsec/connections/del_local",
	SearchEndpoint:      "/api/ipsec/connections/search_local",
	ReconfigureEndpoint: "/api/ipsec/service/reconfigure",
	Monad:               "local",
}

// localResource implements the opnsense_ipsec_local resource.
type localResource struct {
	client *opnsense.Client
}

func newLocalResource() resource.Resource {
	return &localResource{}
}

// Metadata sets the resource type name.
func (r *localResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_local"
}

// Configure extracts the OPNsense API client from provider data.
func (r *localResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new IPsec local auth entry via the OPNsense API.
func (r *localResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LocalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, localReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IPsec local auth entry",
			fmt.Sprintf("Could not create IPsec local auth entry: %s", err),
		)
		return
	}

	result, err := opnsense.Get[localAPIResponse](ctx, r.client, localReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec local auth entry after create",
			fmt.Sprintf("Created local entry %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *localResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LocalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[localAPIResponse](ctx, r.client, localReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading IPsec local auth entry",
			fmt.Sprintf("Could not read IPsec local auth entry %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing IPsec local auth entry via the OPNsense API.
func (r *localResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LocalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LocalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, localReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IPsec local auth entry",
			fmt.Sprintf("Could not update IPsec local auth entry %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[localAPIResponse](ctx, r.client, localReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec local auth entry after update",
			fmt.Sprintf("Updated local entry %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an IPsec local auth entry from the OPNsense API.
func (r *localResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LocalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, localReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting IPsec local auth entry",
			fmt.Sprintf("Could not delete IPsec local auth entry %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing IPsec local auth entry by UUID.
func (r *localResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
