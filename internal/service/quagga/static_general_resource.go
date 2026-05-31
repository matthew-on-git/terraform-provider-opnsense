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
	_ resource.Resource                = &staticGeneralResource{}
	_ resource.ResourceWithImportState = &staticGeneralResource{}
)

const staticGeneralID = "static"

var staticGeneralReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/quagga/static/get",
	UpdateEndpoint:      "/api/quagga/static/set",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "static",
}

type staticGeneralResource struct{ client *opnsense.Client }

func newStaticGeneralResource() resource.Resource { return &staticGeneralResource{} }

func (r *staticGeneralResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_static"
}

func (r *staticGeneralResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *staticGeneralResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StaticGeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, staticGeneralReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error applying static routing settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[staticGeneralAPIResponse](ctx, r.client, staticGeneralReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading static routing settings after apply", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, staticGeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *staticGeneralResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StaticGeneralResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.GetSingleton[staticGeneralAPIResponse](ctx, r.client, staticGeneralReqOpts)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading static routing settings", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, staticGeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *staticGeneralResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StaticGeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, staticGeneralReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error updating static routing settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[staticGeneralAPIResponse](ctx, r.client, staticGeneralReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading static routing settings after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, staticGeneralID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *staticGeneralResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *staticGeneralResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(staticGeneralID))...)
}
