// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Wire options and relink persistence follow OPNsense 26.7.8
// Interfaces/NetworkInterface.xml, NetworkInterface.php and
// FieldTypes/NetworkInterfaceField.php at core commit
// 96978001a8343669154e4492172ce0a07dac7e4b. Inactive prefix requests
// are also guarded against the newer serialization behavior in core commit
// 855dbea44aef4b88712b6e4008998a8fa6d0d697.
func assignmentControlAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"device":                           schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Existing device to bind to this assignment; must be advertised by the API before rebinding. Omit to retain the saved value.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"vlan_hardware_filtering":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "VLAN hardware filtering: enable, disable, or default. Applies when per-interface hardware settings are overridden. Omit to retain the saved value.", Validators: []validator.String{stringvalidator.OneOf("enable", "disable", "default")}},
		"dhcp6_prefix_length":              schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Requested DHCPv6 prefix length: decimal 48 through 64, or none. Exposes displayed prefix length, not the API encoding. Omit to retain the saved value.", Validators: []validator.String{stringvalidator.OneOf("none", "48", "49", "50", "51", "52", "53", "54", "55", "56", "57", "58", "59", "60", "61", "62", "63", "64")}},
		"dhcp6_prefix_only":                schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Request only an IPv6 prefix through DHCPv6. Omit to retain the saved value."},
		"dhcp6_send_prefix_hint":           schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Send the requested prefix length as a DHCPv6 prefix hint. Omit to retain the saved value."},
		"dhcp6_prefix_id":                  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "DHCPv6 hexadecimal prefix ID (0x prefix), up to 0x7fffffffffffffff because the appliance serializes through a signed 64-bit integer. Empty clears the override. Omit to retain the saved value.", Validators: []validator.String{assignmentHexValidator{}}},
		"dhcp6_interface_id":               schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "DHCPv6 hexadecimal interface ID (0x prefix), up to 0x7fffffffffffffff because the appliance serializes through a signed 64-bit integer. Empty clears the override. Omit to retain the saved value.", Validators: []validator.String{assignmentHexValidator{}}},
		"dhcp6_identity_association_id":    schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "DHCPv6 identity association ID as a canonical nonnegative signed-64-bit decimal string; empty selects the default. Omit to retain the saved value.", Validators: []validator.String{assignmentAssociationIDValidator{}}},
		"description":                      schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Interface description (letters, digits and underscores). Omit to retain the saved value.", Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_]{0,255}$`), "must contain at most 255 letters, digits or underscores")}},
		"mtu":                              schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "MTU in bytes (576-65535); zero clears the override and uses the system default. Omit to retain the saved value.", Validators: []validator.Int64{int64validator.Any(int64validator.OneOf(0), int64validator.Between(576, 65535))}},
		"mss":                              schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "MSS in bytes (576-65535); zero clears the override and uses the system default. Omit to retain the saved value.", Validators: []validator.Int64{int64validator.Any(int64validator.OneOf(0), int64validator.Between(576, 65535))}},
		"block_private":                    schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Block private networks on this interface. Omit to retain the saved value."},
		"block_bogons":                     schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Block bogon networks on this interface. Omit to retain the saved value."},
		"override_hardware_settings":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Override global hardware offload settings for this interface. Omit to retain the saved value."},
		"disable_checksum_offloading":      schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Disable hardware checksum offloading when hardware settings are overridden. Omit to retain the saved value."},
		"disable_segmentation_offloading":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Disable TCP segmentation offloading when hardware settings are overridden. Omit to retain the saved value."},
		"disable_large_receive_offloading": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Disable large receive offloading when hardware settings are overridden. Omit to retain the saved value."},
	}
}

