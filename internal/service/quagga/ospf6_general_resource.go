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
	_ resource.Resource                = &ospf6GeneralResource{}
	_ resource.ResourceWithImportState = &ospf6GeneralResource{}
)

const ospf6GeneralID = "ospf6"

var ospf6GeneralReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/quagga/ospf6settings/get",
	UpdateEndpoint:      "/api/quagga/ospf6settings/set",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "ospf6",
}

type ospf6GeneralResource struct{ client *opnsense.Client }

func newOSPF6GeneralResource() resource.Resource { return &ospf6GeneralResource{} }

func (r *ospf6GeneralResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_general"
}

func (r *ospf6GeneralResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ospf6GeneralResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OSPF6GeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, ospf6GeneralReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error applying OSPFv3 settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[ospf6GeneralAPIResponse](ctx, r.client, ospf6GeneralReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OSPFv3 settings after apply", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, ospf6GeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ospf6GeneralResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OSPF6GeneralResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.GetSingleton[ospf6GeneralAPIResponse](ctx, r.client, ospf6GeneralReqOpts)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading OSPFv3 settings", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, ospf6GeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ospf6GeneralResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OSPF6GeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, ospf6GeneralReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error updating OSPFv3 settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[ospf6GeneralAPIResponse](ctx, r.client, ospf6GeneralReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OSPFv3 settings after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, ospf6GeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ospf6GeneralResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *ospf6GeneralResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(ospf6GeneralID))...)
}
