// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// AssignmentResourceModel is the Terraform state model for an existing base
// interface assignment. The assignment name is both its API key and state ID.
type AssignmentResourceModel struct {
	Device                        types.String `tfsdk:"device"`
	VLANHardwareFiltering         types.String `tfsdk:"vlan_hardware_filtering"`
	DHCP6PrefixLength             types.String `tfsdk:"dhcp6_prefix_length"`
	DHCP6PrefixOnly               types.Bool   `tfsdk:"dhcp6_prefix_only"`
	DHCP6SendPrefixHint           types.Bool   `tfsdk:"dhcp6_send_prefix_hint"`
	DHCP6PrefixID                 types.String `tfsdk:"dhcp6_prefix_id"`
	DHCP6InterfaceID              types.String `tfsdk:"dhcp6_interface_id"`
	DHCP6IdentityAssociationID    types.String `tfsdk:"dhcp6_identity_association_id"`
	Description                   types.String `tfsdk:"description"`
	MTU                           types.Int64  `tfsdk:"mtu"`
	MSS                           types.Int64  `tfsdk:"mss"`
	BlockPrivate                  types.Bool   `tfsdk:"block_private"`
	BlockBogons                   types.Bool   `tfsdk:"block_bogons"`
	OverrideHardwareSettings      types.Bool   `tfsdk:"override_hardware_settings"`
	DisableChecksumOffloading     types.Bool   `tfsdk:"disable_checksum_offloading"`
	DisableSegmentationOffloading types.Bool   `tfsdk:"disable_segmentation_offloading"`
	DisableLargeReceiveOffloading types.Bool   `tfsdk:"disable_large_receive_offloading"`
	IPv6TrackInterface            types.String `tfsdk:"ipv6_track_interface"`
	IPv6PrefixID                  types.String `tfsdk:"ipv6_prefix_id"`
	IPv6InterfaceID               types.String `tfsdk:"ipv6_interface_id"`
	IPv6PrefixRange               types.String `tfsdk:"ipv6_prefix_range"`
	IPv6IdentityAssociationID     types.String `tfsdk:"ipv6_identity_association_id"`
	IPv6AllowManualDHCPv6         types.Bool   `tfsdk:"ipv6_allow_manual_dhcpv6"`
	ID                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	Enabled                       types.Bool   `tfsdk:"enabled"`
	IPv4Mode                      types.String `tfsdk:"ipv4_mode"`
	IPv4Address                   types.String `tfsdk:"ipv4_address"`
	IPv4PrefixLength              types.Int64  `tfsdk:"ipv4_prefix_length"`
	IPv4Gateway                   types.String `tfsdk:"ipv4_gateway"`
	IPv6Mode                      types.String `tfsdk:"ipv6_mode"`
	IPv6Address                   types.String `tfsdk:"ipv6_address"`
	IPv6PrefixLength              types.Int64  `tfsdk:"ipv6_prefix_length"`
	IPv6Gateway                   types.String `tfsdk:"ipv6_gateway"`
	DHCPHostname                  types.String `tfsdk:"dhcp_hostname"`
	DHCPRejectFrom                types.Set    `tfsdk:"dhcp_reject_from"`
	DHCPHonorMTU                  types.Bool   `tfsdk:"dhcp_honor_mtu"`
	DHCPVLANPriority              types.String `tfsdk:"dhcp_vlan_priority"`
	DHCPSendOptions               types.String `tfsdk:"dhcp_send_options"`
	DHCPRequestOptions            types.String `tfsdk:"dhcp_request_options"`
	DHCP6VLANPriority             types.String `tfsdk:"dhcp6_vlan_priority"`
	DHCP6RequestDNS               types.Bool   `tfsdk:"dhcp6_request_dns"`
	DHCP6InformationOnly          types.Bool   `tfsdk:"dhcp6_information_only"`
	DHCP6RapidCommit              types.Bool   `tfsdk:"dhcp6_rapid_commit"`
	DHCP6SendOptions              types.String `tfsdk:"dhcp6_send_options"`
	DHCP6RequestOptions           types.String `tfsdk:"dhcp6_request_options"`
}

