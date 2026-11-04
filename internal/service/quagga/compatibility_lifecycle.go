// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const bgpPendingStateField = "frr_pending_activation_v1"

type bgpPrivate interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

func pendingBGP(ctx context.Context, p bgpPrivate) (string, diag.Diagnostics) {
	if p == nil || reflect.ValueOf(p).IsNil() {
		return "", nil
	}
	raw, d := p.GetKey(ctx, bgpPendingStateField)
	if len(raw) > 0 && string(raw) != `"mutation"` && string(raw) != `"delete"` {
		d.AddError("Invalid FRR pending activation", "Unrecognized private recovery marker")
	}
	return string(raw), d
}

func markBGP(ctx context.Context, p bgpPrivate, marker string) diag.Diagnostics {
	if p == nil || reflect.ValueOf(p).IsNil() {
		return nil
	}
	if marker == "" {
		return p.SetKey(ctx, bgpPendingStateField, nil)
	}
	return p.SetKey(ctx, bgpPendingStateField, []byte(marker))
}

func strictBGPOpts(opts opnsense.ReqOpts) opnsense.ReqOpts {
	opts.StrictDeleteResponse = true
	opts.StrictMissingObjectResponse = true
	opts.StrictReconfigureResponse = true
	return opts
}

func bgpValueUnknown(value any) bool {
	if v, ok := value.(interface{ IsUnknown() bool }); ok && v.IsUnknown() {
		return true
	}
	if set, ok := value.(types.Set); ok {
		for _, element := range set.Elements() {
			if str, ok := element.(types.String); ok && !str.IsUnknown() && !str.IsNull() && strings.TrimSpace(str.ValueString()) == "" {
				return true
			}
			if element.IsNull() || bgpValueUnknown(element) {
				return true
			}
		}
	}
	return false
}

func resolveBGPModel(ctx context.Context, config tfsdk.Config, model any, prior ...any) error {
	value := reflect.ValueOf(model).Elem()
	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		name := typ.Field(i).Tag.Get("tfsdk")
		if name == "id" {
			continue
		}
		if !bgpValueUnknown(field.Interface()) {
			continue
		}
		// Only omitted Optional+Computed values may become null. A configured
		// expression is never mistaken for omission, including after state reuse.
		if config.Schema != nil {
			var configured attr.Value
			d := config.GetAttribute(ctx, path.Root(name), &configured)
			if d.HasError() {
				return fmt.Errorf("cannot resolve configured BGP %s", name)
			}
			if configured.IsNull() {
				if len(prior) != 0 {
					previous := reflect.ValueOf(prior[0]).Elem().Field(i)
					if !bgpValueUnknown(previous.Interface()) {
						field.Set(previous)
						continue
					}
				}
				field.Set(reflect.ValueOf(configured))
				continue
			}
		}
		return fmt.Errorf("BGP %s intent is unknown; resolve configured values before applying", name)
	}
	for _, name := range []string{"Capabilities", "ListenRanges"} {
		f := value.FieldByName(name)
		if !f.IsValid() {
			continue
		}
		set := f.Interface().(types.Set)
		if set.IsNull() || set.IsUnknown() {
			continue
		}
		var members []string
		if set.ElementsAs(ctx, &members, false).HasError() {
			return fmt.Errorf("BGP %s members are invalid", name)
		}
		for _, member := range members {
			if strings.Contains(member, ",") {
				return fmt.Errorf("BGP %s requires one value per set member, not CSV", name)
			}
			wire := "capabilities"
			if name == "ListenRanges" {
				wire = "listenranges"
			}
			if err := validateAdvancedWire(map[string]string{wire: member}); err != nil {
				return err
			}
		}
	}
	if m, ok := model.(*BGPNeighborResourceModel); ok {
		if m.Family.IsNull() {
			f, err := legacyFamilies(m.Address.ValueString(), m.MultiProtocol)
			if err != nil {
				return err
			}
			m.Family = f
		}
		if _, err := validateFamilies(ctx, m.Family); err != nil {
			return err
		}
		if config.Schema != nil {
			var multi types.Bool
			var family types.Set
			if config.GetAttribute(ctx, path.Root("multi_protocol"), &multi).HasError() || config.GetAttribute(ctx, path.Root("family"), &family).HasError() {
				return fmt.Errorf("cannot inspect BGP family aliases")
			}
			if multi.IsUnknown() || family.IsUnknown() {
				return fmt.Errorf("configured BGP family intent remains unknown")
			}
			if family.IsNull() {
				derived := !multi.IsNull()
				selected := multi
				if !derived && len(prior) != 0 {
					if old, ok := prior[0].(*BGPNeighborResourceModel); ok && !old.Address.Equal(m.Address) && !old.MultiProtocol.IsNull() && !old.MultiProtocol.IsUnknown() && !old.MultiProtocol.ValueBool() {
						previous, err := legacyFamilies(old.Address.ValueString(), old.MultiProtocol)
						if err == nil && previous.Equal(old.Family) {
							derived = true
							selected = old.MultiProtocol
						}
					}
				}
				if derived {
					resolved, err := legacyFamilies(m.Address.ValueString(), selected)
					if err != nil {
						return err
					}
					m.Family = resolved
					m.MultiProtocol = selected
				}
			}
			if !multi.IsNull() && !family.IsNull() {
				expected, err := legacyFamilies(m.Address.ValueString(), multi)
				if err != nil || !expected.Equal(family) {
					return fmt.Errorf("explicit family conflicts with multi_protocol")
				}
			}
		}
	}
	if m, ok := model.(*BGPPeerGroupResourceModel); ok {
		if m.Family.IsNull() {
			m.Family = sliceToSet([]string{"ipv4"})
		}
		if _, err := validateFamilies(ctx, m.Family); err != nil {
			return err
		}
	}
	return nil
}

