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

// Ensure keyPairResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &keyPairResource{}
	_ resource.ResourceWithImportState = &keyPairResource{}
)

// keyPairReqOpts configures the OPNsense API endpoints for IPsec local auth entries.
var keyPairReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/ipsec/key_pairs/addItem",
	GetEndpoint:         "/api/ipsec/key_pairs/getItem",
	UpdateEndpoint:      "/api/ipsec/key_pairs/setItem",
	DeleteEndpoint:      "/api/ipsec/key_pairs/delItem",
	SearchEndpoint:      "/api/ipsec/key_pairs/searchItem",
	ReconfigureEndpoint: "/api/ipsec/service/reconfigure",
	Monad:               "keyPair",
}

// keyPairResource implements the opnsense_ipsec_key_pair resource.
type keyPairResource struct {
	client *opnsense.Client
}

func newKeyPairResource() resource.Resource {
	return &keyPairResource{}
}

// Metadata sets the resource type name.
func (r *keyPairResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_key_pair"
}

// Configure extracts the OPNsense API client from provider data.
func (r *keyPairResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new IPsec key pair via the OPNsense API.
func (r *keyPairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KeyPairResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, keyPairReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IPsec key pair",
			fmt.Sprintf("Could not create IPsec key pair: %s", err),
		)
		return
	}

	result, err := opnsense.Get[keyPairAPIResponse](ctx, r.client, keyPairReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec key pair after create",
			fmt.Sprintf("Created key pair %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *keyPairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KeyPairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[keyPairAPIResponse](ctx, r.client, keyPairReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading IPsec key pair",
			fmt.Sprintf("Could not read IPsec key pair %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing IPsec key pair via the OPNsense API.
func (r *keyPairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan KeyPairResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state KeyPairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, keyPairReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IPsec key pair",
			fmt.Sprintf("Could not update IPsec key pair %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[keyPairAPIResponse](ctx, r.client, keyPairReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec key pair after update",
			fmt.Sprintf("Updated key pair %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an IPsec key pair from the OPNsense API.
func (r *keyPairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KeyPairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, keyPairReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting IPsec key pair",
			fmt.Sprintf("Could not delete IPsec key pair %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing IPsec key pair by UUID.
func (r *keyPairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
