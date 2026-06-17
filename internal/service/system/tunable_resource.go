// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &tunableResource{}
	_ resource.ResourceWithImportState = &tunableResource{}
)

var tunableReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/core/tunables/add_item",
	GetEndpoint:         "/api/core/tunables/get_item",
	UpdateEndpoint:      "/api/core/tunables/set_item",
	DeleteEndpoint:      "/api/core/tunables/del_item",
	SearchEndpoint:      "/api/core/tunables/search_item",
	ReconfigureEndpoint: "/api/core/tunables/reconfigure",
	Monad:               "sysctl",
}

type tunableResource struct{ client *opnsense.Client }

func newTunableResource() resource.Resource { return &tunableResource{} }

func (r *tunableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_tunable"
}

func (r *tunableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tunableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TunableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, tunableReqOpts, plan.toAPI(ctx))
	if err != nil {
		var mutationErr *opnsense.MutationReconfigureError
		if errors.As(err, &mutationErr) && uuid != "" {
			if result, readErr := opnsense.Get[tunableAPIResponse](ctx, r.client, tunableReqOpts, uuid); readErr == nil {
				plan.fromAPI(ctx, result, uuid)
				resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			}
			resp.Diagnostics.AddError(
				"Error applying system tunable after create",
				fmt.Sprintf("%s\n\nThe tunable was saved with ID %q, but OPNsense reconfigure failed. Re-run apply after fixing the OPNsense reconfigure failure, or import/remove the saved tunable if Terraform state was not persisted.", err, uuid),
			)
			return
		}
		resp.Diagnostics.AddError("Error creating system tunable", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[tunableAPIResponse](ctx, r.client, tunableReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading system tunable after create", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tunableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TunableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[tunableAPIResponse](ctx, r.client, tunableReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading system tunable", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tunableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state TunableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if err := opnsense.Update(ctx, r.client, tunableReqOpts, plan.toAPI(ctx), id); err != nil {
		resp.Diagnostics.AddError("Error updating system tunable", fmt.Sprintf("%s", err))
		return
	}
	result, err := opnsense.Get[tunableAPIResponse](ctx, r.client, tunableReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading system tunable after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tunableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TunableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, tunableReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting system tunable", fmt.Sprintf("%s", err))
	}
}

func (r *tunableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
