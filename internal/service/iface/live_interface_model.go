// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// liveShape is both the public allowlist and its exact typed null contract.
// Unknown source fields never enter Terraform values or diagnostics.
type liveShape struct {
	strings, bools, ints, sets []string
	objects                    map[string]liveShape
	lists                      map[string]liveShape
}

var (
	diagnosticAddress4 = liveShape{strings: []string{"ipaddr"}, ints: []string{"subnetbits"}, bools: []string{"tunnel"}}
	diagnosticAddress6 = liveShape{strings: []string{"ipaddr"}, ints: []string{"subnetbits"}, bools: []string{"tunnel", "autoconf", "deprecated", "link_local", "tentative"}}
	overviewAddress    = liveShape{strings: []string{"ipaddr", "vhid", "carp_status", "adv_base", "adv_skew", "peer", "peer6"}}
	diagnosticShape    = liveShape{
		strings: []string{"device", "media", "media_raw", "macaddr", "status"},
		bools:   []string{"is_physical"}, ints: []string{"mtu"},
		sets:  []string{"flags", "capabilities", "options", "supported_media", "groups"},
		lists: map[string]liveShape{"ipv4": diagnosticAddress4, "ipv6": diagnosticAddress6},
	}
)

var overviewShape = liveShape{
	strings: []string{"identifier", "description", "device", "link_type", "addr4", "addr6", "macaddr", "macaddr_hw", "mtu", "status", "media", "media_raw", "vlan_tag", "lagg_proto", "lagg_hash"},
	bools:   []string{"enabled", "is_physical"},
	sets:    []string{"flags", "capabilities", "options", "supported_media", "groups", "gateways", "routes", "nd6_flags", "ifctl_nameserver", "ifctl_router", "ifctl_prefix", "ifctl_searchdomain"},
	lists:   map[string]liveShape{"ipv4": overviewAddress, "ipv6": overviewAddress},
	objects: map[string]liveShape{
		"vlan":            {strings: []string{"tag", "proto", "pcp", "parent"}},
		"lagg_options":    {strings: []string{"flowid_shift"}, sets: []string{"flags"}},
		"lagg_statistics": {strings: []string{"active_ports", "flapping"}},
	},
}

func (s liveShape) attributes() map[string]dsschema.Attribute {
	a := map[string]dsschema.Attribute{}
	for _, k := range s.strings {
		a[k] = dsschema.StringAttribute{Computed: true, MarkdownDescription: liveFieldDescription(k, false)}
	}
	for _, k := range s.bools {
		a[k] = dsschema.BoolAttribute{Computed: true, MarkdownDescription: liveFieldDescription(k, false)}
	}
	for _, k := range s.ints {
		a[k] = dsschema.Int64Attribute{Computed: true, MarkdownDescription: liveFieldDescription(k, false)}
	}
	for _, k := range s.sets {
		a[k] = dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: liveFieldDescription(k, true) + " Set members are not trimmed or split."}
	}
	for k, child := range s.objects {
		a[k] = dsschema.SingleNestedAttribute{Computed: true, MarkdownDescription: liveFieldDescription(k, false), Attributes: child.attributes()}
	}
	for k, child := range s.lists {
		a[k] = dsschema.ListNestedAttribute{Computed: true, MarkdownDescription: liveFieldDescription(k, true), NestedObject: dsschema.NestedAttributeObject{Attributes: child.attributes()}}
	}
	return a
}

func (s liveShape) types() map[string]attr.Type {
	result := map[string]attr.Type{}
	for k, a := range s.attributes() {
		result[k] = a.GetType()
	}
	return result
}

