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
	_ resource.Resource                = &bgpNeighborResource{}
	_ resource.ResourceWithImportState = &bgpNeighborResource{}
)

var bgpNeighborReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/quagga/bgp/add_neighbor",
	GetEndpoint:         "/api/quagga/bgp/get_neighbor",
	UpdateEndpoint:      "/api/quagga/bgp/set_neighbor",
	DeleteEndpoint:      "/api/quagga/bgp/del_neighbor",
	SearchEndpoint:      "/api/quagga/bgp/search_neighbor",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "neighbor",
}

type bgpNeighborResource struct{ client *opnsense.Client }

func newBGPNeighborResource() resource.Resource { return &bgpNeighborResource{} }

func (r *bgpNeighborResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_neighbor"
}

func (r *bgpNeighborResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bgpNeighborResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BGPNeighborResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, bgpNeighborReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating BGP neighbor", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[bgpNeighborAPIResponse](ctx, r.client, bgpNeighborReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP neighbor after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpNeighborResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BGPNeighborResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[bgpNeighborAPIResponse](ctx, r.client, bgpNeighborReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading BGP neighbor", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bgpNeighborResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BGPNeighborResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, bgpNeighborReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating BGP neighbor", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[bgpNeighborAPIResponse](ctx, r.client, bgpNeighborReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP neighbor after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpNeighborResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BGPNeighborResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, bgpNeighborReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting BGP neighbor", fmt.Sprintf("%s", err))
	}
}

func (r *bgpNeighborResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
