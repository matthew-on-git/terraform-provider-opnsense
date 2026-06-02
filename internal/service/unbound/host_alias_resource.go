// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure hostAliasResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &hostAliasResource{}
	_ resource.ResourceWithImportState = &hostAliasResource{}
)

// hostAliasReqOpts configures the OPNsense API endpoints for Unbound host aliases.
var hostAliasReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/unbound/settings/add_host_alias",
	GetEndpoint:         "/api/unbound/settings/get_host_alias",
	UpdateEndpoint:      "/api/unbound/settings/set_host_alias",
	DeleteEndpoint:      "/api/unbound/settings/del_host_alias",
	SearchEndpoint:      "/api/unbound/settings/search_host_alias",
	ReconfigureEndpoint: "/api/unbound/service/reconfigure",
	Monad:               "alias",
}

// hostAliasResource implements the opnsense_unbound_host_alias resource.
type hostAliasResource struct {
	client *opnsense.Client
}

func newHostAliasResource() resource.Resource {
	return &hostAliasResource{}
}

// Metadata sets the resource type name.
func (r *hostAliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_host_alias"
}

// Configure extracts the OPNsense API client from provider data.
func (r *hostAliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new Unbound host alias via the OPNsense API.
func (r *hostAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HostAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, hostAliasReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Unbound host alias",
			fmt.Sprintf("Could not create Unbound host alias: %s", err),
		)
		return
	}

	result, err := opnsense.Get[hostAliasAPIResponse](ctx, r.client, hostAliasReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Unbound host alias after create",
			fmt.Sprintf("Created host alias %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *hostAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HostAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[hostAliasAPIResponse](ctx, r.client, hostAliasReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Unbound host alias",
			fmt.Sprintf("Could not read Unbound host alias %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing Unbound host alias via the OPNsense API.
func (r *hostAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan HostAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state HostAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, hostAliasReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Unbound host alias",
			fmt.Sprintf("Could not update Unbound host alias %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[hostAliasAPIResponse](ctx, r.client, hostAliasReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Unbound host alias after update",
			fmt.Sprintf("Updated host alias %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an Unbound host alias from the OPNsense API.
func (r *hostAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state HostAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, hostAliasReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Unbound host alias",
			fmt.Sprintf("Could not delete Unbound host alias %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing Unbound host alias by UUID.
func (r *hostAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