func (s liveShape) decode(ctx context.Context, raw json.RawMessage, diagnostic bool) (types.Object, error) {
	var row map[string]json.RawMessage
	if err := json.Unmarshal(raw, &row); err != nil || row == nil {
		return types.Object{}, fmt.Errorf("expected interface object")
	}
	values := map[string]attr.Value{}
	for k, typ := range s.types() {
		v, present, err := liveWireValue(row, k)
		if err != nil {
			return types.Object{}, err
		}
		absent := !present || string(v) == "null"
		switch typ.(type) {
		case basetypes.StringType:
			values[k] = types.StringNull()
			if !absent {
				var value string
				if json.Unmarshal(v, &value) != nil {
					return types.Object{}, fmt.Errorf("invalid string field %s", k)
				}
				values[k] = types.StringValue(value)
			}
		case basetypes.BoolType:
			values[k] = types.BoolNull()
			if !absent {
				var value bool
				if json.Unmarshal(v, &value) != nil {
					return types.Object{}, fmt.Errorf("invalid boolean field %s", k)
				}
				values[k] = types.BoolValue(value)
			}
		case basetypes.Int64Type:
			values[k] = types.Int64Null()
			if !absent {
				var value int64
				if diagnostic && k == "mtu" {
					var str string
					if json.Unmarshal(v, &str) != nil {
						return types.Object{}, fmt.Errorf("invalid decimal string field %s", k)
					}
					var err error
					value, err = strconv.ParseInt(str, 10, 64)
					if err != nil {
						return types.Object{}, fmt.Errorf("invalid decimal string field %s", k)
					}
				} else if json.Unmarshal(v, &value) != nil {
					return types.Object{}, fmt.Errorf("invalid integer field %s", k)
				}
				values[k] = types.Int64Value(value)
			}
		case basetypes.SetType:
			members := []string{}
			if !absent && json.Unmarshal(v, &members) != nil {
				return types.Object{}, fmt.Errorf("invalid string collection field %s", k)
			}
			// Decode elements individually: encoding/json otherwise accepts null as an empty string.
			if !absent {
				var elements []json.RawMessage
				if json.Unmarshal(v, &elements) != nil {
					return types.Object{}, fmt.Errorf("invalid collection field %s", k)
				}
				for _, element := range elements {
					if string(element) == "null" {
						return types.Object{}, fmt.Errorf("invalid string member in %s", k)
					}
				}
			}
			// Source arrays can repeat values (notably routes). Terraform sets
			// require unique members; preserve each exact string without normalization.
			unique := make([]string, 0, len(members))
			seen := make(map[string]struct{}, len(members))
			for _, member := range members {
				if _, exists := seen[member]; !exists {
					seen[member] = struct{}{}
					unique = append(unique, member)
				}
			}
			set, diags := types.SetValueFrom(ctx, types.StringType, unique)
			if diags.HasError() {
				return types.Object{}, fmt.Errorf("invalid set field %s", k)
			}
			values[k] = set
		case basetypes.ObjectType:
			child := s.objects[k]
			values[k] = types.ObjectNull(child.types())
			if !absent {
				value, err := child.decode(ctx, v, false)
				if err != nil {
					return types.Object{}, fmt.Errorf("%s: %w", k, err)
				}
				values[k] = value
			}
		case basetypes.ListType:
			child := s.lists[k]
			members := []json.RawMessage{}
			if !absent && json.Unmarshal(v, &members) != nil {
				return types.Object{}, fmt.Errorf("invalid address list %s", k)
			}
			objects := []attr.Value{}
			for _, member := range members {
				object, err := child.decode(ctx, member, false)
				if err != nil {
					return types.Object{}, fmt.Errorf("%s: %w", k, err)
				}
				objects = append(objects, object)
			}
			sort.Slice(objects, func(i, j int) bool {
				return liveAddressTuple(objects[i].(types.Object)) < liveAddressTuple(objects[j].(types.Object))
			})
			value, diags := types.ListValue(types.ObjectType{AttrTypes: child.types()}, objects)
			if diags.HasError() {
				return types.Object{}, fmt.Errorf("invalid address list %s", k)
			}
			values[k] = value
		}
	}
	result, diags := types.ObjectValue(s.types(), values)
	if diags.HasError() {
		return types.Object{}, fmt.Errorf("invalid interface values")
	}
	return result, nil
}

func liveWireKey(public string) string {
	switch public {
	case "link_local":
		return "link-local"
	case "carp_status":
		return "status"
	case "adv_base":
		return "advbase"
	case "adv_skew":
		return "advskew"
	case "active_ports":
		return "active ports"
	case "lagg_proto":
		return "laggproto"
	case "lagg_hash":
		return "lagghash"
	case "lagg_options":
		return "laggoptions"
	case "lagg_statistics":
		return "laggstatistics"
	case "ifctl_nameserver":
		return "ifctl.nameserver"
	case "ifctl_router":
		return "ifctl.router"
	case "ifctl_prefix":
		return "ifctl.prefix"
	case "ifctl_searchdomain":
		return "ifctl.searchdomain"
	default:
		return public
	}
}

func liveWireValue(row map[string]json.RawMessage, key string) (json.RawMessage, bool, error) {
	if key == "nd6_flags" {
		raw, ok := row["nd6"]
		if !ok || string(raw) == "null" {
			return nil, false, nil
		}
		var nd6 map[string]json.RawMessage
		if json.Unmarshal(raw, &nd6) != nil || nd6 == nil {
			return nil, false, fmt.Errorf("invalid nd6 object")
		}
		raw, ok = nd6["flags"]
		return raw, ok, nil
	}
	raw, ok := row[liveWireKey(key)]
	return raw, ok, nil
}

