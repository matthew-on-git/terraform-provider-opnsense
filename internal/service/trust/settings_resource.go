// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ resource.ResourceWithImportState = &settingsResource{}

type settingsResource struct{ client *opnsense.Client }

func newSettingsResource() resource.Resource { return &settingsResource{} }
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trust_settings"
}

func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	r.client = c
}

func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "trust_settings" {
		resp.Diagnostics.AddError("Invalid Trust settings import identity", "Import requires exactly trust_settings.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *settingsResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Import required", "Import the existing singleton with ID trust_settings before applying changes. Create never changes appliance settings.")
}

func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config, prior, plan SettingsResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	pending, d := settingsPending(ctx, req.Private, prior.ID.ValueString())
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(saveSettings(ctx, r.client, config, plan, pending, func(m SettingsResourceModel) diag.Diagnostics {
		return resp.State.Set(context.WithoutCancel(ctx), &m)
	}, resp.Private)...)
}

func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	pending, d := settingsPending(ctx, req.Private, state.ID.ValueString())
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Trust settings client", "The provider client is unavailable.")
		return
	}
	if err := r.client.LockMutex(ctx); err != nil {
		resp.Diagnostics.AddError("Trust settings refresh canceled", err.Error())
		return
	}
	defer r.client.UnlockMutex()
	if err := requireSettings(r.client); err != nil {
		resp.Diagnostics.AddError("Trust settings unavailable", err.Error())
		return
	}
	resp.Diagnostics.Append(recoverSettings(ctx, r.client, pending, func(m SettingsResourceModel) diag.Diagnostics {
		return resp.State.Set(context.WithoutCancel(ctx), &m)
	}, resp.Private)...)
}

// Destroy intentionally abandons even pending activation. No client access,
// capability check, recovery, stop, reset or reconfigure belongs here.
func (r *settingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(context.WithoutCancel(ctx))
}
