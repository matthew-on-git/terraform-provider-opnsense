// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure natPortForwardResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &natPortForwardResource{}
	_ resource.ResourceWithImportState = &natPortForwardResource{}
)

// natPortForwardReqOpts configures the OPNsense API endpoints for NAT port-forward rules.
// Uses standard ReconfigureEndpoint (apply without savepoint).
var natPortForwardReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/firewall/d_nat/add_rule",
	GetEndpoint:         "/api/firewall/d_nat/get_rule",
	UpdateEndpoint:      "/api/firewall/d_nat/set_rule",
	DeleteEndpoint:      "/api/firewall/d_nat/del_rule",
	SearchEndpoint:      "/api/firewall/d_nat/search_rule",
	ReconfigureEndpoint: "/api/firewall/d_nat/apply",
	Monad:               "rule",
}

// natPortForwardResource implements the opnsense_firewall_nat_port_forward resource.
type natPortForwardResource struct {
	client *opnsense.Client
}

func newNatPortForwardResource() resource.Resource {
	return &natPortForwardResource{}
}

// Metadata sets the resource type name.
func (r *natPortForwardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_nat_port_forward"
}

// Configure extracts the OPNsense API client from provider data.
func (r *natPortForwardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			"Expected *opnsense.Client, got something else.",
		)
		return
	}
	r.client = client
}

// Create creates a new NAT port-forward rule via the OPNsense API.
func (r *natPortForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NatPortForwardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, err := r.prepareAPI(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving NAT port-forward categories", err.Error())
		return
	}
	plannedInterface := plan.Interface

	uuid, err := opnsense.Add(ctx, r.client, natPortForwardReqOpts, apiReq)
	if err != nil {
		if uuid != "" {
			recovered := false
			if result, readErr := opnsense.Get[natPortForwardAPIResponse](ctx, r.client, natPortForwardReqOpts, uuid); readErr == nil {
				apiDiags := validateNATAPIResponse(result)
				resp.Diagnostics.Append(apiDiags...)
				if !apiDiags.HasError() {
					plan.fromAPI(ctx, result, uuid)
					if setContains(ctx, plan.Interfaces, plannedInterface.ValueString()) {
						plan.Interface = plannedInterface
					}
					resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
					recovered = true
				}
			}
			if !recovered {
				identity := NatPortForwardResourceModel{ID: types.StringValue(uuid), Interfaces: types.SetNull(types.StringType), Categories: types.SetNull(types.StringType)}
				resp.Diagnostics.Append(resp.State.Set(ctx, &identity)...)
			}
		}
		resp.Diagnostics.AddError(
			"Error creating NAT port-forward rule",
			fmt.Sprintf("Could not create NAT port-forward rule: %s", err),
		)
		return
	}

	result, err := opnsense.Get[natPortForwardAPIResponse](ctx, r.client, natPortForwardReqOpts, uuid)
	if err != nil {
		identity := NatPortForwardResourceModel{ID: types.StringValue(uuid), Interfaces: types.SetNull(types.StringType), Categories: types.SetNull(types.StringType)}
		resp.Diagnostics.Append(resp.State.Set(ctx, &identity)...)
		resp.Diagnostics.AddError(
			"Error reading NAT port-forward rule after create",
			fmt.Sprintf("Created rule %s but could not read it back: %s", uuid, err),
		)
		return
	}
	apiDiags := validateNATAPIResponse(result)
	if apiDiags.HasError() {
		identity := NatPortForwardResourceModel{ID: types.StringValue(uuid), Interfaces: types.SetNull(types.StringType), Categories: types.SetNull(types.StringType)}
		resp.Diagnostics.Append(resp.State.Set(ctx, &identity)...)
		resp.Diagnostics.Append(apiDiags...)
		return
	}

	plan.fromAPI(ctx, result, uuid)
	if setContains(ctx, plan.Interfaces, plannedInterface.ValueString()) {
		plan.Interface = plannedInterface
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the OPNsense API.
func (r *natPortForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NatPortForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := opnsense.Get[natPortForwardAPIResponse](ctx, r.client, natPortForwardReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading NAT port-forward rule",
			fmt.Sprintf("Could not read NAT port-forward rule %s: %s", state.ID.ValueString(), err),
		)
		return
	}
	resp.Diagnostics.Append(validateNATAPIResponse(result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	priorInterface := state.Interface
	state.fromAPI(ctx, result, state.ID.ValueString())
	if setContains(ctx, state.Interfaces, priorInterface.ValueString()) {
		state.Interface = priorInterface
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing NAT port-forward rule via the OPNsense API.
func (r *natPortForwardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NatPortForwardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NatPortForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, err := r.prepareAPI(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving NAT port-forward categories", err.Error())
		return
	}
	plannedInterface := plan.Interface
	id := state.ID.ValueString()

	err = opnsense.Update(ctx, r.client, natPortForwardReqOpts, apiReq, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating NAT port-forward rule",
			fmt.Sprintf("Could not update NAT port-forward rule %s: %s", id, err),
		)
		return
	}
	result, err := opnsense.Get[natPortForwardAPIResponse](ctx, r.client, natPortForwardReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading NAT port-forward rule after update",
			fmt.Sprintf("Updated rule %s but could not read it back: %s", id, err),
		)
		return
	}
	resp.Diagnostics.Append(validateNATAPIResponse(result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.fromAPI(ctx, result, id)
	if setContains(ctx, plan.Interfaces, plannedInterface.ValueString()) {
		plan.Interface = plannedInterface
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a NAT port-forward rule from the OPNsense API.
func (r *natPortForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NatPortForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := opnsense.Delete(ctx, r.client, natPortForwardReqOpts, state.ID.ValueString())
	if err != nil {
		var notFoundErr *opnsense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting NAT port-forward rule",
			fmt.Sprintf("Could not delete NAT port-forward rule %s: %s", state.ID.ValueString(), err),
		)
	}
}

// ImportState imports an existing NAT port-forward rule by UUID.
func (r *natPortForwardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// prepareAPI retains UUIDs in Terraform while resolving names for DNat writes.
// OPNsense's plural "categories" is a volatile read mapper: writing UUIDs there
// is silently ignored. The persisted singular "category" field accepts names.
func (r *natPortForwardResource) prepareAPI(ctx context.Context, plan *NatPortForwardResourceModel) (*natPortForwardAPIRequest, error) {
	request := plan.toAPI(ctx)
	// Optional/computed categories may be unknown when omitted. Preserve remote
	// labels in that case; an explicitly empty set must instead clear them.
	if plan.Categories.IsNull() || plan.Categories.IsUnknown() {
		return request, nil
	}
	var ids []string
	if diags := plan.Categories.ElementsAs(ctx, &ids, false); diags.HasError() {
		return nil, fmt.Errorf("category UUIDs must be known strings: %v", diags)
	}
	sort.Strings(ids)
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		category, err := opnsense.Get[categoryAPIResponse](ctx, r.client, categoryReqOpts, id)
		if err != nil {
			return nil, fmt.Errorf("resolve category %q before writing NAT rule: %w", id, err)
		}
		if strings.TrimSpace(category.Name) == "" || strings.Contains(category.Name, ",") {
			return nil, fmt.Errorf("category %q has an empty or comma-containing name that cannot be serialized", id)
		}
		names = append(names, category.Name)
	}
	sort.Strings(names)
	value := strings.Join(names, ",")
	request.Category = &value
	return request, nil
}