// liveAddressTuple uses only public scalar values, never framework debug strings.
// Sorted attribute names pin the complete tuple; JSON escaping distinguishes
// embedded delimiters, empty strings and null without depending on input order.
func liveAddressTuple(object types.Object) string {
	attrs := object.Attributes()
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	tuple := make([]any, 0, len(keys))
	for _, key := range keys {
		value := attrs[key]
		if value.IsNull() {
			tuple = append(tuple, nil)
			continue
		}
		switch v := value.(type) {
		case types.String:
			tuple = append(tuple, v.ValueString())
		case types.Bool:
			tuple = append(tuple, v.ValueBool())
		case types.Int64:
			tuple = append(tuple, v.ValueInt64())
		}
	}
	// These tuple members are only nil, string, bool and int64; encoding cannot fail.
	encoded, _ := json.Marshal(tuple)
	return string(encoded)
}

var liveFieldDescriptions = map[string]string{
	"device":             "Exact kernel device name; the stable selector, not an assignment identifier or description.",
	"identifier":         "Assignment identifier; empty for unassigned interfaces. Not a device alias.",
	"description":        "Assignment description or localized unassigned label. Locale dependent; never use as a stable identity or for_each key.",
	"link_type":          "IPv4 configuration mode (for example dhcp, static or none), not link media. IPv6 mode link_typev6 is outside this public allowlist.",
	"addr4":              "Explicit primary IPv4 observation from overview; address-list order does not identify the primary address.",
	"addr6":              "Explicit primary IPv6 observation from overview; address-list order does not identify the primary address.",
	"ipaddr":             "Diagnostics: bare address with separate subnetbits. Overview: assigned address includes CIDR prefix; unassigned address may remain bare.",
	"subnetbits":         "Diagnostic address prefix length, separate from the bare ipaddr.",
	"mtu":                "Diagnostics expose Int64 parsed from the wire decimal string; overview retains the String observation.",
	"carp_status":        "CARP state from the address wire field status; absent metadata is null.",
	"adv_base":           "CARP advertisement base from wire advbase; absent metadata is null.",
	"adv_skew":           "CARP advertisement skew from wire advskew; absent metadata is null.",
	"link_local":         "IPv6 link-local flag from wire link-local.",
	"nd6_flags":          "IPv6 neighbor-discovery flags from nested wire nd6.flags.",
	"lagg_proto":         "LAGG protocol from wire laggproto; null when absent.",
	"lagg_hash":          "LAGG hash configuration from wire lagghash; null when absent.",
	"lagg_options":       "Optional LAGG options from wire laggoptions; null when absent.",
	"lagg_statistics":    "Optional LAGG statistics from wire laggstatistics; null when absent.",
	"active_ports":       "LAGG active port count as supplied by wire active ports (String).",
	"ifctl_nameserver":   "DHCP-derived wire ifctl.nameserver members, preserved without trimming or splitting.",
	"ifctl_router":       "DHCP-derived wire ifctl.router members, preserved without trimming or splitting.",
	"ifctl_prefix":       "DHCP-derived wire ifctl.prefix members, preserved without trimming or splitting.",
	"ifctl_searchdomain": "DHCP-derived wire ifctl.searchdomain members, preserved without trimming or splitting.",
	"ipv4":               "IPv4 observations sorted by complete public tuple, not primary priority; missing collection is empty.",
	"ipv6":               "IPv6 observations sorted by complete public tuple, not primary priority; missing collection is empty.",
	"macaddr_hw":         "Hardware MAC address observation, distinct from the current macaddr.",
	"macaddr":            "Current MAC address observation.",
	"media":              "Reported link media; distinct from the overview IPv4 configuration mode link_type.",
	"media_raw":          "Unnormalized link media observation.",
	"status":             "Runtime interface status observation.",
	"enabled":            "Assignment enabled observation; absent on unassigned rows remains null.",
	"is_physical":        "Whether the upstream observation identifies a physical interface.",
	"flags":              "Reported flags; exact duplicates collapse and distinct members remain unchanged.",
	"capabilities":       "Reported interface capabilities.",
	"options":            "Reported enabled interface options.",
	"supported_media":    "Reported supported link media.",
	"groups":             "Reported interface group memberships.",
	"gateways":           "Reported overview gateways.",
	"routes":             "Reported overview routes; exact duplicate strings collapse.",
	"vlan":               "Optional VLAN observations; null when absent.",
	"vlan_tag":           "Reported overview VLAN tag; may be explicitly null on assigned non-VLAN interfaces.",
}

func liveFieldDescription(key string, collection bool) string {
	description := liveFieldDescriptions[key]
	if description == "" {
		description = "Upstream " + key + " observation."
	}
	if collection {
		return description + " Missing or null collections are known empty."
	}
	return description + " Absent optional values remain null."
}
