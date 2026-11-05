// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/netip"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type (
	frrSelectedMap     string
	frrSelectedMapList []string
)

func decodeFRRSelection(raw []byte) ([]string, error) {
	value := bytes.TrimSpace(raw)
	if bytes.Equal(value, []byte("[]")) {
		return []string{}, nil
	}
	if len(value) > 0 && value[0] == '"' {
		var csv string
		if err := json.Unmarshal(value, &csv); err != nil {
			return nil, fmt.Errorf("malformed FRR selector")
		}
		if csv == "" {
			return []string{}, nil
		}
		values := strings.Split(csv, ",")
		seen := map[string]bool{}
		for _, id := range values {
			if seen[id] {
				return nil, fmt.Errorf("FRR selector contains duplicate members")
			}
			seen[id] = true
			if id == "" || strings.TrimSpace(id) != id {
				return nil, fmt.Errorf("malformed FRR selector member")
			}
		}
		return values, nil
	}
	var options map[string]struct {
		Selected json.RawMessage `json:"selected"`
	}
	if len(value) == 0 || value[0] != '{' || json.Unmarshal(value, &options) != nil {
		return nil, fmt.Errorf("FRR selector must be a selection map, scalar CSV, or empty array")
	}
	selected := []string{}
	for key, entry := range options {
		selection := strings.Trim(string(entry.Selected), "\"")
		if selection != "0" && selection != "1" {
			return nil, fmt.Errorf("FRR selector contains malformed selected flag")
		}
		if selection == "1" && key != "" {
			if strings.Contains(key, ",") {
				return nil, fmt.Errorf("FRR selector map key must contain one member, not CSV")
			}
			selected = append(selected, key)
		}
	}
	sort.Strings(selected)
	return selected, nil
}

func (s *frrSelectedMapList) UnmarshalJSON(raw []byte) error {
	values, err := decodeFRRSelection(raw)
	if err == nil {
		*s = values
	}
	return err
}

func (s *frrSelectedMap) UnmarshalJSON(raw []byte) error {
	values, err := decodeFRRSelection(raw)
	if err != nil {
		return err
	}
	if len(values) > 1 {
		return fmt.Errorf("FRR single-select field selected multiple options")
	}
	*s = ""
	if len(values) == 1 {
		*s = frrSelectedMap(values[0])
	}
	return nil
}

func validateBGPResponse(value any, capability frrCapability) error {
	v := reflect.ValueOf(value).Elem()
	_, neighborAdvanced := value.(*bgpNeighborAPIResponse)
	_, peerAdvanced := value.(*bgpPeerGroupAPIResponse)
	exempt := func(wire string) bool {
		return (neighborAdvanced && advancedNeighborFields[wire]) || (peerAdvanced && wire == "listenranges")
	}
	presence, ok := value.(interface{ hasField(string) bool })
	if !ok {
		return fmt.Errorf("FRR response lacks strict field presence")
	}
	for i := 0; i < v.NumField(); i++ {
		wire := strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]
		if wire == "-" {
			continue
		}
		if wire == "family" && !capability.Grouped && v.FieldByName("MultiProtocol").IsValid() {
			continue
		}
		if wire == "multiprotocol" && capability.Grouped {
			if presence.hasField(wire) {
				return fmt.Errorf("FRR2 response unexpectedly contains removed multiprotocol field")
			}
			continue
		}
		if exempt(wire) && !advancedSupported(capability, wire) && !presence.hasField(wire) {
			continue
		}
		if !presence.hasField(wire) {
			return fmt.Errorf("FRR %s is absent from read-back; no clear/default was inferred", wire)
		}
	}
	for _, name := range []string{"Enabled", "NextHopSelf", "DefaultOriginate", "MultiProtocol", "GracefulRestart", "NetworkImportCheck", "EnforceFirstAS", "LogNeighborChanges", "EnableCARP", "EnableSyslog", "EnableSNMP", "FirewallRules", "NextHopSelfAll", "Multihop", "RouteReflectorClient", "SoftReconfigurationInbound", "BFD", "BFDStrict", "ASOverride", "DisableConnectedCheck"} {
		field := v.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		if exempt(jsonWire(v.Type(), name)) && !presence.hasField(jsonWire(v.Type(), name)) {
			continue
		}
		raw := field.String()
		if name == "MultiProtocol" && capability.Grouped {
			if raw != "" {
				return fmt.Errorf("FRR2 neighbor unexpectedly returned multiprotocol")
			}
			continue
		}
		if raw != "0" && raw != "1" {
			return fmt.Errorf("FRR %s read-back is missing or not a boolean flag", name)
		}
	}
	if f := v.FieldByName("Family"); f.IsValid() && (capability.Grouped || !v.FieldByName("MultiProtocol").IsValid()) {
		if f.Len() == 0 {
			return fmt.Errorf("FRR read-back has empty address family")
		}
		for i := 0; i < f.Len(); i++ {
			if f.Index(i).String() != "ipv4" && f.Index(i).String() != "ipv6" {
				return fmt.Errorf("FRR read-back has unsupported address family")
			}
		}
	}

	if err := validateFRRFields(value); err != nil {
		return err
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := v.Type().Field(i).Name
		if field.Type() == reflect.TypeFor[frrSelectedMapList]() && field.IsNil() {
			if exempt(strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]) && !presence.hasField(strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]) {
				continue
			}
			if name == "Family" && !capability.Grouped && v.FieldByName("MultiProtocol").IsValid() {
				continue
			}
			return fmt.Errorf("FRR %s selector is missing from read-back", name)
		}
	}
	fields := map[string]string{}
	for i := 0; i < v.NumField(); i++ {
		wire := strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]
		f := v.Field(i)
		if wire != "-" && presence.hasField(wire) {
			if f.Kind() == reflect.String {
				fields[wire] = f.String()
			} else if f.Type() == reflect.TypeFor[frrSelectedMapList]() {
				fields[wire] = strings.Join(f.Interface().(frrSelectedMapList), ",")
			}
		}
	}
	if err := validateAdvancedWire(fields); err != nil {
		return err
	}
	return nil
}