func policyStateIDs(ctx context.Context, value any) ([]string, bool, error) {
	switch selected := value.(type) {
	case types.String:
		if selected.IsNull() || selected.IsUnknown() {
			return nil, false, nil
		}
		if selected.ValueString() == "" {
			return nil, true, nil
		}
		return []string{selected.ValueString()}, true, nil
	case types.Set:
		if selected.IsNull() || selected.IsUnknown() {
			return nil, false, nil
		}
		var ids []string
		if d := selected.ElementsAs(ctx, &ids, false); d.HasError() {
			return nil, false, fmt.Errorf("FRR policy state is malformed")
		}
		return ids, true, nil
	default:
		return nil, false, fmt.Errorf("FRR policy state has unsupported type")
	}
}

func normalizeBGPRead[M any](ctx context.Context, c *opnsense.Client, capability frrCapability, prior *M, current *M) error {
	outValue := reflect.ValueOf(current).Elem()
	if relation := outValue.FieldByName("PeerGroup"); relation.IsValid() {
		id := relation.Interface().(types.String).ValueString()
		if id != "" {
			remote, err := opnsense.Get[bgpPeerGroupAPIResponse](ctx, c, strictBGPOpts(bgpPeerGroupReqOpts), id)
			if err != nil {
				return fmt.Errorf("FRR peer-group relation is stale or unreadable: %w", err)
			}
			if err := validateBGPResponse(remote, capability); err != nil {
				return err
			}
		}
	}
	if !capability.Grouped {
		return nil
	}
	old := reflect.ValueOf(prior).Elem()
	out := reflect.ValueOf(current).Elem()
	links := map[string]string{"LinkedPrefixlistIn": "prefixlist", "LinkedPrefixlistOut": "prefixlist", "LinkedRoutemapIn": "routemap", "LinkedRoutemapOut": "routemap", "RouteMap": "routemap", "MatchPrefix": "prefixlist", "MatchASPath": "aspath", "MatchCommunity": "communitylist"}
	cache := map[string]policyMembership{}
	for name, kind := range links {
		field := out.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		got, known, err := policyStateIDs(ctx, field.Interface())
		if err != nil {
			return err
		}
		if !known {
			return fmt.Errorf("FRR policy read-back is unknown")
		}
		if len(got) == 0 {
			continue
		}
		previous := old.FieldByName(name).Interface()
		want, priorKnown, err := policyStateIDs(ctx, previous)
		if err != nil {
			return err
		}
		members, ok := cache[kind]
		if !ok {
			members, err = loadPolicyMembership(ctx, c, kind)
			if err != nil {
				return err
			}
			cache[kind] = members
		}
		groups, err := policyGroups(got, members)
		if err != nil {
			return err
		}
		if name != "MatchPrefix" && name != "MatchASPath" && name != "MatchCommunity" && len(groups) > 1 {
			return fmt.Errorf("FRR %s returned multiple logical policy groups", name)
		}
		if priorKnown && equivalentPolicies(want, got, members) {
			field.Set(reflect.ValueOf(previous))
		}
	}
	return nil
}

