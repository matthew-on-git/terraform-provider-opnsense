// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure aliasResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &aliasResource{}
	_ resource.ResourceWithImportState = &aliasResource{}
)

// aliasReqOpts configures the OPNsense API endpoints for firewall aliases.
var aliasReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/firewall/alias/addItem",
	GetEndpoint:         "/api/firewall/alias/getItem",
	UpdateEndpoint:      "/api/firewall/alias/setItem",
	DeleteEndpoint:      "/api/firewall/alias/delItem",
	SearchEndpoint:      "/api/firewall/alias/searchItem",
	ReconfigureEndpoint: "/api/firewall/alias/reconfigure",
	Monad:               "alias",
}

// aliasResource implements the opnsense_firewall_alias resource.
type aliasResource struct {
	client *opnsense.Client
}

func newAliasResource() resource.Resource {
	return &aliasResource{}
}

// Metadata sets the resource type name.
func (r *aliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_alias"
}

// Configure extracts the OPNsense API client from provider data.
func (r *aliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new firewall alias via the OPNsense API.
func (r *aliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, aliasReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating firewall alias",
			fmt.Sprintf("Could not create firewall alias: %s", err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[aliasAPIResponse](ctx, r.client, aliasReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading firewall alias after create",
			fmt.Sprintf("Created alias %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *aliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[aliasAPIResponse](ctx, r.client, aliasReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading firewall alias",
			fmt.Sprintf("Could not read firewall alias %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing firewall alias via the OPNsense API.
func (r *aliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, aliasReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating firewall alias",
			fmt.Sprintf("Could not update firewall alias %s: %s", id, err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[aliasAPIResponse](ctx, r.client, aliasReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading firewall alias after update",
			fmt.Sprintf("Updated alias %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a firewall alias from the OPNsense API.
func (r *aliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, aliasReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting firewall alias",
			fmt.Sprintf("Could not delete firewall alias %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing firewall alias by UUID.
func (r *aliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