// assignmentAPIResponse retains the complete assignment object. Update
// payloads are derived from this raw object so fields outside Terraform's safe
// first-release surface are carried forward instead of reset.
type assignmentAPIResponse struct {
	Raw map[string]json.RawMessage
}

func (a *assignmentAPIResponse) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &a.Raw)
}

type assignmentAPIRequest map[string]any

var assignmentRequiredFields = []string{
	"if", "enable", "type4", "ipaddr", "type6", "ipaddrv6", "gateway", "gatewayv6",
	"dhcphostname", "dhcprejectfrom", "dhcphonourmtu", "dhcpvlanprio",
	"dhcp_send_options", "dhcp_request_options", "dhcp6vlanprio",
	"dhcp6_request_dns", "dhcp6-information-only", "dhcp6_rapid_commit",
	"dhcp6_send_options", "dhcp6_request_options",
}

func (a *assignmentAPIResponse) missingRequiredFields() []string {
	missing := make([]string, 0)
	for _, field := range assignmentRequiredFields {
		if _, ok := a.Raw[field]; !ok {
			missing = append(missing, field)
		}
	}
	return missing
}

func (a *assignmentAPIResponse) toRequest() (assignmentAPIRequest, error) {
	request := make(assignmentAPIRequest, len(a.Raw))
	for key, raw := range a.Raw {
		switch key {
		case "identifier", "pending_action", "icon", "optgroup":
			continue
		}
		value, err := assignmentRequestValue(raw)
		if err != nil {
			return nil, fmt.Errorf("normalize assignment field %q: %w", key, err)
		}
		request[key] = value
	}
	return request, nil
}

func assignmentRequestValue(raw json.RawMessage) (any, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return "", nil
	}
	if trimmed[0] == '[' {
		if selected, ok := selectedIndexesInList(trimmed); ok {
			return strings.Join(selected, ","), nil
		}
	}
	if trimmed[0] != '{' {
		var value any
		if err := json.Unmarshal(trimmed, &value); err != nil {
			return nil, err
		}
		if values, ok := value.([]any); ok && len(values) == 0 {
			return "", nil
		}
		return value, nil
	}

	if selected, ok := selectedValuesInOrder(trimmed); ok {
		return strings.Join(selected, ","), nil
	}

	var value any
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func selectedValuesInOrder(data []byte) ([]string, bool) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, false
	}
	selected := make([]string, 0)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, false
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, false
		}
		var rawChoice map[string]json.RawMessage
		if err := decoder.Decode(&rawChoice); err != nil {
			return nil, false
		}
		rawSelected, ok := rawChoice["selected"]
		if !ok {
			return nil, false
		}
		var value json.Number
		choiceDecoder := json.NewDecoder(bytes.NewReader(rawSelected))
		choiceDecoder.UseNumber()
		if err := choiceDecoder.Decode(&value); err != nil {
			var stringValue string
			if err := json.Unmarshal(rawSelected, &stringValue); err != nil {
				return nil, false
			}
			value = json.Number(stringValue)
		}
		if value.String() == "1" {
			selected = append(selected, key)
		}
	}
	if _, err := decoder.Token(); err != nil {
		return nil, false
	}
	return selected, true
}

func selectedIndexesInList(data []byte) ([]string, bool) {
	var choices []map[string]json.RawMessage
	if err := json.Unmarshal(data, &choices); err != nil || len(choices) == 0 {
		return nil, false
	}
	selected := make([]string, 0)
	for index, choice := range choices {
		rawSelected, ok := choice["selected"]
		if !ok {
			return nil, false
		}
		var selectedValue any
		if err := json.Unmarshal(rawSelected, &selectedValue); err != nil {
			return nil, false
		}
		if fmt.Sprint(selectedValue) == "1" {
			selected = append(selected, strconv.Itoa(index))
		}
	}
	return selected, true
}

