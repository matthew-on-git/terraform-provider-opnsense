// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithUpgradeState = &frontendResource{}

type frontendResourceModelV0 struct {
	ID                 types.String `tfsdk:"id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Bind               types.String `tfsdk:"bind"`
	Mode               types.String `tfsdk:"mode"`
	DefaultBackend     types.String `tfsdk:"default_backend"`
	SSLEnabled         types.Bool   `tfsdk:"ssl_enabled"`
	Certificates       types.Set    `tfsdk:"certificates"`
	DefaultCertificate types.String `tfsdk:"default_certificate"`
	LinkedActions      types.Set    `tfsdk:"linked_actions"`
	ForwardFor         types.Bool   `tfsdk:"forward_for"`
	TimeoutClient      types.String `tfsdk:"timeout_client"`
}

func (r *frontendResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	priorSchema := frontendSchemaV0()

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &priorSchema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior frontendResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgraded := FrontendResourceModel{
					ID:                 prior.ID,
					Enabled:            prior.Enabled,
					Name:               prior.Name,
					Description:        prior.Description,
					Bind:               prior.Bind,
					Mode:               prior.Mode,
					DefaultBackend:     prior.DefaultBackend,
					SSLEnabled:         prior.SSLEnabled,
					Certificates:       prior.Certificates,
					DefaultCertificate: prior.DefaultCertificate,
					LinkedActions:      linkedActionsSetToList(prior.LinkedActions),
					ForwardFor:         prior.ForwardFor,
					TimeoutClient:      prior.TimeoutClient,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func linkedActionsSetToList(value types.Set) types.List {
	switch {
	case value.IsNull():
		return types.ListNull(types.StringType)
	case value.IsUnknown():
		return types.ListUnknown(types.StringType)
	default:
		elements := make([]attr.Value, 0, len(value.Elements()))
		elements = append(elements, value.Elements()...)
		return types.ListValueMust(types.StringType, elements)
	}
}
