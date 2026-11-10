// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestLiveInterfaceFullAllowlist(t *testing.T) {
	for _, overview := range []bool{false, true} {
		shape := diagnosticShape
		if overview {
			shape = overviewShape
		}
		assertLiveContractNames(t, shape, overview)
		raw := fullLiveRow(shape)
		raw["device"] = "test0"
		if !overview {
			raw["mtu"] = "0"
		}
		// Public underscores map explicitly to source dotted DHCP keys.
		if overview {
			raw["ifctl.nameserver"] = []string{"", "192.0.2.1 192.0.2.2\n"}
		}
		raw["config"] = map[string]string{"secret": "never-public"}
		encoded, err := json.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		object, err := shape.decode(context.Background(), encoded, !overview)
		if err != nil {
			t.Fatal(err)
		}
		attrs := object.Attributes()
		if len(attrs) != len(shape.attributes()) || strings.Contains(object.String(), "never-public") {
			t.Fatal("allowlist leaked or omitted fields")
		}
		for key, value := range attrs {
			if value.IsNull() || value.IsUnknown() {
				t.Fatalf("explicit public field %s lost value", key)
			}
		}
		if !overview && attrs["mtu"].(types.Int64).ValueInt64() != 0 {
			t.Fatal("explicit zero lost")
		}
		if !attrs["is_physical"].Equal(types.BoolValue(false)) {
			t.Fatal("explicit false lost")
		}
		if overview {
			var members []string
			diags := attrs["ifctl_nameserver"].(types.Set).ElementsAs(context.Background(), &members, false)
			if diags.HasError() || len(members) != 2 || !reflect.DeepEqual(members, []string{"", "192.0.2.1 192.0.2.2\n"}) {
				t.Fatalf("DHCP members changed: %#v", members)
			}
		}
	}
}

func fullLiveRow(s liveShape) map[string]any {
	row := map[string]any{}
	for _, k := range s.strings {
		row[liveWireKey(k)] = ""
	}
	for _, k := range s.bools {
		row[liveWireKey(k)] = false
	}
	for _, k := range s.ints {
		row[liveWireKey(k)] = 0
	}
	for _, k := range s.sets {
		if k == "nd6_flags" {
			row["nd6"] = map[string]any{"flags": []string{"", "unchanged value"}}
		} else {
			row[liveWireKey(k)] = []string{"", "unchanged value"}
		}
	}
	for k, child := range s.objects {
		row[liveWireKey(k)] = fullLiveRow(child)
	}
	for k, child := range s.lists {
		row[k] = []any{fullLiveRow(child)}
	}
	return row
}

func TestLiveInterfaceNullAndOrdering(t *testing.T) {
	for _, overview := range []bool{false, true} {
		body := `{"test1":{"device":"test1"},"test0":{"device":"test0","ipv4":[{"ipaddr":"192.0.2.1","subnetbits":24,"tunnel":true},{"ipaddr":"192.0.2.1","subnetbits":24,"tunnel":false}]}}`
		reverse := `{"test0":{"ipv4":[{"ipaddr":"192.0.2.1","tunnel":false,"subnetbits":24},{"subnetbits":24,"tunnel":true,"ipaddr":"192.0.2.1"}],"device":"test0"},"test1":{"device":"test1"}}`
		if overview {
			body = `{"total":2,"rowCount":2,"current":1,"rows":[{"device":"test1"},{"device":"test0","ipv4":[{"ipaddr":"192.0.2.1","vhid":"2"},{"ipaddr":"192.0.2.1","vhid":"1"}]}]}`
			reverse = `{"total":2,"rowCount":2,"current":1,"rows":[{"device":"test0","ipv4":[{"vhid":"1","ipaddr":"192.0.2.1"},{"vhid":"2","ipaddr":"192.0.2.1"}]},{"device":"test1"}]}`
		}
		a, err := decodeLiveInterfaces(context.Background(), []byte(body), overview)
		if err != nil {
			t.Fatal(err)
		}
		b, err := decodeLiveInterfaces(context.Background(), []byte(reverse), overview)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatal("reordering observations changed public values")
		}
		absent := a[1].Attributes()
		shape := diagnosticShape
		if overview {
			shape = overviewShape
		}
		for _, key := range shape.sets {
			if !absent[key].Equal(types.SetValueMust(types.StringType, nil)) {
				t.Fatalf("missing %s must be a known empty set", key)
			}
		}
		if !absent["media"].IsNull() || !absent["ipv4"].Equal(types.ListValueMust(types.ObjectType{AttrTypes: func() map[string]attr.Type {
			if overview {
				return overviewAddress.types()
			}
			return diagnosticAddress4.types()
		}()}, nil)) {
			t.Fatal("missing optional scalar/collection fabricated")
		}
		if overview && !absent["vlan"].IsNull() {
			t.Fatal("absent object fabricated")
		}
	}
}

