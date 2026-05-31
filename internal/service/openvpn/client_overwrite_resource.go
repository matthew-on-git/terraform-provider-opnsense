// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &clientOverwriteResource{}
	_ resource.ResourceWithImportState = &clientOverwriteResource{}
)

var clientOverwriteReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/openvpn/client_overwrites/add",
	GetEndpoint:         "/api/openvpn/client_overwrites/get",
	UpdateEndpoint:      "/api/openvpn/client_overwrites/set",
	DeleteEndpoint:      "/api/openvpn/client_overwrites/del",
	SearchEndpoint:      "/api/openvpn/client_overwrites/search",
	ReconfigureEndpoint: "/api/openvpn/service/reconfigure",
	Monad:               "overwrite",
}

type clientOverwriteResource struct{ client *opnsense.Client }

func newClientOverwriteResource() resource.Resource { return &clientOverwriteResource{} }

func (r *clientOverwriteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openvpn_client_overwrite"
}

func (r *clientOverwriteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *clientOverwriteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClientOverwriteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, clientOverwriteReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating OpenVPN client overwrite", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[clientOverwriteAPIResponse](ctx, r.client, clientOverwriteReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OpenVPN client overwrite after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientOverwriteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClientOverwriteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[clientOverwriteAPIResponse](ctx, r.client, clientOverwriteReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading OpenVPN client overwrite", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clientOverwriteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ClientOverwriteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, clientOverwriteReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating OpenVPN client overwrite", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[clientOverwriteAPIResponse](ctx, r.client, clientOverwriteReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OpenVPN client overwrite after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientOverwriteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClientOverwriteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, clientOverwriteReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting OpenVPN client overwrite", fmt.Sprintf("%s", err))
	}
}

func (r *clientOverwriteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
