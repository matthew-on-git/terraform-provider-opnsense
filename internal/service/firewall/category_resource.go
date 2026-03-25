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

// Ensure categoryResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &categoryResource{}
	_ resource.ResourceWithImportState = &categoryResource{}
)

// categoryReqOpts configures the OPNsense API endpoints for firewall categories.
// No ReconfigureEndpoint — categories are metadata labels that don't affect running config.
var categoryReqOpts = opnsense.ReqOpts{
	AddEndpoint:    "/api/firewall/category/addItem",
	GetEndpoint:    "/api/firewall/category/getItem",
	UpdateEndpoint: "/api/firewall/category/setItem",
	DeleteEndpoint: "/api/firewall/category/delItem",
	SearchEndpoint: "/api/firewall/category/searchItem",
	Monad:          "category",
}

// categoryResource implements the opnsense_firewall_category resource.
type categoryResource struct {
	client *opnsense.Client
}

func newCategoryResource() resource.Resource {
	return &categoryResource{}
}

// Metadata sets the resource type name.
func (r *categoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_category"
}

// Configure extracts the OPNsense API client from provider data.
func (r *categoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new firewall category via the OPNsense API.
func (r *categoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CategoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)

	uuid, err := opnsense.Add(ctx, r.client, categoryReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating firewall category",
			fmt.Sprintf("Could not create firewall category: %s", err),
		)
		return
	}

	result, err := opnsense.Get[categoryAPIResponse](ctx, r.client, categoryReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading firewall category after create",
			fmt.Sprintf("Created category %s but could not read it back: %s", uuid, err),
		)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *categoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[categoryAPIResponse](ctx, r.client, categoryReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading firewall category",
			fmt.Sprintf("Could not read firewall category %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing firewall category via the OPNsense API.
func (r *categoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CategoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()

	err := opnsense.Update(ctx, r.client, categoryReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating firewall category",
			fmt.Sprintf("Could not update firewall category %s: %s", id, err),
		)
		return
	}

	result, err := opnsense.Get[categoryAPIResponse](ctx, r.client, categoryReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading firewall category after update",
			fmt.Sprintf("Updated category %s but could not read it back: %s", id, err),
		)
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a firewall category from the OPNsense API.
func (r *categoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, categoryReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting firewall category",
			fmt.Sprintf("Could not delete firewall category %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing firewall category by UUID.
func (r *categoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
