// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure accountResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &accountResource{}
	_ resource.ResourceWithImportState = &accountResource{}
)

// accountReqOpts configures the OPNsense API endpoints for ddclient accounts.
var accountReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/dyndns/accounts/add_item",
	GetEndpoint:         "/api/dyndns/accounts/get_item",
	UpdateEndpoint:      "/api/dyndns/accounts/set_item",
	DeleteEndpoint:      "/api/dyndns/accounts/del_item",
	SearchEndpoint:      "/api/dyndns/accounts/search_item",
	ReconfigureEndpoint: "/api/dyndns/service/reconfigure",
	Monad:               "account",
}

// accountResource implements the opnsense_ddclient_account resource.
type accountResource struct {
	client *opnsense.Client
}

func newAccountResource() resource.Resource {
	return &accountResource{}
}

// Metadata sets the resource type name.
func (r *accountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ddclient_account"
}

// Configure extracts the OPNsense API client from provider data.
func (r *accountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new ddclient account via the OPNsense API.
func (r *accountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, accountReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ddclient account",
			fmt.Sprintf("Could not create ddclient account: %s", err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[ddclientAccountAPIResponse](ctx, r.client, accountReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ddclient account after create",
			fmt.Sprintf("Created account %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *accountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[ddclientAccountAPIResponse](ctx, r.client, accountReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading ddclient account",
			fmt.Sprintf("Could not read ddclient account %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing ddclient account via the OPNsense API.
func (r *accountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, accountReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ddclient account",
			fmt.Sprintf("Could not update ddclient account %s: %s", id, err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[ddclientAccountAPIResponse](ctx, r.client, accountReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ddclient account after update",
			fmt.Sprintf("Updated account %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a ddclient account from the OPNsense API.
func (r *accountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, accountReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting ddclient account",
			fmt.Sprintf("Could not delete ddclient account %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing ddclient account by UUID.
func (r *accountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
