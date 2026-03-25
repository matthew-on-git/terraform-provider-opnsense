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

// Ensure frontendResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &frontendResource{}
	_ resource.ResourceWithImportState = &frontendResource{}
)

// frontendReqOpts configures the OPNsense API endpoints for HAProxy frontends.
var frontendReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/haproxy/settings/addFrontend",
	GetEndpoint:         "/api/haproxy/settings/getFrontend",
	UpdateEndpoint:      "/api/haproxy/settings/setFrontend",
	DeleteEndpoint:      "/api/haproxy/settings/delFrontend",
	SearchEndpoint:      "/api/haproxy/settings/searchFrontends",
	ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
	Monad:               "frontend",
}

// frontendResource implements the opnsense_haproxy_frontend resource.
type frontendResource struct {
	client *opnsense.Client
}

func newFrontendResource() resource.Resource {
	return &frontendResource{}
}

// Metadata sets the resource type name.
func (r *frontendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_frontend"
}

// Configure extracts the OPNsense API client from provider data.
func (r *frontendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new HAProxy frontend via the OPNsense API.
func (r *frontendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FrontendResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, frontendReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating HAProxy frontend",
			fmt.Sprintf("Could not create HAProxy frontend: %s", err),
		)
		return
	}

	result, err := opnsense.Get[frontendAPIResponse](ctx, r.client, frontendReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HAProxy frontend after create",
			fmt.Sprintf("Created frontend %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *frontendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FrontendResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[frontendAPIResponse](ctx, r.client, frontendReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading HAProxy frontend",
			fmt.Sprintf("Could not read HAProxy frontend %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing HAProxy frontend via the OPNsense API.
func (r *frontendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FrontendResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FrontendResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, frontendReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating HAProxy frontend",
			fmt.Sprintf("Could not update HAProxy frontend %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[frontendAPIResponse](ctx, r.client, frontendReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HAProxy frontend after update",
			fmt.Sprintf("Updated frontend %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an HAProxy frontend from the OPNsense API.
func (r *frontendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FrontendResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, frontendReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting HAProxy frontend",
			fmt.Sprintf("Could not delete HAProxy frontend %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing HAProxy frontend by UUID.
func (r *frontendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