func TestLiveInterfaceMalformed(t *testing.T) {
	diagnostics := []string{`null`, `[{}]`, `{"test0":null}`, `{"test0":{"device":"wrong"}}`, `{"test0":{"device":""}}`, `{"test0":{"device":1}}`, `{"test0":{"device":"test0","mtu":1500}}`, `{"test0":{"device":"test0","mtu":"bad"}}`, `{"test0":{"device":"test0","mtu":"9223372036854775808"}}`, `{"test0":{"device":"test0","is_physical":0}}`, `{"test0":{"device":"test0","media":false}}`, `{"test0":{"device":"test0","flags":[null]}}`, `{"test0":{"device":"test0","flags":[1]}}`, `{"test0":{"device":"test0","ipv4":[null]}}`, `{"test0":{"device":"test0","ipv4":[{"subnetbits":"24"}]}}`, `{"test0":{"device":"test0","ipv6":[{"tentative":1}]}}`}
	overview := []string{`null`, `{}`, `{"total":0,"rowCount":0,"current":1,"rows":null}`, `{"total":1,"rowCount":0,"current":1,"rows":[]}`, `{"total":0,"rowCount":0,"current":2,"rows":[]}`, `{"total":"0","rowCount":0,"current":1,"rows":[]}`, `{"total":null,"rowCount":0,"current":1,"rows":[]}`, `{"total":2,"rowCount":2,"current":1,"rows":[{"device":"test0"},{"device":"test0"}]}`, `{"total":1,"rowCount":1,"current":1,"rows":[{"identifier":"test0"}]}`, `{"total":1,"rowCount":1,"current":1,"rows":[{"device":"test0","vlan":[]}]}`, `{"total":1,"rowCount":1,"current":1,"rows":[{"device":"test0","mtu":1500}]}`}
	for mode, cases := range map[bool][]string{false: diagnostics, true: overview} {
		for i, raw := range cases {
			t.Run(fmt.Sprintf("%t/%d", mode, i), func(t *testing.T) {
				if _, err := decodeLiveInterfaces(context.Background(), []byte(raw), mode); err == nil {
					t.Fatal("accepted malformed response", raw)
				}
			})
		}
	}
	for _, raw := range []string{`{}`, `[]`, `[ ]`} {
		rows, err := decodeLiveInterfaces(context.Background(), []byte(raw), false)
		if err != nil || len(rows) != 0 {
			t.Fatalf("legitimate empty inventory: %s %v", raw, err)
		}
	}
	rows, err := decodeLiveInterfaces(context.Background(), []byte(`{"total":0,"rowCount":0,"current":1,"rows":[]}`), true)
	if err != nil || len(rows) != 0 {
		t.Fatal("empty overview", err)
	}
}

