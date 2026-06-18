// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

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
	_ resource.Resource                = &settingsResource{}
	_ resource.ResourceWithImportState = &settingsResource{}
)

const settingsID = "ddclient-settings"

var ddclientSettingsReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/dyndns/settings/get",
	UpdateEndpoint:      "/api/dyndns/settings/set",
	ReconfigureEndpoint: "/api/dyndns/service/reconfigure",
	Monad:               "ddclient.general",
}

type settingsResource struct{ client *opnsense.Client }

func newSettingsResource() resource.Resource { return &settingsResource{} }

func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ddclient_settings"
}

func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, ddclientSettingsReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error applying ddclient settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[settingsAPIResponse](ctx, r.client, ddclientSettingsReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ddclient settings after apply", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, settingsID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.GetSingleton[settingsAPIResponse](ctx, r.client, ddclientSettingsReqOpts)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading ddclient settings", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, settingsID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.UpdateSingleton(ctx, r.client, ddclientSettingsReqOpts, plan.toAPI(ctx)); err != nil {
		resp.Diagnostics.AddError("Error updating ddclient settings", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.GetSingleton[settingsAPIResponse](ctx, r.client, ddclientSettingsReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ddclient settings after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, settingsID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *settingsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(settingsID))...)
}
