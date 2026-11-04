// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// OPNsense uses the track6 fields for both modern identity association and
// legacy track-interface configuration. These are assignment settings, not
// ownership of the upstream DHCPv6 client's prefix-delegation request.
var assignmentDelegationAPIFields = []string{
	"track6-interface", "track6-prefix-id", "track6_ifid",
	"track6_prefix_range", "track6_assoc_pd", "dhcpd6track6allowoverride",
}

func assignmentDelegationAttributes() map[string]schema.Attribute {
	stringAttribute := func(description, pattern, validation string) schema.StringAttribute {
		return schema.StringAttribute{
			Optional: true, Computed: true, MarkdownDescription: description,
			Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(pattern), validation)},
		}
	}
	return map[string]schema.Attribute{
		"ipv6_track_interface":         stringAttribute("Upstream assignment for `idassoc6` or legacy `track6`, such as `wan`. Omission preserves the existing parent; these modes require a nonempty parent distinct from this assignment.", `^(wan|lan|opt[0-9]+)?$`, "must be empty, wan, lan, or optN"),
		"ipv6_prefix_id":               stringAttribute("Delegated subnet prefix ID, in the appliance's hexadecimal encoding (for example `0x10`). Empty retains the appliance's default representation; omission preserves the saved value.", `^(0[xX][a-fA-F0-9]{0,30})?$`, "must be empty or a 0x-prefixed hexadecimal value of up to 30 digits"),
		"ipv6_interface_id":            stringAttribute("IPv6 interface identifier in hexadecimal (for example `0x1`), or empty for the appliance default.", `^(0[xX][a-fA-F0-9]{0,30})?$`, "must be empty or a 0x-prefixed hexadecimal value of up to 30 digits"),
		"ipv6_prefix_range":            stringAttribute("Identity-association prefix range, as a positive decimal string or empty for the appliance default. Strings preserve the distinction between an unset value and a number.", `^([1-9][0-9]*)?$`, "must be empty or a positive decimal integer"),
		"ipv6_identity_association_id": stringAttribute("Prefix-delegation identity association ID, as a nonnegative decimal string or empty for the appliance default. Empty and explicit zero are distinct saved values.", `^([0-9]+)?$`, "must be empty or a nonnegative decimal integer"),
		"ipv6_allow_manual_dhcpv6": schema.BoolAttribute{
			Optional: true, Computed: true,
			MarkdownDescription: "Allow manual DHCPv6/router-advertisement configuration for legacy track-interface mode. Preserves the appliance's dhcpd6track6allowoverride field; does not manage DHCP or RA services.",
		},
	}
}