func bgpCreate[M, A, P any](ctx context.Context, c *opnsense.Client, opts opnsense.ReqOpts, req resource.CreateRequest, resp *resource.CreateResponse, to func(*M) *P, from func(*M, *A, string)) {
	var plan M
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := resolveBGPModel(ctx, req.Config, &plan); err != nil {
		resp.Diagnostics.AddError("Invalid BGP intent", frrDiagnostic(err))
		return
	}
	if err := bgpOwnerMu.Acquire(ctx, 1); err != nil {
		resp.Diagnostics.AddError("FRR operation canceled while waiting", frrDiagnostic(err))
		return
	}
	defer bgpOwnerMu.Release(1)
	capability, err := discoverFRR(ctx, c)
	if err != nil {
		resp.Diagnostics.AddError("FRR compatibility check failed", frrDiagnostic(err))
		return
	}
	payload, err := prepareBGPRequest(ctx, c, capability, to(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Invalid BGP policy", frrDiagnostic(err))
		return
	}
	id, err := createBGPConfirmed(ctx, c, strictBGPOpts(opts), payload)
	if id == "" {
		if err != nil {
			resp.Diagnostics.AddError("BGP create outcome unconfirmed", frrDiagnostic(err)+"; inspect appliance objects before retrying; no ambiguous identity was adopted")
		}
		return
	}
	reflect.ValueOf(&plan).Elem().FieldByName("ID").Set(reflect.ValueOf(types.StringValue(id)))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
	if err != nil {
		resp.Diagnostics.AddError("BGP saved; activation pending", frrDiagnostic(err))
		return
	}
	result, err := opnsense.Get[A](ctx, c, strictBGPOpts(opts), id)
	if err != nil {
		resp.Diagnostics.AddError("BGP saved; read-back pending", frrDiagnostic(err))
		return
	}
	if err := validateBGPResponse(result, capability); err != nil {
		resp.Diagnostics.AddError("Invalid FRR read-back", frrDiagnostic(err))
		return
	}
	current := plan
	from(&current, result, id)
	if err := normalizeBGPRead(ctx, c, capability, &plan, &current); err != nil {
		resp.Diagnostics.AddError("BGP saved; policy read-back failed", frrDiagnostic(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &current)...)
	resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
}

func bgpRead[M, A any](ctx context.Context, c *opnsense.Client, opts opnsense.ReqOpts, req resource.ReadRequest, resp *resource.ReadResponse, from func(*M, *A, string)) {
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := bgpOwnerMu.Acquire(ctx, 1); err != nil {
		resp.Diagnostics.AddError("FRR operation canceled while waiting", frrDiagnostic(err))
		return
	}
	defer bgpOwnerMu.Release(1)
	capability, err := discoverFRR(ctx, c)
	if err != nil {
		resp.Diagnostics.AddError("FRR compatibility check failed", frrDiagnostic(err))
		return
	}
	pending, d := pendingBGP(ctx, resp.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pending != "" {
		if err := opnsense.RetryReconfigure(ctx, c, strictBGPOpts(opts)); err != nil {
			resp.Diagnostics.AddError("BGP activation pending", frrDiagnostic(err))
			return
		}
		resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
	}
	if pending == `"delete"` {
		resp.State.RemoveResource(ctx)
		return
	}
	id := reflect.ValueOf(state).FieldByName("ID").Interface().(types.String).ValueString()
	result, err := opnsense.Get[A](ctx, c, strictBGPOpts(opts), id)
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading BGP object", frrDiagnostic(err))
		return
	}
	if err := validateBGPResponse(result, capability); err != nil {
		resp.Diagnostics.AddError("Invalid FRR read-back", frrDiagnostic(err))
		return
	}
	current := state
	from(&current, result, id)
	if err := normalizeBGPRead(ctx, c, capability, &state, &current); err != nil {
		resp.Diagnostics.AddError("Error resolving BGP policy identities", frrDiagnostic(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &current)...)
}

func bgpUpdate[M, A, P any](ctx context.Context, c *opnsense.Client, opts opnsense.ReqOpts, req resource.UpdateRequest, resp *resource.UpdateResponse, to func(*M) *P, from func(*M, *A, string)) {
	var plan, state M
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := resolveBGPModel(ctx, req.Config, &plan, &state); err != nil {
		resp.Diagnostics.AddError("Invalid BGP intent", frrDiagnostic(err))
		return
	}
	if err := bgpOwnerMu.Acquire(ctx, 1); err != nil {
		resp.Diagnostics.AddError("FRR operation canceled while waiting", frrDiagnostic(err))
		return
	}
	defer bgpOwnerMu.Release(1)
	capability, err := discoverFRR(ctx, c)
	if err != nil {
		resp.Diagnostics.AddError("FRR compatibility check failed", frrDiagnostic(err))
		return
	}
	pending, d := pendingBGP(ctx, resp.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pending != "" {
		if err := opnsense.RetryReconfigure(ctx, c, strictBGPOpts(opts)); err != nil {
			resp.Diagnostics.AddError("BGP activation pending", frrDiagnostic(err))
			return
		}
		resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
	}
	payload, err := prepareBGPRequest(ctx, c, capability, to(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Invalid BGP policy", frrDiagnostic(err))
		return
	}
	id := reflect.ValueOf(state).FieldByName("ID").Interface().(types.String).ValueString()
	if pending == `"mutation"` && reflect.DeepEqual(plan, state) {
		err = nil
	} else {
		err = opnsense.Update(ctx, c, strictBGPOpts(opts), payload, id)
	}
	if err != nil {
		var saved *opnsense.MutationReconfigureError
		if !errors.As(err, &saved) {
			var validation *opnsense.ValidationError
			if !errors.As(err, &validation) {
				resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
			}
			resp.Diagnostics.AddError("Error updating BGP object", frrDiagnostic(err))
			return
		}
	}
	reflect.ValueOf(&plan).Elem().FieldByName("ID").Set(reflect.ValueOf(types.StringValue(id)))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"mutation"`)...)
	if err != nil {
		resp.Diagnostics.AddError("BGP saved; activation pending", frrDiagnostic(err))
		return
	}
	result, err := opnsense.Get[A](ctx, c, strictBGPOpts(opts), id)
	if err != nil {
		resp.Diagnostics.AddError("BGP saved; read-back pending", frrDiagnostic(err))
		return
	}
	if err := validateBGPResponse(result, capability); err != nil {
		resp.Diagnostics.AddError("Invalid FRR read-back", frrDiagnostic(err))
		return
	}
	current := plan
	from(&current, result, id)
	if err := normalizeBGPRead(ctx, c, capability, &plan, &current); err != nil {
		resp.Diagnostics.AddError("BGP saved; policy read-back failed", frrDiagnostic(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &current)...)
	resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
}

func bgpDelete[M any](ctx context.Context, c *opnsense.Client, opts opnsense.ReqOpts, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := bgpOwnerMu.Acquire(ctx, 1); err != nil {
		resp.Diagnostics.AddError("FRR operation canceled while waiting", frrDiagnostic(err))
		return
	}
	defer bgpOwnerMu.Release(1)
	if _, err := discoverFRR(ctx, c); err != nil {
		resp.Diagnostics.AddError("FRR compatibility check failed", frrDiagnostic(err))
		return
	}
	pending, d := pendingBGP(ctx, resp.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pending != "" {
		if err := opnsense.RetryReconfigure(ctx, c, strictBGPOpts(opts)); err != nil {
			resp.Diagnostics.AddError("BGP activation pending", frrDiagnostic(err))
			return
		}
		resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
		if pending == `"delete"` {
			return
		}
	}
	id := reflect.ValueOf(state).FieldByName("ID").Interface().(types.String).ValueString()
	err := opnsense.Delete(ctx, c, strictBGPOpts(opts), id)
	if err != nil {
		var saved *opnsense.MutationReconfigureError
		var nf *opnsense.NotFoundError
		switch {
		case errors.As(err, &saved):
			resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"delete"`)...)
		case errors.As(err, &nf):
			return
		default:
			var validation *opnsense.ValidationError
			if !errors.As(err, &validation) {
				_, readErr := opnsense.Get[map[string]any](ctx, c, strictBGPOpts(opts), id)
				if errors.As(readErr, &nf) {
					resp.Diagnostics.Append(markBGP(ctx, resp.Private, `"delete"`)...)
					if activationErr := opnsense.RetryReconfigure(ctx, c, strictBGPOpts(opts)); activationErr != nil {
						resp.Diagnostics.AddError("BGP deletion confirmed; activation pending", frrDiagnostic(activationErr))
						return
					}
					resp.Diagnostics.Append(markBGP(ctx, resp.Private, "")...)
					return
				}
			}
		}
		resp.Diagnostics.AddError("Error deleting BGP object", frrDiagnostic(err))
	}
}
