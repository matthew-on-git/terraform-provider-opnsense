// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"strings"
	"testing"
)

// This is a template-shaped fixture, not appliance evidence. The released
// 5ac2959c8514b4f60e94d65439be5f629f10521f bgpd.conf strict-BFD branch
// intentionally renders two plain bfd commands and one bfd strict command.
const frrAdvancedGeneratedFixture = `router bgp 64512
 neighbor 2001:db8::34 remote-as external
 neighbor 2001:db8::34 local-as 64513
 neighbor 2001:db8::34 password oracle-fixture-input
 neighbor 2001:db8::34 weight 0
 neighbor 2001:db8::34 interface vtnet0
 neighbor 2001:db8::34 bfd
 neighbor 2001:db8::34 bfd
 neighbor 2001:db8::34 bfd strict
 neighbor 2001:db8::34 ebgp-multihop 255
 neighbor 2001:db8::34 timers connect 30
 neighbor 2001:db8::34 attribute-unchanged as-path
 neighbor 2001:db8::34 disable-connected-check
 neighbor 2001:db8::34 capability extended-nexthop
 neighbor 2001:db8::34 capability dynamic
 neighbor 2001:db8::34 capability fqdn
 neighbor 2001:db8::34 capability link-local
 neighbor 2001:db8::35 remote-as internal
 bgp listen range 192.0.2.0/24 peer-group fixture_unused
 bgp listen range 2001:db8:31::/64 peer-group fixture_unused
 address-family ipv6 unicast
  neighbor 2001:db8::34 activate
  neighbor 2001:db8::34 next-hop-self all
  neighbor 2001:db8::34 soft-reconfiguration inbound
  neighbor 2001:db8::34 default-originate
  neighbor 2001:db8::34 as-override
  neighbor 2001:db8::34 remove-private-AS all replace-AS
  neighbor 2001:db8::34 allowas-in origin
  neighbor 2001:db8::35 activate
  neighbor 2001:db8::35 route-reflector-client
 exit-address-family
`

const frrAdvancedAuthenticationFixture = `[policy_fixture-owner_in]
src=2001:db8::34
dst=2001:db8:2::15
protocol=tcp
spi=0x1000
aalgo=tcp-md5
key=oracle-fixture-input

[policy_fixture-owner_out]
src=2001:db8:2::15
dst=2001:db8::34
protocol=tcp
spi=0x1000
aalgo=tcp-md5
key=oracle-fixture-input
`

func frrTestGeneratedProof(config, authentication string) error {
	return frrAdvancedGeneratedProof(config, authentication, "fixture-owner", "fixture_unused", "oracle-fixture-input", "2001:db8:2::15")
}

func TestFRRAdvancedGeneratedOracle(t *testing.T) {
	if err := frrTestGeneratedProof(frrAdvancedGeneratedFixture, frrAdvancedAuthenticationFixture); err != nil {
		t.Fatal("released-template-shaped fixture rejected", err)
	}
	normalized := strings.ReplaceAll(frrAdvancedGeneratedFixture, " neighbor ", "\t neighbor  ")
	if err := frrTestGeneratedProof(normalized, frrAdvancedAuthenticationFixture); err != nil {
		t.Fatal("insignificant command whitespace rejected", err)
	}
	// Each tracked command independently rejects absence, additional occurrences,
	// suffixes and incorrect placement. Error output must never reveal the input.
	seen := map[string]bool{}
	for _, raw := range strings.Split(frrAdvancedGeneratedFixture, "\n") {
		line := strings.TrimSpace(raw)
		if (!strings.HasPrefix(line, "neighbor ") && !strings.HasPrefix(line, "bgp listen range ")) || seen[line] {
			continue
		}
		seen[line] = true
		for _, change := range []struct{ name, replacement string }{
			{"missing", ""},
			{"duplicate", raw + "\n" + raw},
			{"suffix", raw + " unexpected-suffix"},
			{"wrong-family", " address-family ipv4 unicast\n" + raw + "\n exit-address-family"},
		} {
			// Do not use a command or its credential as a subtest name.
			t.Run(change.name, func(t *testing.T) {
				config := strings.Replace(frrAdvancedGeneratedFixture, raw+"\n", change.replacement+"\n", 1)
				err := frrTestGeneratedProof(config, frrAdvancedAuthenticationFixture)
				if err == nil {
					t.Fatal("counterfactual generated command accepted")
				}
				if strings.Contains(err.Error(), "oracle-fixture-input") || strings.Contains(err.Error(), "2001:db8") {
					t.Fatal("oracle diagnostic exposed private input")
				}
			})
		}
	}
	for _, command := range []string{"neighbor 2001:db8::34 update-source vtnet0", "neighbor 2001:db8::34 route-reflector-client"} {
		if frrTestGeneratedProof(command+"\n"+frrAdvancedGeneratedFixture, frrAdvancedAuthenticationFixture) == nil {
			t.Fatal("unexpected IPv6 transport/reflector command accepted")
		}
	}
}

func TestFRRAdvancedAuthenticationOracle(t *testing.T) {
	for _, tc := range []struct{ name, authentication string }{
		{"absent", ""},
		{"wrong-secret", strings.ReplaceAll(frrAdvancedAuthenticationFixture, "oracle-fixture-input", "wrong-input")},
		{"missing-secret", strings.ReplaceAll(frrAdvancedAuthenticationFixture, "key=oracle-fixture-input\n", "")},
		{"missing-local-address", strings.Replace(frrAdvancedAuthenticationFixture, "dst=2001:db8:2::15\n", "", 1)},
		{"missing-peer-address", strings.Replace(frrAdvancedAuthenticationFixture, "src=2001:db8::34\n", "", 1)},
		{"wrong-local-address", strings.ReplaceAll(frrAdvancedAuthenticationFixture, "2001:db8:2::15", "2001:db8:2::16")},
		{"wrong-peer-address", strings.ReplaceAll(frrAdvancedAuthenticationFixture, "2001:db8::34", "2001:db8::35")},
		{"wrong-owner", strings.ReplaceAll(frrAdvancedAuthenticationFixture, "fixture-owner", "another-owner")},
		{"missing-outbound", strings.Split(frrAdvancedAuthenticationFixture, "[policy_fixture-owner_out]")[0]},
		{"duplicate-section", frrAdvancedAuthenticationFixture + frrAdvancedAuthenticationFixture},
		{"duplicate-key", strings.Replace(frrAdvancedAuthenticationFixture, "key=oracle-fixture-input", "key=oracle-fixture-input\nkey=oracle-fixture-input", 1)},
		{"wrong-algorithm", strings.ReplaceAll(frrAdvancedAuthenticationFixture, "tcp-md5", "other")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := frrTestGeneratedProof(frrAdvancedGeneratedFixture, tc.authentication)
			if err == nil {
				t.Fatal("counterfactual authentication accepted")
			}
			if strings.Contains(err.Error(), "oracle-fixture-input") || strings.Contains(err.Error(), "2001:db8") {
				t.Fatal("oracle diagnostic exposed private input")
			}
		})
	}
}
