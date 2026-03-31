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

// Ensure pskResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &pskResource{}
	_ resource.ResourceWithImportState = &pskResource{}
)

// pskReqOpts configures the OPNsense API endpoints for IPsec pre-shared keys.
var pskReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/ipsec/pre_shared_keys/add_item",
	GetEndpoint:         "/api/ipsec/pre_shared_keys/get_item",
	UpdateEndpoint:      "/api/ipsec/pre_shared_keys/set_item",
	DeleteEndpoint:      "/api/ipsec/pre_shared_keys/del_item",
	SearchEndpoint:      "/api/ipsec/pre_shared_keys/search_item",
	ReconfigureEndpoint: "/api/ipsec/service/reconfigure",
	Monad:               "preSharedKey",
}

// pskResource implements the opnsense_ipsec_psk resource.
type pskResource struct {
	client *opnsense.Client
}

func newPSKResource() resource.Resource {
	return &pskResource{}
}

// Metadata sets the resource type name.
func (r *pskResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_psk"
}

// Configure extracts the OPNsense API client from provider data.
func (r *pskResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new IPsec pre-shared key via the OPNsense API.
func (r *pskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PSKResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, pskReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IPsec pre-shared key",
			fmt.Sprintf("Could not create IPsec pre-shared key: %s", err),
		)
		return
	}

	result, err := opnsense.Get[ipsecPSKAPIResponse](ctx, r.client, pskReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec pre-shared key after create",
			fmt.Sprintf("Created pre-shared key %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *pskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PSKResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[ipsecPSKAPIResponse](ctx, r.client, pskReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading IPsec pre-shared key",
			fmt.Sprintf("Could not read IPsec pre-shared key %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing IPsec pre-shared key via the OPNsense API.
func (r *pskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PSKResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PSKResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, pskReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IPsec pre-shared key",
			fmt.Sprintf("Could not update IPsec pre-shared key %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[ipsecPSKAPIResponse](ctx, r.client, pskReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec pre-shared key after update",
			fmt.Sprintf("Updated pre-shared key %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an IPsec pre-shared key from the OPNsense API.
func (r *pskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PSKResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, pskReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting IPsec pre-shared key",
			fmt.Sprintf("Could not delete IPsec pre-shared key %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing IPsec pre-shared key by UUID.
func (r *pskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