func TestLiveInterfaceHTTPAndBoundedRead(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet || r.ContentLength > 0 || r.URL.RawQuery != "" {
			t.Error("non-read request")
		}
		key, secret, ok := r.BasicAuth()
		if !ok || key != "fixture-key" || secret != "fixture-secret" {
			t.Error("auth not reused")
		}
		if r.URL.Path != diagnosticEndpoint && r.URL.Path != overviewEndpoint {
			t.Error("unexpected endpoint")
		}
		_, _ = fmt.Fprint(w, `{}`)
	}))
	defer server.Close()
	client, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: server.URL, APIKey: "fixture-key", APISecret: "fixture-secret", Insecure: true, MaxReadConcurrency: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err = client.LockMutex(context.Background()); err != nil {
		t.Fatal(err)
	}
	readCtx, readCancel := context.WithTimeout(context.Background(), time.Second)
	defer readCancel()
	if _, err = readLiveInterfaces(readCtx, client, false); err != nil {
		t.Fatal(err)
	}
	client.UnlockMutex()
	if err = client.AcquireRead(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err = readLiveInterfaces(ctx, client, false); err == nil {
		t.Fatal("read escaped bounded semaphore")
	}
	client.ReleaseRead()
	if calls.Load() != 1 {
		t.Fatal("cancelled waiter issued request")
	}
	if _, err = readLiveInterfaces(context.Background(), nil, false); err == nil {
		t.Fatal("nil client accepted")
	}
	for _, status := range []int{401, 403, 404, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var attempts atomic.Int64
			failure := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				attempts.Add(1)
				w.WriteHeader(status)
				_, _ = fmt.Fprint(w, "secret-bearing malformed body")
			}))
			defer failure.Close()
			c, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: failure.URL, APIKey: "fixture-key", APISecret: "fixture-secret", RetryMax: 1, RetryWaitMin: time.Millisecond, RetryWaitMax: time.Millisecond})
			if err != nil {
				t.Fatal(err)
			}
			_, err = readLiveInterfaces(context.Background(), c, true)
			expectedAttempts := int64(1)
			if status == 500 {
				expectedAttempts = 2
			}
			if attempts.Load() != expectedAttempts {
				t.Fatalf("configured retry budget lost: %d calls", attempts.Load())
			}
			if err == nil || strings.Contains(err.Error(), "secret-bearing") || strings.Contains(err.Error(), "plugin") {
				t.Fatalf("unsafe status diagnostic: %v", err)
			}
		})
	}
}

func TestLiveInterfaceNoRedirectCredentialForwarding(t *testing.T) {
	var secondaryHits atomic.Int64
	secondary := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { secondaryHits.Add(1) }))
	defer secondary.Close()
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, secondary.URL, http.StatusFound) }))
	defer primary.Close()
	client, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: primary.URL, APIKey: "fixture-key", APISecret: "fixture-secret"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = readLiveInterfaces(context.Background(), client, false)
	if err == nil || secondaryHits.Load() != 0 || strings.Contains(err.Error(), secondary.URL) {
		t.Fatalf("unsafe redirect handling: hits=%d error=%v", secondaryHits.Load(), err)
	}
}

func TestLiveInterfaceDuplicateDiagnosticKey(t *testing.T) {
	_, err := decodeLiveInterfaces(context.Background(), []byte(`{"test0":{"device":"test0"},"test0":{"device":"test0"}}`), false)
	if err == nil {
		t.Fatal("duplicate device key silently overwritten")
	}
}

