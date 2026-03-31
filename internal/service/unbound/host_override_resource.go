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

// Ensure hostOverrideResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &hostOverrideResource{}
	_ resource.ResourceWithImportState = &hostOverrideResource{}
)

// hostOverrideReqOpts configures the OPNsense API endpoints for Unbound host overrides.
var hostOverrideReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/unbound/settings/add_host_override",
	GetEndpoint:         "/api/unbound/settings/get_host_override",
	UpdateEndpoint:      "/api/unbound/settings/set_host_override",
	DeleteEndpoint:      "/api/unbound/settings/del_host_override",
	SearchEndpoint:      "/api/unbound/settings/search_host_override",
	ReconfigureEndpoint: "/api/unbound/service/reconfigure",
	Monad:               "host_override",
}

// hostOverrideResource implements the opnsense_unbound_host_override resource.
type hostOverrideResource struct {
	client *opnsense.Client
}

func newHostOverrideResource() resource.Resource {
	return &hostOverrideResource{}
}

// Metadata sets the resource type name.
func (r *hostOverrideResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_host_override"
}

// Configure extracts the OPNsense API client from provider data.
func (r *hostOverrideResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new Unbound host override via the OPNsense API.
func (r *hostOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HostOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, hostOverrideReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Unbound host override",
			fmt.Sprintf("Could not create Unbound host override: %s", err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[hostOverrideAPIResponse](ctx, r.client, hostOverrideReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Unbound host override after create",
			fmt.Sprintf("Created host override %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *hostOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HostOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[hostOverrideAPIResponse](ctx, r.client, hostOverrideReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Unbound host override",
			fmt.Sprintf("Could not read Unbound host override %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing Unbound host override via the OPNsense API.
func (r *hostOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan HostOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state HostOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, hostOverrideReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Unbound host override",
			fmt.Sprintf("Could not update Unbound host override %s: %s", id, err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[hostOverrideAPIResponse](ctx, r.client, hostOverrideReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Unbound host override after update",
			fmt.Sprintf("Updated host override %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an Unbound host override from the OPNsense API.
func (r *hostOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state HostOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, hostOverrideReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Unbound host override",
			fmt.Sprintf("Could not delete Unbound host override %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing Unbound host override by UUID.
func (r *hostOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
