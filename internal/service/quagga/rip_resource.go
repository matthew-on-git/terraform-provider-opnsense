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
	_ resource.Resource                = &ripResource{}
	_ resource.ResourceWithImportState = &ripResource{}
)

const ripID = "rip"

var ripReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/quagga/rip/get",
	UpdateEndpoint:      "/api/quagga/rip/set",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "rip",
}

type ripResource struct{ client *opnsense.Client }

func newRIPResource() resource.Resource { return &ripResource{} }

func (r *ripResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_rip"
}

func (r *ripResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ripResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, ripReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error applying RIP settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[ripAPIResponse](ctx, r.client, ripReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading RIP settings after apply", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, ripID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ripResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.GetSingleton[ripAPIResponse](ctx, r.client, ripReqOpts)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading RIP settings", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, ripID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ripResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, ripReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error updating RIP settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[ripAPIResponse](ctx, r.client, ripReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading RIP settings after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, ripID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ripResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *ripResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(ripID))...)
}