func TestLiveInterfaceSourceWireNames(t *testing.T) {
	diagnostic, err := decodeLiveInterfaces(context.Background(), []byte(`{"test0":{"device":"test0","ipv6":[{"ipaddr":"2001:db8::1","link-local":true}]}}`), false)
	if err != nil {
		t.Fatal(err)
	}
	address := diagnostic[0].Attributes()["ipv6"].(types.List).Elements()[0].(types.Object)
	if !address.Attributes()["link_local"].Equal(types.BoolValue(true)) {
		t.Fatal("link-local not mapped")
	}
	overview, err := decodeLiveInterfaces(context.Background(), []byte(`{"total":1,"rowCount":1,"current":1,"rows":[{"device":"test0","laggproto":"lacp","lagghash":"l2","laggoptions":{"flags":["FAST"],"flowid_shift":"0"},"laggstatistics":{"active ports":"2","flapping":"0"},"nd6":{"flags":["PERFORMNUD"]},"ipv4":[{"ipaddr":"192.0.2.1/24","status":"MASTER","advbase":"1","advskew":"0"}],"ifctl.nameserver":[""],"ifctl.router":["192.0.2.1 192.0.2.2"],"ifctl.prefix":["2001:db8::/64"],"ifctl.searchdomain":["example.invalid"]}]}`), true)
	if err != nil {
		t.Fatal(err)
	}
	attrs := overview[0].Attributes()
	for key, want := range map[string]string{"lagg_proto": "lacp", "lagg_hash": "l2"} {
		if !attrs[key].Equal(types.StringValue(want)) {
			t.Fatal("wire name not mapped", key)
		}
	}
	for _, key := range []string{"lagg_options", "lagg_statistics", "nd6_flags", "ifctl_nameserver", "ifctl_router", "ifctl_prefix", "ifctl_searchdomain"} {
		if attrs[key].IsNull() {
			t.Fatal("wire object not mapped", key)
		}
	}
	statistics := attrs["lagg_statistics"].(types.Object).Attributes()
	if !statistics["active_ports"].Equal(types.StringValue("2")) {
		t.Fatal("active ports not mapped")
	}
	carp := attrs["ipv4"].(types.List).Elements()[0].(types.Object).Attributes()
	for key, want := range map[string]string{"carp_status": "MASTER", "adv_base": "1", "adv_skew": "0"} {
		if !carp[key].Equal(types.StringValue(want)) {
			t.Fatal("CARP wire name not mapped", key)
		}
	}
}

// Released API observations can repeat route strings. Exercise every public set,
// including nested sets, without copying appliance inventory into this fixture.
func TestLiveInterfaceDuplicateSetMembers(t *testing.T) {
	var verify func(liveShape, types.Object)
	verify = func(shape liveShape, object types.Object) {
		t.Helper()
		attrs := object.Attributes()
		for _, key := range shape.sets {
			set := attrs[key].(types.Set)
			want := types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(""), types.StringValue("192.0.2.0/24"),
				types.StringValue(" 192.0.2.0/24 "), types.StringValue("192.0.2.1 192.0.2.2\n"),
			})
			if !set.Equal(want) {
				t.Fatalf("%s members changed: %s", key, set.String())
			}
		}
		for key, child := range shape.objects {
			verify(child, attrs[key].(types.Object))
		}
	}
	var duplicate func(liveShape, map[string]any)
	duplicate = func(shape liveShape, row map[string]any) {
		for _, key := range shape.sets {
			members := []string{"", "192.0.2.0/24", "192.0.2.0/24", " 192.0.2.0/24 ", "", "192.0.2.1 192.0.2.2\n"}
			if key == "nd6_flags" {
				row["nd6"] = map[string]any{"flags": members}
			} else {
				row[liveWireKey(key)] = members
			}
		}
		for key, child := range shape.objects {
			duplicate(child, row[liveWireKey(key)].(map[string]any))
		}
	}
	for _, overview := range []bool{false, true} {
		shape := diagnosticShape
		if overview {
			shape = overviewShape
		}
		row := fullLiveRow(shape)
		row["device"] = "test0"
		if !overview {
			row["mtu"] = "1500"
		}
		duplicate(shape, row)
		var envelope any = map[string]any{"test0": row}
		if overview {
			envelope = map[string]any{"total": 1, "rowCount": 1, "current": 1, "rows": []any{row}}
		}
		raw, err := json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := decodeLiveInterfaces(context.Background(), raw, overview)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Fatal("lost interface")
		}
		verify(shape, rows[0])
	}
}