func (m *AssignmentResourceModel) readControls(a *assignmentAPIResponse) error {
	var err error
	m.Device, err = assignmentControlSelection(a, "if")
	if err != nil {
		return err
	}
	m.VLANHardwareFiltering, err = assignmentControlSelection(a, "disablevlanhwfilter")
	if err != nil {
		return err
	}
	m.DHCP6PrefixLength, err = assignmentControlSelection(a, "dhcp6-ia-pd-len")
	if err != nil {
		return err
	}
	m.DHCP6PrefixOnly, err = assignmentControlBool(a, "dhcp6prefixonly")
	if err != nil {
		return err
	}
	m.DHCP6SendPrefixHint, err = assignmentControlBool(a, "dhcp6-ia-pd-send-hint")
	if err != nil {
		return err
	}
	m.DHCP6PrefixID, err = assignmentControlHexString(a, "dhcp6-prefix-id", m.DHCP6PrefixID)
	if err != nil {
		return err
	}
	m.DHCP6InterfaceID, err = assignmentControlHexString(a, "dhcp6_ifid", m.DHCP6InterfaceID)
	if err != nil {
		return err
	}
	m.DHCP6IdentityAssociationID, err = assignmentControlString(a, "dhcp6_assoc_pd")
	if err != nil {
		return err
	}
	if err := m.decodeControlOptions(); err != nil {
		return err
	}
	m.Description, err = assignmentControlString(a, "descr")
	if err != nil {
		return err
	}
	m.MTU, err = assignmentControlInt64(a, "mtu")
	if err != nil {
		return err
	}
	m.MSS, err = assignmentControlInt64(a, "mss")
	if err != nil {
		return err
	}
	m.BlockPrivate, err = assignmentControlBool(a, "blockpriv")
	if err != nil {
		return err
	}
	m.BlockBogons, err = assignmentControlBool(a, "blockbogons")
	if err != nil {
		return err
	}
	m.OverrideHardwareSettings, err = assignmentControlBool(a, "hw_settings_overwrite")
	if err != nil {
		return err
	}
	m.DisableChecksumOffloading, err = assignmentControlBool(a, "disablechecksumoffloading")
	if err != nil {
		return err
	}
	m.DisableSegmentationOffloading, err = assignmentControlBool(a, "disablesegmentationoffloading")
	if err != nil {
		return err
	}
	m.DisableLargeReceiveOffloading, err = assignmentControlBool(a, "disablelargereceiveoffloading")
	if err != nil {
		return err
	}
	return nil
}

