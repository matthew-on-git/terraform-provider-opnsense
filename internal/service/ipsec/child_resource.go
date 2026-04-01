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

// Ensure childResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &childResource{}
	_ resource.ResourceWithImportState = &childResource{}
)

// childReqOpts configures the OPNsense API endpoints for IPsec child SAs.
var childReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/ipsec/connections/add_child",
	GetEndpoint:         "/api/ipsec/connections/get_child",
	UpdateEndpoint:      "/api/ipsec/connections/set_child",
	DeleteEndpoint:      "/api/ipsec/connections/del_child",
	SearchEndpoint:      "/api/ipsec/connections/search_child",
	ReconfigureEndpoint: "/api/ipsec/service/reconfigure",
	Monad:               "child",
}

// childResource implements the opnsense_ipsec_child resource.
type childResource struct {
	client *opnsense.Client
}

func newChildResource() resource.Resource {
	return &childResource{}
}

// Metadata sets the resource type name.
func (r *childResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_child"
}

// Configure extracts the OPNsense API client from provider data.
func (r *childResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new IPsec child SA via the OPNsense API.
func (r *childResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ChildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, childReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IPsec child SA",
			fmt.Sprintf("Could not create IPsec child SA: %s", err),
		)
		return
	}

	result, err := opnsense.Get[ipsecChildAPIResponse](ctx, r.client, childReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec child SA after create",
			fmt.Sprintf("Created child SA %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *childResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ChildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[ipsecChildAPIResponse](ctx, r.client, childReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading IPsec child SA",
			fmt.Sprintf("Could not read IPsec child SA %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing IPsec child SA via the OPNsense API.
func (r *childResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ChildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ChildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, childReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IPsec child SA",
			fmt.Sprintf("Could not update IPsec child SA %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[ipsecChildAPIResponse](ctx, r.client, childReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec child SA after update",
			fmt.Sprintf("Updated child SA %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an IPsec child SA from the OPNsense API.
func (r *childResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ChildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, childReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting IPsec child SA",
			fmt.Sprintf("Could not delete IPsec child SA %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing IPsec child SA by UUID.
func (r *childResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
