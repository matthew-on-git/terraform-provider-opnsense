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
	_ resource.Resource                = &bgpGlobalResource{}
	_ resource.ResourceWithImportState = &bgpGlobalResource{}
)

// bgpGlobalID is the synthetic identifier for the BGP global singleton.
const bgpGlobalID = "bgp"

var bgpGlobalReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/quagga/bgp/get",
	UpdateEndpoint:      "/api/quagga/bgp/set",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "bgp",
}

type bgpGlobalResource struct{ client *opnsense.Client }

func newBGPGlobalResource() resource.Resource { return &bgpGlobalResource{} }

func (r *bgpGlobalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_global"
}

func (r *bgpGlobalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bgpGlobalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BGPGlobalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, bgpGlobalReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error applying BGP global settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[bgpGlobalAPIResponse](ctx, r.client, bgpGlobalReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP global settings after apply", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, bgpGlobalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpGlobalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BGPGlobalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.GetSingleton[bgpGlobalAPIResponse](ctx, r.client, bgpGlobalReqOpts)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading BGP global settings", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, bgpGlobalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bgpGlobalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BGPGlobalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, bgpGlobalReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error updating BGP global settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[bgpGlobalAPIResponse](ctx, r.client, bgpGlobalReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP global settings after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, bgpGlobalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op for the singleton; destroy only removes it from state.
func (r *bgpGlobalResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *bgpGlobalResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(bgpGlobalID))...)
}
