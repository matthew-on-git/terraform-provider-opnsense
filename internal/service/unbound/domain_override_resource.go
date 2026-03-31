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

// Ensure domainOverrideResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &domainOverrideResource{}
	_ resource.ResourceWithImportState = &domainOverrideResource{}
)

// domainOverrideReqOpts configures the OPNsense API endpoints for Unbound domain overrides.
var domainOverrideReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/unbound/settings/add_forward",
	GetEndpoint:         "/api/unbound/settings/get_forward",
	UpdateEndpoint:      "/api/unbound/settings/set_forward",
	DeleteEndpoint:      "/api/unbound/settings/del_forward",
	SearchEndpoint:      "/api/unbound/settings/search_forward",
	ReconfigureEndpoint: "/api/unbound/service/reconfigure",
	Monad:               "forward",
}

// domainOverrideResource implements the opnsense_unbound_domain_override resource.
type domainOverrideResource struct {
	client *opnsense.Client
}

func newDomainOverrideResource() resource.Resource {
	return &domainOverrideResource{}
}

// Metadata sets the resource type name.
func (r *domainOverrideResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_domain_override"
}

// Configure extracts the OPNsense API client from provider data.
func (r *domainOverrideResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new Unbound domain override via the OPNsense API.
func (r *domainOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, domainOverrideReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Unbound domain override",
			fmt.Sprintf("Could not create Unbound domain override: %s", err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[domainOverrideAPIResponse](ctx, r.client, domainOverrideReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Unbound domain override after create",
			fmt.Sprintf("Created domain override %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *domainOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[domainOverrideAPIResponse](ctx, r.client, domainOverrideReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Unbound domain override",
			fmt.Sprintf("Could not read Unbound domain override %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing Unbound domain override via the OPNsense API.
func (r *domainOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DomainOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, domainOverrideReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Unbound domain override",
			fmt.Sprintf("Could not update Unbound domain override %s: %s", id, err),
		)
		return
	}

	// Read back from API to populate state (never echo from config).
	result, err := opnsense.Get[domainOverrideAPIResponse](ctx, r.client, domainOverrideReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Unbound domain override after update",
			fmt.Sprintf("Updated domain override %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an Unbound domain override from the OPNsense API.
func (r *domainOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, domainOverrideReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Unbound domain override",
			fmt.Sprintf("Could not delete Unbound domain override %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing Unbound domain override by UUID.
func (r *domainOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
