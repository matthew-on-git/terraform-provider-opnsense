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
	_ resource.Resource                = &bgpPeerGroupResource{}
	_ resource.ResourceWithImportState = &bgpPeerGroupResource{}
)

var bgpPeerGroupReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/quagga/bgp/add_peergroup",
	GetEndpoint:         "/api/quagga/bgp/get_peergroup",
	UpdateEndpoint:      "/api/quagga/bgp/set_peergroup",
	DeleteEndpoint:      "/api/quagga/bgp/del_peergroup",
	SearchEndpoint:      "/api/quagga/bgp/search_peergroup",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "peergroup",
}

type bgpPeerGroupResource struct{ client *opnsense.Client }

func newBGPPeerGroupResource() resource.Resource { return &bgpPeerGroupResource{} }

func (r *bgpPeerGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_peergroup"
}

func (r *bgpPeerGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bgpPeerGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BGPPeerGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, bgpPeerGroupReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating BGP peer group", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[bgpPeerGroupAPIResponse](ctx, r.client, bgpPeerGroupReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP peer group after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpPeerGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BGPPeerGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[bgpPeerGroupAPIResponse](ctx, r.client, bgpPeerGroupReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading BGP peer group", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bgpPeerGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BGPPeerGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, bgpPeerGroupReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating BGP peer group", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[bgpPeerGroupAPIResponse](ctx, r.client, bgpPeerGroupReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP peer group after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpPeerGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BGPPeerGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, bgpPeerGroupReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting BGP peer group", fmt.Sprintf("%s", err))
	}
}

func (r *bgpPeerGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
