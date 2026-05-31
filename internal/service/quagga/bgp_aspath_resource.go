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
	_ resource.Resource                = &bgpASPathResource{}
	_ resource.ResourceWithImportState = &bgpASPathResource{}
)

var bgpASPathReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/quagga/bgp/add_aspath",
	GetEndpoint:         "/api/quagga/bgp/get_aspath",
	UpdateEndpoint:      "/api/quagga/bgp/set_aspath",
	DeleteEndpoint:      "/api/quagga/bgp/del_aspath",
	SearchEndpoint:      "/api/quagga/bgp/search_aspath",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "aspath",
}

type bgpASPathResource struct{ client *opnsense.Client }

func newBGPASPathResource() resource.Resource { return &bgpASPathResource{} }

func (r *bgpASPathResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_aspath"
}

func (r *bgpASPathResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bgpASPathResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BGPASPathResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, bgpASPathReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating BGP AS-path", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[bgpASPathAPIResponse](ctx, r.client, bgpASPathReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP AS-path after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpASPathResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BGPASPathResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[bgpASPathAPIResponse](ctx, r.client, bgpASPathReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading BGP AS-path", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bgpASPathResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BGPASPathResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, bgpASPathReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating BGP AS-path", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[bgpASPathAPIResponse](ctx, r.client, bgpASPathReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP AS-path after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bgpASPathResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BGPASPathResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, bgpASPathReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting BGP AS-path", fmt.Sprintf("%s", err))
	}
}

func (r *bgpASPathResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
