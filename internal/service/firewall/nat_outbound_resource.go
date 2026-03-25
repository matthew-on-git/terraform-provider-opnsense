// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure natOutboundResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &natOutboundResource{}
	_ resource.ResourceWithImportState = &natOutboundResource{}
)

// natOutboundReqOpts configures the OPNsense API endpoints for outbound NAT rules.
var natOutboundReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/firewall/source_nat/add_rule",
	GetEndpoint:         "/api/firewall/source_nat/get_rule",
	UpdateEndpoint:      "/api/firewall/source_nat/set_rule",
	DeleteEndpoint:      "/api/firewall/source_nat/del_rule",
	SearchEndpoint:      "/api/firewall/source_nat/search_rule",
	ReconfigureEndpoint: "/api/firewall/source_nat/apply",
	Monad:               "rule",
}

// natOutboundResource implements the opnsense_firewall_nat_outbound resource.
type natOutboundResource struct {
	client *opnsense.Client
}

func newNatOutboundResource() resource.Resource {
	return &natOutboundResource{}
}

// Metadata sets the resource type name.
func (r *natOutboundResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_nat_outbound"
}

// Configure extracts the OPNsense API client from provider data.
func (r *natOutboundResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new outbound NAT rule via the OPNsense API.
func (r *natOutboundResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NatOutboundResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, natOutboundReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating outbound NAT rule",
			fmt.Sprintf("Could not create outbound NAT rule: %s", err),
		)
		return
	}

	result, err := opnsense.Get[natOutboundAPIResponse](ctx, r.client, natOutboundReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading outbound NAT rule after create",
			fmt.Sprintf("Created rule %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *natOutboundResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NatOutboundResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[natOutboundAPIResponse](ctx, r.client, natOutboundReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading outbound NAT rule",
			fmt.Sprintf("Could not read outbound NAT rule %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing outbound NAT rule via the OPNsense API.
func (r *natOutboundResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NatOutboundResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NatOutboundResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, natOutboundReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating outbound NAT rule",
			fmt.Sprintf("Could not update outbound NAT rule %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[natOutboundAPIResponse](ctx, r.client, natOutboundReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading outbound NAT rule after update",
			fmt.Sprintf("Updated rule %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an outbound NAT rule from the OPNsense API.
func (r *natOutboundResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NatOutboundResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, natOutboundReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting outbound NAT rule",
			fmt.Sprintf("Could not delete outbound NAT rule %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing outbound NAT rule by UUID.
func (r *natOutboundResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