func (m *AssignmentResourceModel) applyToAPI(ctx context.Context, current *assignmentAPIResponse) (*assignmentAPIRequest, error) {
	var observed AssignmentResourceModel
	if err := observed.readControls(current); err != nil {
		return nil, err
	}
	request, err := current.toRequest()
	if err != nil {
		return nil, err
	}
	setString := func(key string, value types.String) {
		if !value.IsNull() && !value.IsUnknown() {
			request[key] = value.ValueString()
		}
	}
	setBool := func(key string, value types.Bool) {
		if !value.IsNull() && !value.IsUnknown() {
			request[key] = opnsense.BoolToString(value.ValueBool())
		}
	}

	if err := m.applyControls(current, request); err != nil {
		return nil, err
	}
	setBool("enable", m.Enabled)
	setAssignmentMode := func(key string, value types.String) {
		if !value.IsNull() && !value.IsUnknown() {
			mode := value.ValueString()
			if mode == "none" {
				mode = ""
			}
			request[key] = mode
		}
	}
	setAssignmentMode("type4", m.IPv4Mode)
	setString("gateway", m.IPv4Gateway)
	setAssignmentMode("type6", m.IPv6Mode)
	setString("track6-interface", m.IPv6TrackInterface)
	setString("track6-prefix-id", m.IPv6PrefixID)
	setString("track6_ifid", m.IPv6InterfaceID)
	setString("track6_prefix_range", m.IPv6PrefixRange)
	setString("track6_assoc_pd", m.IPv6IdentityAssociationID)
	setBool("dhcpd6track6allowoverride", m.IPv6AllowManualDHCPv6)
	setString("gatewayv6", m.IPv6Gateway)
	setString("dhcphostname", m.DHCPHostname)
	if !m.DHCPRejectFrom.IsNull() && !m.DHCPRejectFrom.IsUnknown() {
		var values []string
		if diags := m.DHCPRejectFrom.ElementsAs(ctx, &values, false); diags.HasError() {
			return nil, fmt.Errorf("convert dhcp_reject_from")
		}
		sort.Strings(values)
		request["dhcprejectfrom"] = strings.Join(values, ",")
	}
	setBool("dhcphonourmtu", m.DHCPHonorMTU)
	setString("dhcpvlanprio", m.DHCPVLANPriority)
	setString("dhcp_send_options", m.DHCPSendOptions)
	setString("dhcp_request_options", m.DHCPRequestOptions)
	setString("dhcp6vlanprio", m.DHCP6VLANPriority)
	setBool("dhcp6_request_dns", m.DHCP6RequestDNS)
	setBool("dhcp6-information-only", m.DHCP6InformationOnly)
	setBool("dhcp6_rapid_commit", m.DHCP6RapidCommit)
	setString("dhcp6_send_options", m.DHCP6SendOptions)
	setString("dhcp6_request_options", m.DHCP6RequestOptions)

	if !m.IPv4Mode.IsNull() && !m.IPv4Mode.IsUnknown() {
		if m.IPv4Mode.ValueString() == "staticv4" {
			request["ipaddr"] = fmt.Sprintf("%s/%d", m.IPv4Address.ValueString(), m.IPv4PrefixLength.ValueInt64())
		} else {
			request["ipaddr"] = ""
			request["gateway"] = ""
		}
	}
	if !m.IPv6Mode.IsNull() && !m.IPv6Mode.IsUnknown() {
		if m.IPv6Mode.ValueString() == "staticv6" {
			request["ipaddrv6"] = fmt.Sprintf("%s/%d", m.IPv6Address.ValueString(), m.IPv6PrefixLength.ValueInt64())
		} else {
			request["ipaddrv6"] = ""
			request["gatewayv6"] = ""
		}
	}
	if mode, _ := request["type6"].(string); mode == "idassoc6" || mode == "track6" {
		for _, key := range assignmentDelegationAPIFields {
			if _, ok := current.Raw[key]; !ok {
				return nil, &assignmentCompatibilityError{detail: "Assignment API response is missing delegation field " + key + "."}
			}
		}
		parent, _ := request["track6-interface"].(string)
		if parent == "" || parent == m.Name.ValueString() {
			return nil, fmt.Errorf("identity association and track-interface modes require a distinct parent interface")
		}
	}
	return &request, nil
}

