// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &caResource{}
	_ resource.ResourceWithImportState = &caResource{}
)

var caReqOpts = opnsense.ReqOpts{
	AddEndpoint:    "/api/trust/ca/add",
	GetEndpoint:    "/api/trust/ca/get",
	UpdateEndpoint: "/api/trust/ca/set",
	DeleteEndpoint: "/api/trust/ca/del",
	SearchEndpoint: "/api/trust/ca/search",
	Monad:          "ca",
}

type caResource struct{ client *opnsense.Client }

func newCAResource() resource.Resource { return &caResource{} }

func (r *caResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trust_ca"
}

func (r *caResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *caResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CAResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.OperationToken.IsUnknown() {
		resp.Diagnostics.AddError("Invalid Trust operation", errUnknownIntent.Error())
		return
	}
	if !trustConfigKnown(req.Config, &resp.Diagnostics) {
		return
	}
	payload, err := plan.payload()
	if err == nil {
		err = trustAction(payload, "ca", true, true)
	}
	if err != nil {
		resp.Diagnostics.AddError("Invalid Trust operation", err.Error())
		return
	}
	if err = r.client.LockMutex(ctx); err != nil {
		resp.Diagnostics.AddError("Trust lock unavailable", "No mutation attempted.")
		return
	}
	defer r.client.UnlockMutex()
	id, key, err := trustWrite(ctx, r.client, "ca", "add", "", payload)
	if err != nil {
		resp.Diagnostics.AddError("Trust create failed", err.Error())
		return
	}
	plan.OperationStatus = types.StringValue("saved")
	plan.ID = types.StringValue(id)
	_ = key
	// Save confirmed identity and one-time material before fallible readback.
	saved := plan
	saved.retainKnownState(nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, &saved)...)
	record, err := trustRecord(ctx, r.client, "ca", id)
	if err != nil {
		resp.Diagnostics.AddError("Trust object saved; readback failed", "Confirmed UUID and any available one-time key are retained. Refresh without replaying creation.")
		return
	}
	plan.fromRecord(record, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *caResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CAResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	record, err := trustRecord(ctx, r.client, "ca", state.ID.ValueString())
	if err == errTrustMissing {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Trust read failed", err.Error())
		return
	}
	state.fromRecord(record, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *caResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state CAResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.OperationToken.IsUnknown() {
		resp.Diagnostics.AddError("Invalid Trust operation", errUnknownIntent.Error())
		return
	}
	if !trustConfigKnown(req.Config, &resp.Diagnostics) {
		return
	}
	payload, err := plan.payload()
	if err != nil {
		resp.Diagnostics.AddError("Invalid Trust operation", err.Error())
		return
	}
	// Removal/empty intent is never an affirmative cryptographic request.
	explicit := (plan.Action.ValueString() != "" && !plan.Action.Equal(state.Action)) || (plan.OperationToken.ValueString() != "" && !plan.OperationToken.Equal(state.OperationToken))
	if plan.Action.ValueString() == "" && explicit {
		resp.Diagnostics.AddError("Explicit Trust action required", "A new operation_token requires a nonempty action. Omitted action defaults apply only to creation.")
		return
	}
	if !explicit && plan.changedOperationInputs(&state) {
		resp.Diagnostics.AddError("Explicit Trust operation required", "Generation or PEM inputs changed. Change action or operation_token deliberately; description-only updates never perform lifecycle operations.")
		return
	}
	if !explicit && plan.Description.Equal(state.Description) {
		plan.retainKnownState(&state)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
	if err = r.client.LockMutex(ctx); err != nil {
		resp.Diagnostics.AddError("Trust lock unavailable", "No mutation attempted.")
		return
	}
	defer r.client.UnlockMutex()
	id := state.ID.ValueString()
	if !explicit {
		current, readErr := trustRecord(ctx, r.client, "ca", id)
		if readErr != nil {
			resp.Diagnostics.AddError("Trust metadata update blocked", "Cannot read current certificate and private key safely.")
			return
		}
		payload["crt_payload"] = current["crt_payload"]
		payload["prv_payload"] = current["prv_payload"]
	}
	if err = trustAction(payload, "ca", false, explicit); err != nil {
		resp.Diagnostics.AddError("Invalid Trust operation", err.Error())
		return
	}
	// Persist attempted intent, not request echoes, before crossing the
	// non-idempotent boundary. A lost response must not replay on next apply.
	plan.OperationStatus = state.OperationStatus
	if explicit {
		attempted := state
		attempted.Action = plan.Action
		attempted.OperationToken = plan.OperationToken
		attempted.OperationStatus = types.StringValue("unknown")
		resp.Diagnostics.Append(resp.State.Set(ctx, &attempted)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if _, _, err = trustWrite(ctx, r.client, "ca", "set", id, payload); err != nil {
		resp.Diagnostics.AddError("Trust update failed", err.Error()+" Explicit attempted intent is retained as unknown; refresh to reconcile and use a new operation_token only for a deliberate retry.")
		return
	}
	if explicit {
		plan.OperationStatus = types.StringValue("saved")
	}
	plan.ID = state.ID
	saved := plan
	saved.retainKnownState(&state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &saved)...)
	record, err := trustRecord(ctx, r.client, "ca", id)
	if err != nil {
		resp.Diagnostics.AddError("Trust object saved; readback failed", "Managed identity and available key retained; refresh before another explicit operation.")
		return
	}
	plan.fromRecord(record, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *caResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CAResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.LockMutex(ctx); err != nil {
		resp.Diagnostics.AddError("Trust lock unavailable", "No mutation attempted.")
		return
	}
	defer r.client.UnlockMutex()
	if _, _, err := trustWrite(ctx, r.client, "ca", "del", state.ID.ValueString(), nil); err != nil {
		// A committed delete may lose its response, or the object may have
		// disappeared after refresh. Only affirmative absence proves completion.
		probeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if _, readErr := trustRecord(probeCtx, r.client, "ca", state.ID.ValueString()); readErr == errTrustMissing {
			return
		}
		resp.Diagnostics.AddError("Trust delete failed", err.Error()+" Authoritative absence was not confirmed; managed state retained.")
	}
}

func (r *caResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
