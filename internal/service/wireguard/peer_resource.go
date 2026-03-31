// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure peerResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &peerResource{}
	_ resource.ResourceWithImportState = &peerResource{}
)

// peerReqOpts configures the OPNsense API endpoints for WireGuard peers (clients).
var peerReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/wireguard/client/add_client",
	GetEndpoint:         "/api/wireguard/client/get_client",
	UpdateEndpoint:      "/api/wireguard/client/set_client",
	DeleteEndpoint:      "/api/wireguard/client/del_client",
	SearchEndpoint:      "/api/wireguard/client/search_client",
	ReconfigureEndpoint: "/api/wireguard/service/reconfigure",
	Monad:               "client",
}

// peerResource implements the opnsense_wireguard_peer resource.
type peerResource struct {
	client *opnsense.Client
}

func newPeerResource() resource.Resource {
	return &peerResource{}
}

// Metadata sets the resource type name.
func (r *peerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard_peer"
}

// Configure extracts the OPNsense API client from provider data.
func (r *peerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new WireGuard peer via the OPNsense API.
func (r *peerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PeerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, peerReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating WireGuard peer",
			fmt.Sprintf("Could not create WireGuard peer: %s", err),
		)
		return
	}

	result, err := opnsense.Get[wireguardPeerAPIResponse](ctx, r.client, peerReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading WireGuard peer after create",
			fmt.Sprintf("Created peer %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *peerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[wireguardPeerAPIResponse](ctx, r.client, peerReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading WireGuard peer",
			fmt.Sprintf("Could not read WireGuard peer %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing WireGuard peer via the OPNsense API.
func (r *peerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PeerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, peerReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating WireGuard peer",
			fmt.Sprintf("Could not update WireGuard peer %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[wireguardPeerAPIResponse](ctx, r.client, peerReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading WireGuard peer after update",
			fmt.Sprintf("Updated peer %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a WireGuard peer from the OPNsense API.
func (r *peerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, peerReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting WireGuard peer",
			fmt.Sprintf("Could not delete WireGuard peer %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing WireGuard peer by UUID.
func (r *peerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