func (m *AssignmentResourceModel) applyControls(current *assignmentAPIResponse, request assignmentAPIRequest) error {
	if !m.DHCP6PrefixLength.IsNull() && !m.DHCP6PrefixLength.IsUnknown() {
		wanted := m.DHCP6PrefixLength.ValueString()
		length, err := strconv.Atoi(wanted)
		if wanted != "none" && (err != nil || length < 48 || length > 64) {
			return fmt.Errorf("invalid requested DHCPv6 prefix length")
		}
		mode := current.stringValue("type6")
		if !m.IPv6Mode.IsNull() && !m.IPv6Mode.IsUnknown() {
			mode = m.IPv6Mode.ValueString()
		}
		if mode != "dhcp6" && wanted != "64" {
			return fmt.Errorf("configured non-default DHCPv6 prefix length requires dhcp6 mode, even when unchanged; newer appliance versions discard inactive prefix-length requests on any update")
		}
	}
	if !m.VLANHardwareFiltering.IsNull() && !m.VLANHardwareFiltering.IsUnknown() {
		switch m.VLANHardwareFiltering.ValueString() {
		case "enable", "disable", "default":
		default:
			return fmt.Errorf("invalid VLAN hardware filtering option")
		}
	}

	for _, key := range []string{"if", "disablevlanhwfilter", "dhcp6-ia-pd-len", "dhcp6prefixonly", "dhcp6-ia-pd-send-hint", "dhcp6-prefix-id", "dhcp6_ifid", "dhcp6_assoc_pd", "descr", "mtu", "mss", "blockpriv", "blockbogons", "hw_settings_overwrite", "disablechecksumoffloading", "disablesegmentationoffloading", "disablelargereceiveoffloading"} {
		if !assignmentControlPresent(current, key) {
			delete(request, key)
		}
	}
	if !m.Device.IsNull() && !m.Device.IsUnknown() {
		if !assignmentControlPresent(current, "if") {
			return fmt.Errorf("assignment API does not expose configured field if")
		}
		if err := validateAssignmentDevice(current, m.Device.ValueString()); err != nil {
			return err
		}
		request["if"] = m.Device.ValueString()
	}
	if !m.VLANHardwareFiltering.IsNull() && !m.VLANHardwareFiltering.IsUnknown() {
		if !assignmentControlPresent(current, "disablevlanhwfilter") {
			return fmt.Errorf("assignment API does not expose configured field disablevlanhwfilter")
		}
		request["disablevlanhwfilter"] = map[string]string{"enable": "0", "disable": "1", "default": "2"}[m.VLANHardwareFiltering.ValueString()]
	}
	if !m.DHCP6PrefixLength.IsNull() && !m.DHCP6PrefixLength.IsUnknown() {
		if !assignmentControlPresent(current, "dhcp6-ia-pd-len") {
			return fmt.Errorf("assignment API does not expose configured field dhcp6-ia-pd-len")
		}
		request["dhcp6-ia-pd-len"] = assignmentPrefixLengthWrite(m.DHCP6PrefixLength.ValueString())
	}
	if !m.DHCP6PrefixOnly.IsNull() && !m.DHCP6PrefixOnly.IsUnknown() {
		if !assignmentControlPresent(current, "dhcp6prefixonly") {
			return fmt.Errorf("assignment API does not expose configured field dhcp6prefixonly")
		}
		request["dhcp6prefixonly"] = opnsense.BoolToString(m.DHCP6PrefixOnly.ValueBool())
	}
	if !m.DHCP6SendPrefixHint.IsNull() && !m.DHCP6SendPrefixHint.IsUnknown() {
		if !assignmentControlPresent(current, "dhcp6-ia-pd-send-hint") {
			return fmt.Errorf("assignment API does not expose configured field dhcp6-ia-pd-send-hint")
		}
		request["dhcp6-ia-pd-send-hint"] = opnsense.BoolToString(m.DHCP6SendPrefixHint.ValueBool())
	}
	if !m.DHCP6PrefixID.IsNull() && !m.DHCP6PrefixID.IsUnknown() {
		if !assignmentControlPresent(current, "dhcp6-prefix-id") {
			return fmt.Errorf("assignment API does not expose configured field dhcp6-prefix-id")
		}
		request["dhcp6-prefix-id"] = m.DHCP6PrefixID.ValueString()
	}
	if !m.DHCP6InterfaceID.IsNull() && !m.DHCP6InterfaceID.IsUnknown() {
		if !assignmentControlPresent(current, "dhcp6_ifid") {
			return fmt.Errorf("assignment API does not expose configured field dhcp6_ifid")
		}
		request["dhcp6_ifid"] = m.DHCP6InterfaceID.ValueString()
	}
	if !m.DHCP6IdentityAssociationID.IsNull() && !m.DHCP6IdentityAssociationID.IsUnknown() {
		if !assignmentControlPresent(current, "dhcp6_assoc_pd") {
			return fmt.Errorf("assignment API does not expose configured field dhcp6_assoc_pd")
		}
		request["dhcp6_assoc_pd"] = m.DHCP6IdentityAssociationID.ValueString()
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		if !assignmentControlPresent(current, "descr") {
			return fmt.Errorf("assignment API does not expose configured field descr")
		}
		request["descr"] = m.Description.ValueString()
	}
	if !m.MTU.IsNull() && !m.MTU.IsUnknown() {
		if !assignmentControlPresent(current, "mtu") {
			return fmt.Errorf("assignment API does not expose configured field mtu")
		}
		request["mtu"] = assignmentControlIntegerWrite(m.MTU.ValueInt64())
	}
	if !m.MSS.IsNull() && !m.MSS.IsUnknown() {
		if !assignmentControlPresent(current, "mss") {
			return fmt.Errorf("assignment API does not expose configured field mss")
		}
		request["mss"] = assignmentControlIntegerWrite(m.MSS.ValueInt64())
	}
	if !m.BlockPrivate.IsNull() && !m.BlockPrivate.IsUnknown() {
		if !assignmentControlPresent(current, "blockpriv") {
			return fmt.Errorf("assignment API does not expose configured field blockpriv")
		}
		request["blockpriv"] = opnsense.BoolToString(m.BlockPrivate.ValueBool())
	}
	if !m.BlockBogons.IsNull() && !m.BlockBogons.IsUnknown() {
		if !assignmentControlPresent(current, "blockbogons") {
			return fmt.Errorf("assignment API does not expose configured field blockbogons")
		}
		request["blockbogons"] = opnsense.BoolToString(m.BlockBogons.ValueBool())
	}
	if !m.OverrideHardwareSettings.IsNull() && !m.OverrideHardwareSettings.IsUnknown() {
		if !assignmentControlPresent(current, "hw_settings_overwrite") {
			return fmt.Errorf("assignment API does not expose configured field hw_settings_overwrite")
		}
		request["hw_settings_overwrite"] = opnsense.BoolToString(m.OverrideHardwareSettings.ValueBool())
	}
	if !m.DisableChecksumOffloading.IsNull() && !m.DisableChecksumOffloading.IsUnknown() {
		if !assignmentControlPresent(current, "disablechecksumoffloading") {
			return fmt.Errorf("assignment API does not expose configured field disablechecksumoffloading")
		}
		request["disablechecksumoffloading"] = opnsense.BoolToString(m.DisableChecksumOffloading.ValueBool())
	}
	if !m.DisableSegmentationOffloading.IsNull() && !m.DisableSegmentationOffloading.IsUnknown() {
		if !assignmentControlPresent(current, "disablesegmentationoffloading") {
			return fmt.Errorf("assignment API does not expose configured field disablesegmentationoffloading")
		}
		request["disablesegmentationoffloading"] = opnsense.BoolToString(m.DisableSegmentationOffloading.ValueBool())
	}
	if !m.DisableLargeReceiveOffloading.IsNull() && !m.DisableLargeReceiveOffloading.IsUnknown() {
		if !assignmentControlPresent(current, "disablelargereceiveoffloading") {
			return fmt.Errorf("assignment API does not expose configured field disablelargereceiveoffloading")
		}
		request["disablelargereceiveoffloading"] = opnsense.BoolToString(m.DisableLargeReceiveOffloading.ValueBool())
	}
	return nil
}

