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

// Ensure connectionResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &connectionResource{}
	_ resource.ResourceWithImportState = &connectionResource{}
)

// connectionReqOpts configures the OPNsense API endpoints for IPsec connections.
var connectionReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/ipsec/connections/add_connection",
	GetEndpoint:         "/api/ipsec/connections/get_connection",
	UpdateEndpoint:      "/api/ipsec/connections/set_connection",
	DeleteEndpoint:      "/api/ipsec/connections/del_connection",
	SearchEndpoint:      "/api/ipsec/connections/search_connection",
	ReconfigureEndpoint: "/api/ipsec/service/reconfigure",
	Monad:               "connection",
}

// connectionResource implements the opnsense_ipsec_connection resource.
type connectionResource struct {
	client *opnsense.Client
}

func newConnectionResource() resource.Resource {
	return &connectionResource{}
}

// Metadata sets the resource type name.
func (r *connectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_connection"
}

// Configure extracts the OPNsense API client from provider data.
func (r *connectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new IPsec connection via the OPNsense API.
func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, connectionReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IPsec connection",
			fmt.Sprintf("Could not create IPsec connection: %s", err),
		)
		return
	}

	result, err := opnsense.Get[ipsecConnectionAPIResponse](ctx, r.client, connectionReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec connection after create",
			fmt.Sprintf("Created connection %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[ipsecConnectionAPIResponse](ctx, r.client, connectionReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading IPsec connection",
			fmt.Sprintf("Could not read IPsec connection %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing IPsec connection via the OPNsense API.
func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, connectionReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IPsec connection",
			fmt.Sprintf("Could not update IPsec connection %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[ipsecConnectionAPIResponse](ctx, r.client, connectionReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IPsec connection after update",
			fmt.Sprintf("Updated connection %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an IPsec connection from the OPNsense API.
func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, connectionReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting IPsec connection",
			fmt.Sprintf("Could not delete IPsec connection %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing IPsec connection by UUID.
func (r *connectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