// Literal contract names are independent of liveShape and generated fixtures.
// Removing a public field must fail even if schema and decoder shrink together.
func assertLiveContractNames(t *testing.T, shape liveShape, overview bool) {
	t.Helper()
	expected := map[string]string{
		"diagnostic":               "device media media_raw macaddr status is_physical mtu flags capabilities options supported_media groups ipv4 ipv6",
		"diagnostic.ipv4":          "ipaddr subnetbits tunnel",
		"diagnostic.ipv6":          "ipaddr subnetbits tunnel autoconf deprecated link_local tentative",
		"overview":                 "identifier description device enabled link_type addr4 addr6 macaddr macaddr_hw mtu status media media_raw is_physical vlan_tag lagg_proto lagg_hash lagg_options lagg_statistics flags capabilities options supported_media groups gateways routes nd6_flags ifctl_nameserver ifctl_router ifctl_prefix ifctl_searchdomain ipv4 ipv6 vlan",
		"overview.ipv4":            "ipaddr vhid carp_status adv_base adv_skew peer peer6",
		"overview.ipv6":            "ipaddr vhid carp_status adv_base adv_skew peer peer6",
		"overview.vlan":            "tag proto pcp parent",
		"overview.lagg_options":    "flags flowid_shift",
		"overview.lagg_statistics": "active_ports flapping",
	}
	var verify func(string, liveShape)
	verify = func(path string, current liveShape) {
		t.Helper()
		want := strings.Fields(expected[path])
		got := make([]string, 0, len(current.attributes()))
		for key := range current.attributes() {
			got = append(got, key)
		}
		sort.Strings(want)
		sort.Strings(got)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s contract names changed: got %v want %v", path, got, want)
		}
		for key, child := range current.objects {
			verify(path+"."+key, child)
		}
		for key, child := range current.lists {
			verify(path+"."+key, child)
		}
	}
	root := "diagnostic"
	if overview {
		root = "overview"
	}
	verify(root, shape)
}

func TestLiveInterfaceNullSets(t *testing.T) {
	for _, shape := range []liveShape{diagnosticShape, overviewShape} {
		for _, explicitNull := range []bool{false, true} {
			row := map[string]any{"device": "test0"}
			for _, key := range shape.sets {
				if explicitNull {
					if key == "nd6_flags" {
						row["nd6"] = map[string]any{"flags": nil}
					} else {
						row[liveWireKey(key)] = nil
					}
				}
			}
			// A present LAGG object must also expose a known empty flags set.
			if len(shape.objects) != 0 {
				row["laggoptions"] = map[string]any{}
				if explicitNull {
					row["laggoptions"] = map[string]any{"flags": nil}
				}
			}
			raw, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			object, err := shape.decode(context.Background(), raw, false)
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range shape.sets {
				if !object.Attributes()[key].Equal(types.SetValueMust(types.StringType, nil)) {
					t.Fatalf("null=%t: %s is not known empty", explicitNull, key)
				}
			}
			if len(shape.objects) != 0 && !object.Attributes()["lagg_options"].(types.Object).Attributes()["flags"].Equal(types.SetValueMust(types.StringType, nil)) {
				t.Fatal("nested missing/null set is not known empty")
			}
		}
	}
}

func TestLiveInterfaceUpstreamInventoryAndAddressShapes(t *testing.T) {
	diagnostic, err := decodeLiveInterfaces(context.Background(), []byte(`{"pfsync0":{"device":"pfsync0"},"test0":{"device":"test0","ipv4":[{"ipaddr":"192.0.2.1","subnetbits":24}]}}`), false)
	if err != nil {
		t.Fatal(err)
	}
	overview, err := decodeLiveInterfaces(context.Background(), []byte(`{"total":1,"rowCount":1,"current":1,"rows":[{"device":"test0","ipv4":[{"ipaddr":"192.0.2.1/24"}]}]}`), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostic) != 2 || len(overview) != 1 || !diagnostic[0].Attributes()["device"].Equal(types.StringValue("pfsync0")) {
		t.Fatal("upstream-defined inventories must not be homogenized")
	}
	for _, test := range []struct {
		row  types.Object
		want string
	}{{diagnostic[1], "192.0.2.1"}, {overview[0], "192.0.2.1/24"}} {
		address := test.row.Attributes()["ipv4"].(types.List).Elements()[0].(types.Object)
		if !address.Attributes()["ipaddr"].Equal(types.StringValue(test.want)) {
			t.Fatal("diagnostic bare address / overview CIDR shape changed")
		}
	}
}
