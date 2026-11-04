// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func familyAttribute() schema.SetAttribute {
	return schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true, MarkdownDescription: "Activated address families: lowercase `ipv4`, `ipv6`, or both. Omission preserves legacy multi_protocol/transport behavior. Explicit family may differ from transport only on os-frr 2.0."}
}

func legacyFamilies(address string, multi types.Bool) (types.Set, error) {
	if multi.IsUnknown() {
		return types.SetUnknown(types.StringType), nil
	}
	if multi.ValueBool() {
		return sliceToSet([]string{"ipv4", "ipv6"}), nil
	}
	ip, err := netip.ParseAddr(address)
	if err != nil {
		return types.SetNull(types.StringType), fmt.Errorf("BGP address must be a valid IP before resolving single-stack family")
	}
	if ip.Is6() {
		return sliceToSet([]string{"ipv6"}), nil
	}
	return sliceToSet([]string{"ipv4"}), nil
}

func validateFamilies(ctx context.Context, families types.Set) ([]string, error) {
	if families.IsUnknown() || families.IsNull() {
		return nil, fmt.Errorf("BGP family intent is unresolved; no mutation was sent")
	}
	var values []string
	if d := families.ElementsAs(ctx, &values, false); d.HasError() {
		return nil, fmt.Errorf("BGP family must be a typed set of strings")
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("BGP family must contain ipv4, ipv6, or both")
	}
	for _, v := range values {
		if v != "ipv4" && v != "ipv6" {
			return nil, fmt.Errorf("BGP family must contain only lowercase ipv4 and ipv6")
		}
	}
	return values, nil
}

func (r *bgpNeighborResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan BGPNeighborResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var family types.Set
	var multi types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("family"), &family)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("multi_protocol"), &multi)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if family.IsUnknown() || multi.IsUnknown() {
		return
	} // Unknown configured intent remains unknown until apply.
	switch {
	case !family.IsNull():
		values, err := validateFamilies(ctx, family)
		if err != nil {
			resp.Diagnostics.AddError("Invalid BGP family", err.Error())
			return
		}
		if !multi.IsNull() {
			if !multi.ValueBool() && len(values) == 2 {
				resp.Diagnostics.AddError("Conflicting BGP family aliases", "multi_protocol=false cannot select both families")
				return
			}
			expected, err := legacyFamilies(plan.Address.ValueString(), multi)
			if (err != nil && !plan.Address.IsUnknown()) || (err == nil && !expected.Equal(family)) {
				resp.Diagnostics.AddError("Conflicting BGP family aliases", "family and explicitly configured multi_protocol describe different address families")
				return
			}
		}
		plan.Family = family
		plan.MultiProtocol = types.BoolValue(len(values) == 2)
	case !multi.IsNull() || req.State.Raw.IsNull():
		resolved, err := legacyFamilies(plan.Address.ValueString(), multi)
		if err != nil && !plan.Address.IsUnknown() {
			resp.Diagnostics.AddError("Invalid BGP transport", err.Error())
			return
		}
		if err == nil {
			plan.Family = resolved
		}
	default:
		var state BGPNeighborResourceModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		plan.Family = state.Family
		plan.MultiProtocol = state.MultiProtocol
		if !plan.Address.Equal(state.Address) && !state.MultiProtocol.IsNull() && !state.MultiProtocol.IsUnknown() && !state.MultiProtocol.ValueBool() {
			previous, err := legacyFamilies(state.Address.ValueString(), state.MultiProtocol)
			if err == nil && previous.Equal(state.Family) {
				if plan.Address.IsUnknown() {
					plan.Family = types.SetUnknown(types.StringType)
				} else {
					resolved, err := legacyFamilies(plan.Address.ValueString(), state.MultiProtocol)
					if err != nil {
						resp.Diagnostics.AddError("Invalid BGP transport", err.Error())
						return
					}
					plan.Family = resolved
				}
			}
		}
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() || req.State.Raw.IsNull() {
		return
	}
	// The framework may have marked computed fields unknown when applying
	// the legacy alias's false default, before explicit family reconciles it.
	// Reuse only genuinely omitted state; configured unknown expressions are
	// never defaults/clears and must remain unknown. Known-null timers also
	// matter: UseStateForUnknown alone intentionally ignores null state.
	for _, name := range []string{"remote_as", "remote_as_mode", "local_as", "weight", "local_ip", "link_local_interface", "next_hop_self_all", "multihop", "route_reflector_client", "soft_reconfiguration_inbound", "bfd", "bfd_strict", "connect_timer", "default_originate", "as_override", "remove_private_as", "allow_as_in", "disable_connected_check", "attribute_unchanged", "capabilities", "keepalive", "holddown", "peer_group", "linked_prefixlist_in", "linked_prefixlist_out", "linked_routemap_in", "linked_routemap_out"} {
		var configured, planned, prior attr.Value
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(name), &configured)...)
		resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root(name), &planned)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !configured.IsNull() || !planned.IsUnknown() {
			continue
		}
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(name), &prior)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(name), prior)...)
	}
}

func (r *bgpNeighborResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{0: {StateUpgrader: upgradeNeighborV0}}
}

func upgradeNeighborV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil {
		resp.Diagnostics.AddError("Missing legacy BGP state", "Offline upgrade requires persisted version-0 state")
		return
	}
	var state map[string]json.RawMessage
	if json.Unmarshal(req.RawState.JSON, &state) != nil {
		resp.Diagnostics.AddError("Malformed legacy BGP state", "Expected JSON object")
		return
	}
	var address string
	var multi *bool
	if json.Unmarshal(state["address"], &address) != nil || json.Unmarshal(state["multi_protocol"], &multi) != nil || multi == nil {
		resp.Diagnostics.AddError("Malformed legacy BGP state", "Expected known address and multi_protocol boolean; no default was substituted")
		return
	}
	family, err := legacyFamilies(address, types.BoolValue(*multi))
	if err != nil {
		resp.Diagnostics.AddError("Malformed legacy BGP state", err.Error())
		return
	}
	var values []string
	resp.Diagnostics.Append(family.ElementsAs(ctx, &values, false)...)
	state["family"], _ = json.Marshal(values)
	state["peer_group"] = json.RawMessage(`null`)
	for _, name := range []string{"remote_as_mode", "local_as", "weight", "local_ip", "link_local_interface", "next_hop_self_all", "multihop", "route_reflector_client", "soft_reconfiguration_inbound", "bfd", "bfd_strict", "connect_timer", "default_originate", "as_override", "remove_private_as", "allow_as_in", "disable_connected_check", "attribute_unchanged", "capabilities", "password"} {
		if _, exists := state[name]; !exists {
			state[name] = json.RawMessage(`null`)
		}
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		resp.Diagnostics.AddError("Error upgrading BGP state", err.Error())
		return
	}
	dynamic := tfprotov6.DynamicValue{JSON: encoded}
	value, err := dynamic.Unmarshal(resp.State.Schema.Type().TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error upgrading BGP state", err.Error())
		return
	}
	resp.State.Raw = value
}

func boolFromWire(value string) types.Bool { return types.BoolValue(value == "1") }
