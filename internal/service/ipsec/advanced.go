// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Names and limits are pinned to core 26.7.8 Swanctl.xml, not daemon defaults.
type ipsecAdvancedField struct {
	name, wire, description string
	boolean, selection      bool
	minimum, maximum        int
	blank                   bool
	options                 []string
}

var connectionAdvancedFields = []ipsecAdvancedField{
	{name: "aggressive", wire: "aggressive", description: "Enable aggressive IKEv1 mode.", boolean: true},
	{name: "dpd_delay", wire: "dpd_delay", description: "DPD delay in seconds; blank or 0..500000.", maximum: 500000, blank: true},
	{name: "dpd_timeout", wire: "dpd_timeout", description: "IKEv1 DPD timeout in seconds; blank or 0..500000.", maximum: 500000, blank: true},
	{name: "ike_lifetime", wire: "over_time", description: "Overtime grace in seconds, NOT total IKE lifetime; blank or 0..500000.", maximum: 500000, blank: true},
	{name: "ip_pools", wire: "pools", description: "Comma-separated pool UUIDs (or the built-in radius pool).", selection: true},
	{name: "keying_tries", wire: "keyingtries", description: "Keying attempts; blank or 0..1000 (zero retries indefinitely).", maximum: 1000, blank: true},
	{name: "local_addresses", wire: "local_addrs", description: "Comma-separated local IKE addresses.", selection: true},
	{name: "local_port", wire: "local_port", description: "Blank uses daemon port 500; otherwise 4500.", options: []string{"", "4500"}},
	{name: "mobike", wire: "mobike", description: "Enable MOBIKE.", boolean: true},
	{name: "reauthentication_time", wire: "reauth_time", description: "Reauthentication interval in seconds; blank or 0..500000.", maximum: 500000, blank: true},
	{name: "rekey_time", wire: "rekey_time", description: "IKE rekey interval in seconds; blank or 0..500000.", maximum: 500000, blank: true},
	{name: "remote_port", wire: "remote_port", description: "Blank uses daemon port 500; otherwise 4500.", options: []string{"", "4500"}},
	{name: "send_certificate", wire: "send_cert", description: "Certificate sending policy: blank, ifasked, never, always.", options: []string{"", "ifasked", "never", "always"}},
	{name: "send_certificate_request", wire: "send_certreq", description: "Send certificate requests.", boolean: true},
	{name: "udp_encapsulation", wire: "encap", description: "Force UDP encapsulation.", boolean: true},
}

var childAdvancedFields = []ipsecAdvancedField{
	{name: "close_action", wire: "close_action", description: "Action on close: none, trap, start.", options: []string{"none", "trap", "start"}},
	{name: "dpd_action", wire: "dpd_action", description: "Action on DPD timeout: clear, trap, start.", options: []string{"clear", "trap", "start"}},
	{name: "install_policies", wire: "policies", description: "Install IPsec policies.", boolean: true},
	{name: "request_id", wire: "reqid", description: "Request ID: blank or 1..65535.", minimum: 1, maximum: 65535, blank: true},
	{name: "sha256_96", wire: "sha256_96", description: "Use SHA256 96-bit truncation.", boolean: true},
	{name: "rekey_time", wire: "rekey_time", description: "Child rekey interval in seconds: 0..500000; cannot be blank. Omission uses server default 3600 on creation.", maximum: 500000},
}

func (f ipsecAdvancedField) validate(value string) error {
	if f.boolean {
		if value != "0" && value != "1" {
			return fmt.Errorf("expected boolean flag 0 or 1")
		}
		return nil
	}
	if f.selection {
		if value == "" {
			return nil
		}
		seen := map[string]bool{}
		for _, member := range strings.Split(value, ",") {
			member = strings.TrimSpace(member)
			if member == "" || seen[member] {
				return fmt.Errorf("selection must contain nonempty, distinct normalized members")
			}
			seen[member] = true
			if f.name == "ip_pools" && member != "radius" {
				parsed, err := uuid.Parse(member)
				if err != nil || parsed.String() != member {
					return fmt.Errorf("pool selections must be canonical UUIDs or radius")
				}
			}
		}
		return nil
	}
	if f.options != nil {
		for _, option := range f.options {
			if value == option {
				return nil
			}
		}
		return fmt.Errorf("expected one of %q", f.options)
	}
	if value == "" && f.blank {
		return nil
	}
	if value == "" || strings.Trim(value, "0123456789") != "" {
		return fmt.Errorf("expected decimal digits in range %d..%d", f.minimum, f.maximum)
	}
	number, err := strconv.Atoi(value)
	if err != nil || number < f.minimum || number > f.maximum {
		return fmt.Errorf("expected decimal value in range %d..%d", f.minimum, f.maximum)
	}
	return nil
}

