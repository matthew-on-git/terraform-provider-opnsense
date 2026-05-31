// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &generalResource{}
	_ resource.ResourceWithImportState = &generalResource{}
)

// generalID is the synthetic identifier for the FRR general singleton. The
// settings object has no UUID, so a constant id is used for Terraform state.
const generalID = "general"

var generalReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/quagga/general/get",
	UpdateEndpoint:      "/api/quagga/general/set",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "general",
}

type generalResource struct{ client *opnsense.Client }

func newGeneralResource() resource.Resource { return &generalResource{} }

func (r *generalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_general"
}

func (r *generalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	r.client = client
}

// Create applies the settings (the singleton always exists on the appliance,
// so create is implemented as an update + read-back).
func (r *generalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, generalReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error applying FRR general settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[generalAPIResponse](ctx, r.client, generalReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading FRR general settings after apply", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, generalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *generalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GeneralResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.GetSingleton[generalAPIResponse](ctx, r.client, generalReqOpts)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading FRR general settings", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, generalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *generalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, generalReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error updating FRR general settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[generalAPIResponse](ctx, r.client, generalReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading FRR general settings after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, generalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op for the singleton: the appliance always retains an FRR
// general configuration. Destroy only removes the resource from Terraform state.
func (r *generalResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// ImportState sets the fixed singleton id regardless of the supplied value.
func (r *generalResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(generalID))...)
}