func assignmentControlString(a *assignmentAPIResponse, key string) (types.String, error) {
	raw, present := a.Raw[key]
	if !present || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return types.StringNull(), nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return types.StringNull(), fmt.Errorf("invalid assignment %s: %w", key, err)
	}
	return types.StringValue(value), nil
}

func assignmentControlBool(a *assignmentAPIResponse, key string) (types.Bool, error) {
	value, err := assignmentControlString(a, key)
	if err != nil || value.IsNull() {
		return types.BoolNull(), err
	}
	switch value.ValueString() {
	case "", "0":
		return types.BoolValue(false), nil
	case "1":
		return types.BoolValue(true), nil
	default:
		return types.BoolNull(), fmt.Errorf("invalid assignment boolean %s", key)
	}
}

func assignmentControlInt64(a *assignmentAPIResponse, key string) (types.Int64, error) {
	value, err := assignmentControlString(a, key)
	if err != nil || value.IsNull() {
		return types.Int64Null(), err
	}
	if value.ValueString() == "" {
		return types.Int64Value(0), nil
	}
	parsed, err := strconv.ParseInt(value.ValueString(), 10, 64)
	if err != nil || parsed < 576 || parsed > 65535 {
		return types.Int64Null(), fmt.Errorf("invalid assignment %s: expected empty or 576-65535", key)
	}
	return types.Int64Value(parsed), nil
}

func assignmentControlIntegerWrite(value int64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}