type ipsecStringValidator struct{ field ipsecAdvancedField }

func (v ipsecStringValidator) Description(context.Context) string { return v.field.description }
func (v ipsecStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ipsecStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if err := v.field.validate(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid IPsec value", err.Error())
	}
}

func addIPsecResourceAttributes(attributes map[string]schema.Attribute, fields []ipsecAdvancedField) {
	for _, field := range fields {
		description := field.description + " Omission preserves imported values; explicit false, zero and permitted empty strings are written."
		if field.boolean {
			attributes[field.name] = schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: description}
		} else {
			attributes[field.name] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: description, Validators: []validator.String{ipsecStringValidator{field}}}
		}
	}
}

func addIPsecLookupAttributes(attributes map[string]dsschema.Attribute, fields []ipsecAdvancedField) {
	for _, field := range fields {
		if field.boolean {
			attributes[field.name] = dsschema.BoolAttribute{Computed: true, MarkdownDescription: field.description}
		} else {
			attributes[field.name] = dsschema.StringAttribute{Computed: true, MarkdownDescription: field.description}
		}
	}
}

// Configuration, not the plan/state, establishes ownership of additive fields.
func configuredIPsecRequest(ctx context.Context, config tfsdk.Config, legacy any, fields []ipsecAdvancedField, diagnostics *diag.Diagnostics) *map[string]string {
	if !config.Raw.IsFullyKnown() {
		diagnostics.AddError("Unresolved IPsec configuration", "All explicitly configured values must be known before mutation.")
		return nil
	}
	body, err := json.Marshal(legacy)
	if err != nil {
		diagnostics.AddError("Invalid IPsec request", "Could not serialize request.")
		return nil
	}
	request := map[string]string{}
	if err = json.Unmarshal(body, &request); err != nil {
		diagnostics.AddError("Invalid IPsec request", "Could not prepare request.")
		return nil
	}
	for _, field := range fields {
		var value string
		if field.boolean {
			var configured types.Bool
			diagnostics.Append(config.GetAttribute(ctx, path.Root(field.name), &configured)...)
			if configured.IsNull() {
				continue
			}
			if configured.IsUnknown() {
				diagnostics.AddAttributeError(path.Root(field.name), "Unresolved IPsec configuration", "An explicitly configured value must be known before mutation.")
				continue
			}
			value = opnsense.BoolToString(configured.ValueBool())
		} else {
			var configured types.String
			diagnostics.Append(config.GetAttribute(ctx, path.Root(field.name), &configured)...)
			if configured.IsNull() {
				continue
			}
			if configured.IsUnknown() {
				diagnostics.AddAttributeError(path.Root(field.name), "Unresolved IPsec configuration", "An explicitly configured value must be known before mutation.")
				continue
			}
			value = configured.ValueString()
		}
		if err = field.validate(value); err != nil {
			diagnostics.AddAttributeError(path.Root(field.name), "Invalid IPsec configuration", err.Error())
			continue
		}
		if field.selection {
			value = strings.Join(opnsense.CSVToSlice(value), ",")
		}
		request[field.wire] = value
	}
	return &request
}

// Strict local selectors avoid weakening/changing shared SelectedMap behavior.
type ipsecSelection []string

func (s *ipsecSelection) UnmarshalJSON(data []byte) error {
	data = []byte(strings.TrimSpace(string(data)))
	if string(data) == "null" {
		return fmt.Errorf("null IPsec selection")
	}
	if len(data) > 0 && data[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(data, &items); err != nil || len(items) != 0 {
			return fmt.Errorf("expected an empty IPsec selection array")
		}
		*s = []string{}
		return nil
	}
	var scalar string
	if json.Unmarshal(data, &scalar) == nil {
		*s = opnsense.CSVToSlice(scalar)
		sort.Strings(*s)
		return nil
	}
	var options map[string]struct {
		Selected json.Number `json:"selected"`
	}
	if err := json.Unmarshal(data, &options); err != nil {
		return fmt.Errorf("invalid IPsec selection shape")
	}
	selected := []string{}
	for key, option := range options {
		switch option.Selected.String() {
		case "1":
			if key != "" {
				selected = append(selected, key)
			}
		case "0":
		default:
			return fmt.Errorf("invalid IPsec selection flag")
		}
	}
	sort.Strings(selected)
	*s = selected
	return nil
}

type ipsecEnum string

