// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"errors"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &bgpGlobalResource{}
	_ resource.ResourceWithImportState = &bgpGlobalResource{}
)

// bgpGlobalID is the synthetic identifier for the BGP global singleton.
const bgpGlobalID = "bgp"

var bgpGlobalReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/quagga/bgp/get",
	UpdateEndpoint:      "/api/quagga/bgp/set",
	ReconfigureEndpoint: "/api/quagga/service/reconfigure",
	Monad:               "bgp",
}

type bgpGlobalResource struct{ client *opnsense.Client }

func newBGPGlobalResource() resource.Resource { return &bgpGlobalResource{} }

func (r *bgpGlobalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_global"
}

func (r *bgpGlobalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bgpGlobalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if err := bgpOwnerMu.Acquire(ctx, 1); err != nil {
		resp.Diagnostics.AddError("FRR operation canceled while waiting", frrDiagnostic(err))
		return
	}
	defer bgpOwnerMu.Release(1)
	if _, err := discoverFRR(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("FRR compatibility check failed", frrDiagnostic(err))
		return
	}
	pending, d := pendingBGP(ctx, resp.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pending != "" {
		if err := opnsense.RetryReconfigure(ctx, r.client, strictBGPOpts(bgpGlobalReqOpts)); err != nil {
			resp.Diagnostics.AddError("BGP activation pending", frrDiagnostic(err))
			return
		}
		resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
	}

	var plan BGPGlobalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := resolveBGPModel(ctx, req.Config, &plan); err != nil {
		resp.Diagnostics.AddError("Invalid FRR settings intent", frrDiagnostic(err))
		return
	}
	if err := saveFRRSingleton(ctx, r.client, bgpGlobalReqOpts, plan.toAPI(ctx)); err != nil {
		var saved *opnsense.MutationReconfigureError
		var validation *opnsense.ValidationError
		if !errors.As(err, &validation) {
			resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
		}
		if errors.As(err, &saved) {
			plan.ID = types.StringValue(bgpGlobalID)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
			// A confirmed saved singleton must not be tainted by Core and
			// rewritten on retry. Retain its pending marker for refresh.
			resp.Diagnostics.AddWarning("Saved BGP global settings; activation pending", frrDiagnostic(err))
			return
		}
		resp.Diagnostics.AddError("Error applying BGP global settings", frrDiagnostic(err))
		return
	}
	plan.ID = types.StringValue(bgpGlobalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
	result, err := getFRRSingleton[bgpGlobalAPIResponse](ctx, r.client, bgpGlobalReqOpts)
	if err != nil {
		resp.Diagnostics.AddWarning("Saved BGP global settings; read-back pending", frrDiagnostic(err))
		return
	}
	if err := validateBGPResponse(result, frrCapability{}); err != nil {
		resp.Diagnostics.AddWarning("Saved FRR singleton; invalid read-back", frrDiagnostic(err))
		return
	}
	plan.fromAPI(ctx, result, bgpGlobalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
}

func (r *bgpGlobalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if err := bgpOwnerMu.Acquire(ctx, 1); err != nil {
		resp.Diagnostics.AddError("FRR operation canceled while waiting", frrDiagnostic(err))
		return
	}
	defer bgpOwnerMu.Release(1)
	if _, err := discoverFRR(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("FRR compatibility check failed", frrDiagnostic(err))
		return
	}
	pending, d := pendingBGP(ctx, resp.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pending != "" {
		if err := opnsense.RetryReconfigure(ctx, r.client, strictBGPOpts(bgpGlobalReqOpts)); err != nil {
			resp.Diagnostics.AddError("BGP activation pending", frrDiagnostic(err))
			return
		}
		resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
	}

	var state BGPGlobalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := getFRRSingleton[bgpGlobalAPIResponse](ctx, r.client, bgpGlobalReqOpts)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading BGP global settings", frrDiagnostic(err))
		return
	}
	if err := validateBGPResponse(result, frrCapability{}); err != nil {
		resp.Diagnostics.AddError("Invalid FRR singleton read-back", frrDiagnostic(err))
		return
	}
	state.fromAPI(ctx, result, bgpGlobalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bgpGlobalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if err := bgpOwnerMu.Acquire(ctx, 1); err != nil {
		resp.Diagnostics.AddError("FRR operation canceled while waiting", frrDiagnostic(err))
		return
	}
	defer bgpOwnerMu.Release(1)
	if _, err := discoverFRR(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("FRR compatibility check failed", frrDiagnostic(err))
		return
	}
	pending, d := pendingBGP(ctx, resp.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pending != "" {
		if err := opnsense.RetryReconfigure(ctx, r.client, strictBGPOpts(bgpGlobalReqOpts)); err != nil {
			resp.Diagnostics.AddError("BGP activation pending", frrDiagnostic(err))
			return
		}
		resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
	}

	var plan BGPGlobalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var prior BGPGlobalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := resolveBGPModel(ctx, req.Config, &plan, &prior); err != nil {
		resp.Diagnostics.AddError("Invalid FRR settings intent", frrDiagnostic(err))
		return
	}
	var mutationErr error
	if pending != `"mutation"` || !reflect.DeepEqual(plan, prior) {
		mutationErr = saveFRRSingleton(ctx, r.client, bgpGlobalReqOpts, plan.toAPI(ctx))
	}
	if err := mutationErr; err != nil {
		var saved *opnsense.MutationReconfigureError
		var validation *opnsense.ValidationError
		if !errors.As(err, &validation) {
			resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
		}
		if errors.As(err, &saved) {
			plan.ID = types.StringValue(bgpGlobalID)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
		}
		resp.Diagnostics.AddError("Error updating BGP global settings", frrDiagnostic(err))
		return
	}
	plan.ID = types.StringValue(bgpGlobalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
	result, err := getFRRSingleton[bgpGlobalAPIResponse](ctx, r.client, bgpGlobalReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP global settings after update", frrDiagnostic(err))
		return
	}
	if err := validateBGPResponse(result, frrCapability{}); err != nil {
		resp.Diagnostics.AddError("Invalid FRR singleton read-back", frrDiagnostic(err))
		return
	}
	plan.fromAPI(ctx, result, bgpGlobalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
}

// Delete is a no-op for the singleton; destroy only removes it from state.
func (r *bgpGlobalResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *bgpGlobalResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(bgpGlobalID))...)
}