func assignmentControlPresent(api *assignmentAPIResponse, key string) bool {
	raw, present := api.Raw[key]
	return present && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func assignmentControlSelection(a *assignmentAPIResponse, key string) (types.String, error) {
	if !assignmentControlPresent(a, key) {
		return types.StringNull(), nil
	}
	raw := bytes.TrimSpace(a.Raw[key])
	if len(raw) > 0 && raw[0] == '{' {
		values, ok := selectedValuesInOrder(raw)
		if !ok || len(values) != 1 {
			return types.StringNull(), fmt.Errorf("invalid assignment selection %s: expected one selected option", key)
		}
	}
	if len(raw) > 0 && raw[0] == '[' {
		values, ok := selectedIndexesInList(raw)
		if !ok || len(values) != 1 {
			return types.StringNull(), fmt.Errorf("invalid assignment selection %s: expected one selected option", key)
		}
	}
	value, err := assignmentRequestValue(raw)
	text, ok := value.(string)
	if err != nil || !ok || text == "" {
		return types.StringNull(), fmt.Errorf("invalid assignment selection %s", key)
	}
	return types.StringValue(text), nil
}

func (m *AssignmentResourceModel) decodeControlOptions() error {
	if !m.DHCP6IdentityAssociationID.IsNull() {
		if err := assignmentAssociationIDValue(m.DHCP6IdentityAssociationID.ValueString()); err != nil {
			return err
		}
	}
	if !m.VLANHardwareFiltering.IsNull() {
		value, ok := map[string]string{"0": "enable", "1": "disable", "2": "default"}[m.VLANHardwareFiltering.ValueString()]
		if !ok {
			return fmt.Errorf("invalid assignment VLAN hardware filtering option")
		}
		m.VLANHardwareFiltering = types.StringValue(value)
	}
	if !m.DHCP6PrefixLength.IsNull() && m.DHCP6PrefixLength.ValueString() != "none" {
		encoded, err := strconv.Atoi(m.DHCP6PrefixLength.ValueString())
		if err != nil || encoded < 0 || encoded > 16 || strconv.Itoa(encoded) != m.DHCP6PrefixLength.ValueString() {
			return fmt.Errorf("invalid assignment DHCPv6 prefix length encoding")
		}
		m.DHCP6PrefixLength = types.StringValue(strconv.Itoa(64 - encoded))
	}
	return nil
}

func assignmentPrefixLengthWrite(value string) string {
	if value == "none" {
		return value
	}
	length, _ := strconv.Atoi(value)
	return strconv.Itoa(64 - length)
}

func validateAssignmentDevice(current *assignmentAPIResponse, requested string) error {
	selected, err := assignmentControlSelection(current, "if")
	if err != nil {
		return err
	}
	if requested == selected.ValueString() {
		return nil
	}
	var options map[string]json.RawMessage
	if err := json.Unmarshal(current.Raw["if"], &options); err != nil {
		return fmt.Errorf("rebinding requires an API device option map")
	}
	if _, ok := options[requested]; !ok {
		return fmt.Errorf("requested device is not advertised by the assignment API")
	}
	return nil
}

func assignmentControlHexString(api *assignmentAPIResponse, key string, prior types.String) (types.String, error) {
	value, err := assignmentControlString(api, key)
	if err != nil || value.IsNull() || value.ValueString() == "" {
		return value, err
	}
	parsed, ok := assignmentHexValue(value.ValueString())
	if !ok {
		return types.StringNull(), fmt.Errorf("invalid assignment hexadecimal field %s: expected a 0x-prefixed value from 0 through 0x7fffffffffffffff (appliance signed-64-bit serialization limit)", key)
	}
	if !prior.IsNull() && !prior.IsUnknown() && prior.ValueString() != "" {
		previous, ok := assignmentHexValue(prior.ValueString())
		if ok && previous.Cmp(parsed) == 0 {
			return prior, nil
		}
	}
	return value, nil
}

func assignmentHexValue(value string) (*big.Int, bool) {
	if !regexp.MustCompile(`^0[xX][a-fA-F0-9]{0,30}$`).MatchString(value) {
		return nil, false
	}
	digits := strings.ToLower(value[2:])
	if digits == "" {
		digits = "0"
	}
	number, ok := new(big.Int).SetString(digits, 16)
	return number, ok && number.IsInt64() && number.Sign() >= 0
}

// IntegerField uses FILTER_VALIDATE_INT without octal flags. On the appliance's
// 64-bit runtime this rejects leading-zero spellings and values above INT64_MAX.
type assignmentAssociationIDValidator struct{}

func (assignmentAssociationIDValidator) Description(context.Context) string {
	return "must be empty or a canonical nonnegative signed-64-bit decimal integer"
}

func (v assignmentAssociationIDValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (assignmentAssociationIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if err := assignmentAssociationIDValue(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid DHCPv6 Association ID", err.Error())
	}
}

func assignmentAssociationIDValue(value string) error {
	if value == "" {
		return nil
	}
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil || number < 0 || strconv.FormatInt(number, 10) != value {
		return fmt.Errorf("DHCPv6 association ID must be empty or a canonical decimal integer from 0 through 9223372036854775807")
	}
	return nil
}

type assignmentHexValidator struct{}

func (assignmentHexValidator) Description(context.Context) string {
	return "must be empty or a 0x-prefixed hexadecimal value no greater than 0x7fffffffffffffff"
}

func (v assignmentHexValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (assignmentHexValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}
	if _, ok := assignmentHexValue(req.ConfigValue.ValueString()); !ok {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid DHCPv6 Hexadecimal ID", "Use a 0x-prefixed hexadecimal value from 0 through 0x7fffffffffffffff, or empty to clear. The appliance serializes this field through a signed 64-bit integer; wider values cannot round-trip safely.")
	}
}
