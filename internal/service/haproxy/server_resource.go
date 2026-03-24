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

// Ensure serverResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &serverResource{}
	_ resource.ResourceWithImportState = &serverResource{}
)

// serverReqOpts configures the OPNsense API endpoints for HAProxy servers.
var serverReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/haproxy/settings/addServer",
	GetEndpoint:         "/api/haproxy/settings/getServer",
	UpdateEndpoint:      "/api/haproxy/settings/setServer",
	DeleteEndpoint:      "/api/haproxy/settings/delServer",
	SearchEndpoint:      "/api/haproxy/settings/searchServers",
	ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
	Monad:               "server",
}

// serverResource implements the opnsense_haproxy_server resource.
type serverResource struct {
	client *opnsense.Client
}

func newServerResource() resource.Resource {
	return &serverResource{}
}

// Metadata sets the resource type name.
func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_server"
}

// Configure extracts the OPNsense API client from provider data.
func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new HAProxy server via the OPNsense API.
func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, serverReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating HAProxy server",
			fmt.Sprintf("Could not create HAProxy server: %s", err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[serverAPIResponse](ctx, r.client, serverReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HAProxy server after create",
			fmt.Sprintf("Created server %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[serverAPIResponse](ctx, r.client, serverReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading HAProxy server",
			fmt.Sprintf("Could not read HAProxy server %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing HAProxy server via the OPNsense API.
func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, serverReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating HAProxy server",
			fmt.Sprintf("Could not update HAProxy server %s: %s", id, err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[serverAPIResponse](ctx, r.client, serverReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HAProxy server after update",
			fmt.Sprintf("Updated server %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an HAProxy server from the OPNsense API.
func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, serverReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting HAProxy server",
			fmt.Sprintf("Could not delete HAProxy server %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing HAProxy server by UUID.
func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
