// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &healthcheckResource{}
	_ resource.ResourceWithImportState = &healthcheckResource{}
)

var healthcheckReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/haproxy/settings/addHealthcheck",
	GetEndpoint:         "/api/haproxy/settings/getHealthcheck",
	UpdateEndpoint:      "/api/haproxy/settings/setHealthcheck",
	DeleteEndpoint:      "/api/haproxy/settings/delHealthcheck",
	SearchEndpoint:      "/api/haproxy/settings/searchHealthchecks",
	ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
	Monad:               "healthcheck",
}

type healthcheckResource struct {
	client *opnsense.Client
}

func newHealthcheckResource() resource.Resource {
	return &healthcheckResource{}
}

func (r *healthcheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_healthcheck"
}

func (r *healthcheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client, got something else.")
		return
	}
	r.client = client
}

func (r *healthcheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HealthcheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	uuid, err := opnsense.Add(ctx, r.client, healthcheckReqOpts, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating HAProxy health check", fmt.Sprintf("Could not create health check: %s", err))
		return
	}

	result, err := opnsense.Get[healthcheckAPIResponse](ctx, r.client, healthcheckReqOpts, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy health check after create", fmt.Sprintf("Created health check %s but could not read it back: %s", uuid, err))
		return
	}

	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *healthcheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HealthcheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[healthcheckAPIResponse](ctx, r.client, healthcheckReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading HAProxy health check", fmt.Sprintf("Could not read health check %s: %s", state.ID.ValueString(), err))
		return
	}

	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *healthcheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan HealthcheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state HealthcheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.toAPI(ctx)
	id := state.ID.ValueString()
	err := opnsense.Update(ctx, r.client, healthcheckReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError("Error updating HAProxy health check", fmt.Sprintf("Could not update health check %s: %s", id, err))
		return
	}

	result, err := opnsense.Get[healthcheckAPIResponse](ctx, r.client, healthcheckReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy health check after update", fmt.Sprintf("Updated health check %s but could not read it back: %s", id, err))
		return
	}

	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *healthcheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state HealthcheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, healthcheckReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError("Error deleting HAProxy health check", fmt.Sprintf("Could not delete health check %s: %s", state.ID.ValueString(), err))
	}
}

func (r *healthcheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