func jsonWire(t reflect.Type, name string) string {
	f, _ := t.FieldByName(name)
	return strings.Split(f.Tag.Get("json"), ",")[0]
}

func validateFRRFields(value any) error {
	v := reflect.ValueOf(value).Elem()
	numeric := map[string][2]int64{"RemoteAS": {1, 4294967295}, "ASNumber": {1, 4294967295}, "Order": {1, 65535}, "SeqNumber": {1, 4294967294}, "Number": {0, 4294967295}, "Keepalive": {1, 1000}, "Holddown": {3, 3000}, "Distance": {1, 255}, "MaximumPaths": {1, 128}, "MaximumPathsIBGP": {1, 128}, "LocalAS": {1, 4294967295}, "Weight": {0, 65535}, "ConnectTimer": {1, 65000}}
	_, neighbor := value.(*bgpNeighborAPIResponse)
	_, community := value.(*bgpCommunityListAPIResponse)
	if community {
		numeric["Number"] = [2]int64{1, 500}
		numeric["SeqNumber"] = [2]int64{10, 99}
	}
	for name, bounds := range numeric {
		field := v.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		raw := field.String()
		required := name == "ASNumber" || name == "Order" || name == "SeqNumber" || name == "Number"
		if raw == "" && !required {
			continue
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < bounds[0] || n > bounds[1] {
			return fmt.Errorf("FRR %s read-back is missing, unsupported or malformed; numeric intent was not defaulted", name)
		}
	}
	for name, domain := range map[string][]string{
		"RemoteASMode": {"", "internal", "external"},
		"Redistribute": {"ospf", "connected", "kernel", "rip", "static"},
		"Profile":      {"traditional", "datacenter"},
		"SyslogLevel":  {"critical", "emergencies", "errors", "alerts", "warnings", "notifications", "informational", "debugging"},
	} {
		if field := v.FieldByName(name); field.IsValid() {
			valid := false
			for _, option := range domain {
				if field.String() == option {
					valid = true
				}
			}
			if !valid {
				return fmt.Errorf("FRR %s read-back selected unsupported option", name)
			}
		}
	}
	if field := v.FieldByName("RouterID"); field.IsValid() && field.String() != "" {
		ip, err := netip.ParseAddr(field.String())
		if err != nil || !ip.Is4() {
			return fmt.Errorf("FRR router ID is not an IPv4 host address")
		}
	}
	if field := v.FieldByName("Networks"); field.IsValid() {
		for i := 0; i < field.Len(); i++ {
			prefix, err := netip.ParsePrefix(field.Index(i).String())
			if err != nil || prefix != prefix.Masked() {
				return fmt.Errorf("FRR networks read-back is not strict CIDR")
			}
		}
	}
	if neighbor {
		if _, err := netip.ParseAddr(v.FieldByName("Address").String()); err != nil {
			return fmt.Errorf("FRR neighbor read-back has missing or malformed transport address")
		}
	}
	if field := v.FieldByName("Name"); field.IsValid() && field.String() == "" {
		return fmt.Errorf("FRR policy/peer-group name is missing in read-back")
	}
	if field := v.FieldByName("Action"); field.IsValid() && field.String() != "permit" && field.String() != "deny" {
		return fmt.Errorf("FRR policy action read-back is missing or unsupported")
	}
	if field := v.FieldByName("Version"); field.IsValid() && field.String() != "IPv4" && field.String() != "IPv6" {
		return fmt.Errorf("FRR prefix-list version read-back is missing or unsupported")
	}
	return nil
}