func (m *AssignmentResourceModel) fromAPI(_ context.Context, a *assignmentAPIResponse, name string) error {
	if err := m.readControls(a); err != nil {
		return err
	}
	m.IPv6TrackInterface = types.StringValue(a.stringValue("track6-interface"))
	m.IPv6PrefixID = types.StringValue(a.stringValue("track6-prefix-id"))
	m.IPv6InterfaceID = types.StringValue(a.stringValue("track6_ifid"))
	m.IPv6PrefixRange = types.StringValue(a.stringValue("track6_prefix_range"))
	m.IPv6IdentityAssociationID = types.StringValue(a.stringValue("track6_assoc_pd"))
	m.IPv6AllowManualDHCPv6 = types.BoolValue(opnsense.StringToBool(a.stringValue("dhcpd6track6allowoverride")))
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.stringValue("enable")))
	m.IPv4Mode = types.StringValue(assignmentModeFromAPI(a.stringValue("type4")))
	m.IPv4Gateway = types.StringValue(a.stringValue("gateway"))
	m.IPv6Mode = types.StringValue(assignmentModeFromAPI(a.stringValue("type6")))
	m.IPv6Gateway = types.StringValue(a.stringValue("gatewayv6"))
	m.DHCPHostname = types.StringValue(a.stringValue("dhcphostname"))
	m.DHCPRejectFrom = stringSet(a.stringListValue("dhcprejectfrom"))
	m.DHCPHonorMTU = types.BoolValue(opnsense.StringToBool(a.stringValue("dhcphonourmtu")))
	m.DHCPVLANPriority = types.StringValue(a.stringValue("dhcpvlanprio"))
	m.DHCPSendOptions = types.StringValue(a.stringValue("dhcp_send_options"))
	m.DHCPRequestOptions = types.StringValue(a.stringValue("dhcp_request_options"))
	m.DHCP6VLANPriority = types.StringValue(a.stringValue("dhcp6vlanprio"))
	m.DHCP6RequestDNS = types.BoolValue(opnsense.StringToBool(a.stringValue("dhcp6_request_dns")))
	m.DHCP6InformationOnly = types.BoolValue(opnsense.StringToBool(a.stringValue("dhcp6-information-only")))
	m.DHCP6RapidCommit = types.BoolValue(opnsense.StringToBool(a.stringValue("dhcp6_rapid_commit")))
	m.DHCP6SendOptions = types.StringValue(a.stringValue("dhcp6_send_options"))
	m.DHCP6RequestOptions = types.StringValue(a.stringValue("dhcp6_request_options"))

	var err error
	m.IPv4Address, m.IPv4PrefixLength, err = splitAssignmentCIDR(a.stringValue("ipaddr"), false)
	if err != nil {
		return err
	}
	m.IPv6Address, m.IPv6PrefixLength, err = splitAssignmentCIDR(a.stringValue("ipaddrv6"), true)
	if err != nil {
		return err
	}
	return nil
}

func assignmentModeFromAPI(mode string) string {
	if mode == "" {
		return "none"
	}
	return mode
}

func (a *assignmentAPIResponse) stringValue(key string) string {
	raw, ok := a.Raw[key]
	if !ok {
		return ""
	}
	value, err := assignmentRequestValue(raw)
	if err != nil || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func (a *assignmentAPIResponse) stringListValue(key string) []string {
	value := a.stringValue(key)
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitAssignmentCIDR(value string, ipv6 bool) (types.String, types.Int64, error) {
	if value == "" {
		return types.StringValue(""), types.Int64Null(), nil
	}
	prefix, err := netip.ParsePrefix(value)
	if err != nil || prefix.Addr().Is6() != ipv6 {
		return types.String{}, types.Int64{}, fmt.Errorf("OPNsense returned invalid interface address %q", value)
	}
	return types.StringValue(prefix.Addr().String()), types.Int64Value(int64(prefix.Bits())), nil
}

func stringSet(values []string) types.Set {
	attrs := make([]attr.Value, len(values))
	for i, value := range values {
		attrs[i] = types.StringValue(value)
	}
	return types.SetValueMust(types.StringType, attrs)
}
