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

// Ensure aclResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &aclResource{}
	_ resource.ResourceWithImportState = &aclResource{}
)

// aclReqOpts configures the OPNsense API endpoints for HAProxy ACLs.
var aclReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/haproxy/settings/addAcl",
	GetEndpoint:         "/api/haproxy/settings/getAcl",
	UpdateEndpoint:      "/api/haproxy/settings/setAcl",
	DeleteEndpoint:      "/api/haproxy/settings/delAcl",
	SearchEndpoint:      "/api/haproxy/settings/searchAcls",
	ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
	Monad:               "acl",
}

// aclResource implements the opnsense_haproxy_acl resource.
type aclResource struct {
	client *opnsense.Client
}

func newACLResource() resource.Resource {
	return &aclResource{}
}

// Metadata sets the resource type name.
func (r *aclResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_acl"
}

// Configure extracts the OPNsense API client from provider data.
func (r *aclResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new HAProxy ACL via the OPNsense API.
func (r *aclResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ACLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, aclReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating HAProxy ACL",
			fmt.Sprintf("Could not create HAProxy ACL: %s", err),
		)
		return
	}

	result, err := opnsense.Get[aclAPIResponse](ctx, r.client, aclReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HAProxy ACL after create",
			fmt.Sprintf("Created ACL %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *aclResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ACLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[aclAPIResponse](ctx, r.client, aclReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading HAProxy ACL",
			fmt.Sprintf("Could not read HAProxy ACL %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing HAProxy ACL via the OPNsense API.
func (r *aclResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ACLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ACLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, aclReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating HAProxy ACL",
			fmt.Sprintf("Could not update HAProxy ACL %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[aclAPIResponse](ctx, r.client, aclReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HAProxy ACL after update",
			fmt.Sprintf("Updated ACL %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an HAProxy ACL from the OPNsense API.
func (r *aclResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ACLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, aclReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting HAProxy ACL",
			fmt.Sprintf("Could not delete HAProxy ACL %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing HAProxy ACL by UUID.
func (r *aclResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
