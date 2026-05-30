// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &bgpRedistributionResource{}
	_ resource.ResourceWithImportState = &bgpRedistributionResource{}
)

var bgpRedistributionReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/quagga/bgp/add_redistribution",
	GetEndpoint:         "/api/quagga/bgp/get_redistribution",
	UpdateEndpoint:      "/api/quagga/bgp/set_redistribution",
	DeleteEndpoint:      "/api/quagga/bgp/del_redistribution",
	SearchEndpoint:      "/api/quagga/bgp/search_redistribution",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "redistribution",
}

type bgpRedistributionResource struct{ client *opnsense.Client }

func newBGPRedistributionResource() resource.Resource { return &bgpRedistributionResource{} }

func (r *bgpRedistributionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_redistribution"
}

func (r *bgpRedistributionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bgpRedistributionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BGPRedistributionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, bgpRedistributionReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating BGP redistribution", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[bgpRedistributionAPIResponse](ctx, r.client, bgpRedistributionReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP redistribution after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpRedistributionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BGPRedistributionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[bgpRedistributionAPIResponse](ctx, r.client, bgpRedistributionReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading BGP redistribution", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bgpRedistributionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BGPRedistributionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, bgpRedistributionReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating BGP redistribution", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[bgpRedistributionAPIResponse](ctx, r.client, bgpRedistributionReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP redistribution after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpRedistributionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BGPRedistributionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, bgpRedistributionReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting BGP redistribution", fmt.Sprintf("%s", err))
	}
}

func (r *bgpRedistributionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
