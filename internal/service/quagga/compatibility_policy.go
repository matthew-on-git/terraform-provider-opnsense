// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

type (
	policyMember     struct{ Group, Family string }
	policyMembership map[string]policyMember
)

func loadPolicyMembership(ctx context.Context, c *opnsense.Client, kind string) (policyMembership, error) {
	opts := opnsense.ReqOpts{SearchEndpoint: "/api/quagga/bgp/search_" + kind, GetEndpoint: "/api/quagga/bgp/get_" + kind, Monad: kind, StrictMissingObjectResponse: true}
	rows, err := opnsense.Search[struct {
		UUID string `json:"uuid"`
	}](ctx, c, opts, opnsense.SearchParams{})
	if err != nil {
		return nil, err
	}
	result := policyMembership{}
	for _, row := range rows {
		if row.UUID == "" {
			return nil, fmt.Errorf("FRR policy membership has an empty UUID")
		}
		item, err := opnsense.Get[map[string]json.RawMessage](ctx, c, opts, row.UUID)
		if err != nil {
			return nil, fmt.Errorf("FRR policy membership changed or cannot be read: %w", err)
		}
		key := "name"
		if kind == "aspath" || kind == "communitylist" {
			key = "number"
		}
		var group string
		if json.Unmarshal((*item)[key], &group) != nil || group == "" {
			return nil, fmt.Errorf("FRR %s membership lacks logical identity", kind)
		}
		member := policyMember{Group: group}
		if kind == "prefixlist" {
			var family opnsense.SelectedMap
			if json.Unmarshal((*item)["version"], &family) != nil || (family != "IPv4" && family != "IPv6") {
				return nil, fmt.Errorf("FRR prefix group has malformed address family")
			}
			member.Family = string(family)
		}
		if _, exists := result[row.UUID]; exists {
			return nil, fmt.Errorf("FRR policy membership contains a duplicate UUID")
		}
		result[row.UUID] = member
	}
	return result, nil
}

func policyGroups(values []string, members policyMembership) (map[string]bool, error) {
	groups := map[string]bool{}
	for _, id := range values {
		member, ok := members[id]
		if !ok {
			return nil, fmt.Errorf("FRR linked policy UUID is stale or absent from authoritative membership; refresh/import dependencies before retrying")
		}
		for _, sibling := range members {
			if sibling.Group == member.Group && sibling.Family != member.Family {
				return nil, fmt.Errorf("FRR selected prefix group mixes address families; split it before managing references")
			}
		}
		groups[member.Group] = true
	}
	return groups, nil
}

// The comparison uses the complete authoritative membership, never UUID order,
// string resemblance, or the grouped dropdown's incomplete sibling list.
func equivalentPolicies(prior, remote []string, members policyMembership) bool {
	a, err := policyGroups(prior, members)
	if err != nil {
		return false
	}
	b, err := policyGroups(remote, members)
	if err != nil || len(a) != len(b) {
		return false
	}
	for key := range a {
		if !b[key] {
			return false
		}
	}
	return true
}

func policyKind(field string) string {
	switch field {
	case "linkedPrefixlistIn", "linkedPrefixlistOut", "match2":
		return "prefixlist"
	case "linkedRoutemapIn", "linkedRoutemapOut", "linkedRoutemap":
		return "routemap"
	case "match":
		return "aspath"
	case "match3":
		return "communitylist"
	}
	return ""
}

func prepareBGPRequest[K any](ctx context.Context, c *opnsense.Client, capability frrCapability, payload *K) (*map[string]string, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var fields map[string]string
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return nil, err
	}
	if err := validateAdvancedWire(fields); err != nil {
		return nil, err
	}
	if _, neighbor := any(payload).(*bgpNeighborAPIRequest); neighbor {
		if selected := fields["linklocalinterface"]; selected != "" && !capability.InterfaceOptions[selected] {
			return nil, fmt.Errorf("FRR link_local_interface is not offered by the discovered model; no mutation was sent")
		}
		for wire := range fields {
			if advancedNeighborFields[wire] && !advancedSupported(capability, wire) {
				return nil, fmt.Errorf("FRR configured %s is unsupported by the discovered plugin/model; no mutation was sent", wire)
			}
		}
	}
	if _, peer := any(payload).(*bgpPeerGroupAPIRequest); peer {
		if _, exists := fields["listenranges"]; exists && !advancedSupported(capability, "listenranges") {
			if c == nil {
				return nil, fmt.Errorf("FRR listen_ranges requires authoritative peer-group field proof; no mutation was sent")
			}
			var template struct {
				PeerGroup map[string]json.RawMessage `json:"peergroup"`
			}
			if err := frrRead(ctx, c, "/api/quagga/bgp/get_peergroup", &template); err != nil {
				return nil, err
			}
			raw, exists := template.PeerGroup["listenranges"]
			var ranges frrSelectedMapList
			if !exists || json.Unmarshal(raw, &ranges) != nil {
				return nil, fmt.Errorf("FRR listen_ranges is unsupported by the discovered peer-group model; no mutation was sent")
			}
		}
	}
	if multi, exists := fields["multiprotocol"]; exists {
		if fields["family"] == "" {
			families, err := legacyFamilies(fields["address"], boolFromWire(multi))
			if err != nil {
				return nil, err
			}
			fields["family"] = setToCSV(ctx, families)
		}
		values, err := validateFamilies(ctx, sliceToSet(strings.Split(fields["family"], ",")))
		if err != nil {
			return nil, err
		}
		if capability.Grouped {
			delete(fields, "multiprotocol")
		} else {
			dual := len(values) == 2
			expected, err := legacyFamilies(fields["address"], boolFromWire("0"))
			if err != nil {
				return nil, err
			}
			if !dual && setToCSV(ctx, expected) != fields["family"] {
				return nil, fmt.Errorf("legacy os-frr cannot represent a cross-transport single address family; no mutation was sent")
			}
			fields["multiprotocol"] = "0"
			if dual {
				fields["multiprotocol"] = "1"
			}
			delete(fields, "family")
		}
	}
	if capability.Grouped {
		cache := map[string]policyMembership{}
		for field, value := range fields {
			kind := policyKind(field)
			if kind == "" || value == "" {
				continue
			}
			members, ok := cache[kind]
			if !ok {
				members, err = loadPolicyMembership(ctx, c, kind)
				if err != nil {
					return nil, err
				}
				cache[kind] = members
			}
			ids := strings.Split(value, ",")
			groups, err := policyGroups(ids, members)
			if err != nil {
				return nil, err
			}
			if field != "match" && field != "match2" && field != "match3" && len(groups) > 1 {
				return nil, fmt.Errorf("FRR 2.0 %s selects one logical policy group; multiple distinct groups cannot be represented", field)
			}
			// Preserve individual UUIDs for the multi-select route-map matches. For a
			// single logical group any configured existing sibling is equivalent.
			if field != "match" && field != "match2" && field != "match3" {
				sort.Strings(ids)
				fields[field] = ids[0]
			}
		}
	}
	return &fields, nil
}