func (s *ipsecEnum) UnmarshalJSON(data []byte) error {
	var selection ipsecSelection
	if err := json.Unmarshal(data, &selection); err != nil {
		return err
	}
	if len(selection) > 1 {
		return fmt.Errorf("multiple values for single IPsec selection")
	}
	*s = ""
	if len(selection) == 1 {
		*s = ipsecEnum(selection[0])
	}
	return nil
}

func stableIPsecCSV(prior types.String, values []string) types.String {
	if !prior.IsNull() && !prior.IsUnknown() {
		previous := opnsense.CSVToSlice(prior.ValueString())
		current := append([]string{}, values...)
		sort.Strings(previous)
		sort.Strings(current)
		if reflect.DeepEqual(previous, current) {
			return prior
		}
	}
	return types.StringValue(strings.Join(values, ","))
}

func decodeIPsecResponse(data []byte, target any, fields []ipsecAdvancedField) error {
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, field := range fields {
		value, present := raw[field.wire]
		if !present {
			return fmt.Errorf("missing advanced IPsec field %s", field.name)
		}
		if field.selection {
			continue
		}
		var text string
		if field.options != nil {
			var option ipsecEnum
			if err := json.Unmarshal(value, &option); err != nil {
				return fmt.Errorf("invalid %s", field.name)
			}
			text = string(option)
		} else if err := json.Unmarshal(value, &text); err != nil || string(value) == "null" {
			return fmt.Errorf("invalid %s scalar", field.name)
		}
		if err := field.validate(text); err != nil {
			return fmt.Errorf("invalid %s: %w", field.name, err)
		}
	}
	return nil
}

// Only a UUID-bound failed activation is replayed; saved object writes are never
// retried here. Read-only lookups deliberately do not consume recovery state.
type ipsecPrivate interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

const ipsecActivationKey = "ipsec_pending_activation_v1"

func markIPsecActivation(ctx context.Context, private ipsecPrivate, id string, mutationErr error, diagnostics *diag.Diagnostics) {
	var pending *opnsense.MutationReconfigureError
	if !errors.As(mutationErr, &pending) {
		return
	}
	if private == nil || reflect.ValueOf(private).IsNil() {
		diagnostics.AddError("Missing IPsec activation recovery state", "The saved object remains identified, but activation must be reconciled manually.")
		return
	}
	raw, _ := json.Marshal(id)
	diagnostics.Append(private.SetKey(context.WithoutCancel(ctx), ipsecActivationKey, raw)...)
}

func recoverIPsecActivation(ctx context.Context, private ipsecPrivate, id string, client *opnsense.Client, opts opnsense.ReqOpts, diagnostics *diag.Diagnostics) {
	if private == nil || reflect.ValueOf(private).IsNil() {
		return
	}
	raw, d := private.GetKey(ctx, ipsecActivationKey)
	diagnostics.Append(d...)
	if diagnostics.HasError() || len(raw) == 0 {
		return
	}
	var saved string
	if json.Unmarshal(raw, &saved) != nil || saved != id || id == "" {
		diagnostics.AddError("Invalid IPsec activation recovery state", "The pending activation is not bound to this object UUID; reconcile state before retrying.")
		return
	}
	if err := opnsense.RetryReconfigure(ctx, client, opts); err != nil {
		diagnostics.AddError("Saved IPsec mutation still requires activation", ipsecErrorMessage(err))
		return
	}
	diagnostics.Append(private.SetKey(context.WithoutCancel(ctx), ipsecActivationKey, nil)...)
}

// API validation/result strings are untrusted and may echo unrelated sensitive
// fields. Diagnose confirmed effects/categories without rendering payload text.
func ipsecErrorMessage(err error) string {
	var saved *opnsense.MutationReconfigureError
	var rejected *opnsense.ValidationError
	var auth *opnsense.AuthError
	switch {
	case errors.As(err, &saved):
		return "The mutation was saved, but service reconfigure failed. Known identity/state are retained; refresh retries activation without repeating the save."
	case errors.As(err, &rejected):
		return "The appliance rejected field validation; no activation was attempted. Check configured fields against the pinned model limits. Server validation payloads are not rendered."
	case errors.As(err, &auth):
		return auth.Error()
	case errors.Is(err, context.Canceled):
		return "The IPsec operation was canceled; confirmed identity/state are retained."
	case errors.Is(err, context.DeadlineExceeded):
		return "The IPsec operation deadline expired; reconcile known identity/state before retrying."
	default:
		return "The IPsec API request failed or its response was malformed or did not affirm success. Check connectivity, TLS, permissions and model compatibility; reconcile known identity/state before retrying."
	}
}
