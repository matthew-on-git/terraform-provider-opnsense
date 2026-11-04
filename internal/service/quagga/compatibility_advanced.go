// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// These fields are pinned to released BGP model 1.1.2/1.1.5, not inferred
// from the oldest supported plugin. Older packages need field-presence proof.
var advancedNeighborFields = map[string]bool{
	"remote_as_mode": true, "localas": true, "password": true, "weight": true, "localip": true, "linklocalinterface": true, "nexthopselfall": true, "multihop": true, "rrclient": true, "soft_reconfiguration_inbound": true, "bfd": true, "bfd_strict": true, "connecttimer": true, "defaultoriginate": true, "asoverride": true, "removeprivateas": true, "allowas_in": true, "disable_connected_check": true, "attributeunchanged": true, "capabilities": true,
}

var advancedEnums = map[string][]string{
	"remote_as_mode":     {"", "internal", "external"},
	"removeprivateas":    {"", "remove-private-AS", "remove-private-AS all", "remove-private-AS replace-AS", "remove-private-AS all replace-AS"},
	"allowas_in":         {"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "origin"},
	"attributeunchanged": {"", "as-path", "next-hop", "med"},
	"capabilities":       {"extended-nexthop", "dynamic", "fqdn", "link-local"},
}
var frrPolicyName = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)

func advancedString(_ context.Context, value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func advancedBool(_ context.Context, value types.Bool) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := "0"
	if value.ValueBool() {
		v = "1"
	}
	return &v
}

func advancedInt64(_ context.Context, value types.Int64) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := strconv.FormatInt(value.ValueInt64(), 10)
	return &v
}

func advancedSet(ctx context.Context, value types.Set) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := setToCSV(ctx, value)
	return &v
}

func optionalFRRInt(value types.Int64) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return strconv.FormatInt(value.ValueInt64(), 10)
}

func readAdvancedString(present bool, value string) types.String {
	if !present {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func readAdvancedBool(present bool, value string) types.Bool {
	if !present {
		return types.BoolNull()
	}
	return types.BoolValue(value == "1")
}

func readAdvancedInt64(present bool, value string) types.Int64 {
	if !present || value == "" {
		return types.Int64Null()
	}
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return types.Int64Null()
	}
	return types.Int64Value(v)
}

func readAdvancedSet(present bool, value []string) types.Set {
	if !present {
		return types.SetNull(types.StringType)
	}
	return sliceToSet(value)
}

func advancedSupported(c frrCapability, wire string) bool {
	return c.Version == "1.55" || c.Version == "2.0" || c.Version == "2.0_1" || c.Fields[wire]
}

func validAdvancedEnum(wire, value string) bool {
	for _, option := range advancedEnums[wire] {
		if option == value {
			return true
		}
	}
	return false
}

func validateAdvancedWire(fields map[string]string) error {
	for wire := range advancedEnums {
		value, exists := fields[wire]
		if !exists {
			continue
		}
		values := []string{value}
		if wire == "capabilities" {
			if value == "" {
				continue
			}
			values = strings.Split(value, ",")
		}
		for _, option := range values {
			if !validAdvancedEnum(wire, option) {
				return fmt.Errorf("FRR %s selects an unsupported source option", wire)
			}
		}
	}
	for wire, bounds := range map[string][2]int64{"localas": {1, 4294967295}, "weight": {0, 65535}, "connecttimer": {1, 65000}, "remoteas": {1, 4294967295}, "id": {1, 65535}} {
		raw, exists := fields[wire]
		if !exists || raw == "" {
			continue
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < bounds[0] || n > bounds[1] {
			return fmt.Errorf("FRR %s is outside the released source numeric range", wire)
		}
	}
	if name, exists := fields["name"]; exists && !frrPolicyName.MatchString(name) {
		return fmt.Errorf("FRR policy or peer-group name does not match the released source pattern")
	}
	if raw, exists := fields["localip"]; exists && raw != "" {
		if _, err := netip.ParseAddr(raw); err != nil {
			return fmt.Errorf("FRR local_ip must be a host address")
		}
	}
	if fields["password"] != "" && fields["localip"] == "" {
		return fmt.Errorf("FRR local_ip is required when authentication input is configured")
	}
	if fields["bfd_strict"] == "1" && fields["bfd"] != "1" {
		return fmt.Errorf("FRR bfd_strict requires bfd to be enabled")
	}
	if fields["nexthopselfall"] == "1" && fields["nexthopself"] != "1" {
		return fmt.Errorf("FRR next_hop_self_all requires next_hop_self to be enabled")
	}
	if raw, exists := fields["listenranges"]; exists && raw != "" {
		for _, value := range strings.Split(raw, ",") {
			p, err := netip.ParsePrefix(value)
			if err != nil || p != p.Masked() {
				return fmt.Errorf("FRR listen_ranges requires strict CIDR networks")
			}
		}
	}
	return nil
}

// A key with null or malformed shape is not authoritative capability proof.
func validAdvancedTemplateField(wire string, raw json.RawMessage) bool {
	if string(raw) == "null" {
		return false
	}
	if wire == "password" {
		var scalar *string
		return json.Unmarshal(raw, &scalar) == nil && scalar != nil
	}
	var probe bgpNeighborAPIResponse
	encoded, err := json.Marshal(map[string]json.RawMessage{wire: raw})
	if err != nil || json.Unmarshal(encoded, &probe) != nil {
		return false
	}
	fields := map[string]string{}
	v := reflect.ValueOf(probe)
	f := v.FieldByName(map[string]string{"localas": "LocalAS", "weight": "Weight", "connecttimer": "ConnectTimer", "remote_as_mode": "RemoteASMode", "removeprivateas": "RemovePrivateAS", "allowas_in": "AllowASIn", "attributeunchanged": "AttributeUnchanged", "localip": "LocalIP"}[wire])
	if f.IsValid() {
		fields[wire] = f.String()
	}
	if wire == "capabilities" {
		fields[wire] = strings.Join(probe.Capabilities, ",")
	}
	if validateAdvancedWire(fields) != nil {
		return false
	}
	switch wire {
	case "nexthopselfall", "multihop", "rrclient", "soft_reconfiguration_inbound", "bfd", "bfd_strict", "defaultoriginate", "asoverride", "disable_connected_check":
		var flag string
		return json.Unmarshal(raw, &flag) == nil && (flag == "0" || flag == "1")
	}
	return true
}
