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

// Ensure backendResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &backendResource{}
	_ resource.ResourceWithImportState = &backendResource{}
)

// backendReqOpts configures the OPNsense API endpoints for HAProxy backends.
var backendReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/haproxy/settings/addBackend",
	GetEndpoint:         "/api/haproxy/settings/getBackend",
	UpdateEndpoint:      "/api/haproxy/settings/setBackend",
	DeleteEndpoint:      "/api/haproxy/settings/delBackend",
	SearchEndpoint:      "/api/haproxy/settings/searchBackends",
	ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
	Monad:               "backend",
}

// backendResource implements the opnsense_haproxy_backend resource.
type backendResource struct {
	client *opnsense.Client
}

func newBackendResource() resource.Resource {
	return &backendResource{}
}

// Metadata sets the resource type name.
func (r *backendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_backend"
}

// Configure extracts the OPNsense API client from provider data.
func (r *backendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new HAProxy backend via the OPNsense API.
func (r *backendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BackendResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, backendReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating HAProxy backend",
			fmt.Sprintf("Could not create HAProxy backend: %s", err),
		)
		return
	}

	result, err := opnsense.Get[backendAPIResponse](ctx, r.client, backendReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HAProxy backend after create",
			fmt.Sprintf("Created backend %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *backendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BackendResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[backendAPIResponse](ctx, r.client, backendReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading HAProxy backend",
			fmt.Sprintf("Could not read HAProxy backend %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing HAProxy backend via the OPNsense API.
func (r *backendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BackendResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state BackendResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, backendReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating HAProxy backend",
			fmt.Sprintf("Could not update HAProxy backend %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[backendAPIResponse](ctx, r.client, backendReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HAProxy backend after update",
			fmt.Sprintf("Updated backend %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an HAProxy backend from the OPNsense API.
func (r *backendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BackendResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, backendReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting HAProxy backend",
			fmt.Sprintf("Could not delete HAProxy backend %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing HAProxy backend by UUID.
func (r *backendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
