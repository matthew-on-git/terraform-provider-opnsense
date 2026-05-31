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
	_ resource.Resource                = &ospfGeneralResource{}
	_ resource.ResourceWithImportState = &ospfGeneralResource{}
)

const ospfGeneralID = "ospf"

var ospfGeneralReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/quagga/ospfsettings/get",
	UpdateEndpoint:      "/api/quagga/ospfsettings/set",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "ospf",
}

type ospfGeneralResource struct{ client *opnsense.Client }

func newOSPFGeneralResource() resource.Resource { return &ospfGeneralResource{} }

func (r *ospfGeneralResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_general"
}

func (r *ospfGeneralResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ospfGeneralResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OSPFGeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, ospfGeneralReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error applying OSPF settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[ospfGeneralAPIResponse](ctx, r.client, ospfGeneralReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OSPF settings after apply", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, ospfGeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ospfGeneralResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OSPFGeneralResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.GetSingleton[ospfGeneralAPIResponse](ctx, r.client, ospfGeneralReqOpts)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading OSPF settings", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, ospfGeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ospfGeneralResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OSPFGeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, ospfGeneralReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error updating OSPF settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[ospfGeneralAPIResponse](ctx, r.client, ospfGeneralReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OSPF settings after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, ospfGeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ospfGeneralResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *ospfGeneralResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(ospfGeneralID))...)
}
