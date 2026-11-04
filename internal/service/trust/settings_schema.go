// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Import-first Trust settings singleton. Owns six audited fields only. Delete is state-only; planning refresh may recover pending activation.", Attributes: map[string]schema.Attribute{
		"id":                        schema.StringAttribute{Computed: true, MarkdownDescription: "Fixed singleton identity trust_settings.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"store_intermediate_certs":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Allow locally defined intermediate certificate authorities in the local trust store. Omission preserves existing policy; explicit false clears.", PlanModifiers: []planmodifier.Bool{settingsOmitted{}}},
		"install_crls":              schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Store all configured certificate revocation lists in the default trust store. Omission preserves existing policy; explicit false clears.", PlanModifiers: []planmodifier.Bool{settingsOmitted{}}},
		"fetch_crls":                schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Schedule hourly CRL downloads using distribution points from CAs in the trust store. Omission preserves existing policy; explicit false clears.", PlanModifiers: []planmodifier.Bool{settingsOmitted{}}},
		"enable_legacy_sect":        schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Enable OpenSSL legacy providers. Omission preserves existing policy; explicit false clears.", PlanModifiers: []planmodifier.Bool{settingsOmitted{}}},
		"enable_config_constraints": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Enable custom OpenSSL configuration constraints. Omission preserves existing policy; explicit false clears.", PlanModifiers: []planmodifier.Bool{settingsOmitted{}}},
		"cipher_string":             schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, MarkdownDescription: "Selected OpenSSL CipherString choice keys. Omission preserves policy; an empty set clears.", PlanModifiers: []planmodifier.Set{settingsOmitted{}}},
	}}
}

type settingsOmitted struct{}

func (settingsOmitted) Description(context.Context) string {
	return "Preserve prior values for omitted configuration only."
}

func (m settingsOmitted) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }

func (settingsOmitted) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.ConfigValue.IsNull() && !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}

func (settingsOmitted) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.ConfigValue.IsNull() && !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}
