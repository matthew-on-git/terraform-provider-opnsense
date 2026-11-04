// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAdvancedNeighborFullWireAndAuthenticationBoundary(t *testing.T) {
	ctx := context.Background()
	model := BGPNeighborResourceModel{Address: types.StringValue("192.0.2.1"), Family: sliceToSet([]string{"ipv4"}), RemoteAS: types.Int64Null(), RemoteASMode: types.StringValue("external"), LocalAS: types.Int64Value(64512), Password: types.StringValue("configured-test-input"), Weight: types.Int64Value(0), LocalIP: types.StringValue("192.0.2.2"), LinkLocalInterface: types.StringValue("wan"), NextHopSelf: types.BoolValue(true), NextHopSelfAll: types.BoolValue(true), Multihop: types.BoolValue(false), RouteReflectorClient: types.BoolValue(true), SoftReconfigurationInbound: types.BoolValue(true), BFD: types.BoolValue(true), BFDStrict: types.BoolValue(true), ConnectTimer: types.Int64Value(1), DefaultOriginate: types.BoolValue(false), ASOverride: types.BoolValue(false), RemovePrivateAS: types.StringValue("remove-private-AS all replace-AS"), AllowASIn: types.StringValue("origin"), DisableConnectedCheck: types.BoolValue(true), AttributeUnchanged: types.StringValue("as-path"), Capabilities: sliceToSet([]string{"dynamic", "fqdn"})}
	encoded, err := marshalAdvancedRequest(t, model)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]string
	if err = json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	for field := range advancedNeighborFields {
		if _, ok := wire[field]; !ok {
			t.Fatalf("missing wire field %s", field)
		}
	}
	if len(advancedNeighborFields) != 20 || wire["weight"] != "0" || wire["remoteas"] != "" || wire["multihop"] != "0" {
		t.Fatal("lost zero/false/null semantics")
	}
	if err := validateAdvancedWire(wire); err != nil {
		t.Fatal(err)
	}
	model.Password = types.StringNull()
	encoded, _ = marshalAdvancedRequest(t, model)
	if strings.Contains(string(encoded), `"password"`) {
		t.Fatal("null password was sent")
	}
	model.Password = types.StringValue("")
	encoded, _ = marshalAdvancedRequest(t, model)
	if !strings.Contains(string(encoded), `"password":""`) {
		t.Fatal("explicit password clear was lost")
	}
	var remote bgpNeighborAPIResponse
	if err := json.Unmarshal([]byte(`{"password":"remote-not-owned","weight":"0","remoteas":"","localas":"","connecttimer":""}`), &remote); err != nil {
		t.Fatal(err)
	}
	model.Password = types.StringValue("prior-configured-input")
	model.fromAPI(ctx, &remote, "stable")
	if model.Password.ValueString() != "prior-configured-input" || !model.LocalAS.IsNull() || !model.ConnectTimer.IsNull() {
		t.Fatal("read synthesized secret or numeric zero")
	}
	var rs resource.SchemaResponse
	(&bgpNeighborResource{}).Schema(ctx, resource.SchemaRequest{}, &rs)
	var ds datasource.SchemaResponse
	(&bgpNeighborDataSource{}).Schema(ctx, datasource.SchemaRequest{}, &ds)
	if _, exists := ds.Schema.Attributes["password"]; exists {
		t.Fatal("lookup exposes password")
	}
	if _, exists := reflect.TypeFor[bgpNeighborAPIResponse]().FieldByName("Password"); exists {
		t.Fatal("unexpected public response field surface")
	}
}

func TestAdvancedNeighborInvalidIntentAndCapabilities(t *testing.T) {
	for name, fields := range map[string]map[string]string{
		"enum-label": {"removeprivateas": "removeallas"}, "allowas-label": {"allowas_in": "val1"}, "attribute-set": {"attributeunchanged": "as-path,med"}, "capability": {"capabilities": "unknown"}, "localas-zero": {"localas": "0"}, "weight-negative": {"weight": "-1"}, "weight-overflow": {"weight": "65536"}, "connect-zero": {"connecttimer": "0"}, "connect-overflow": {"connecttimer": "65001"}, "password-coupling": {"password": "private-input"}, "local-ip": {"localip": "192.0.2.1/24"}, "bfd-dependency": {"bfd_strict": "1", "bfd": "0"}, "nexthop-dependency": {"nexthopselfall": "1", "nexthopself": "0"}, "cidr-host-bits": {"listenranges": "192.0.2.1/24"}, "cidr-host": {"listenranges": "2001:db8::1"}, "name": {"name": "bad name"}, "sequence-uuid": {"id": "00000000-0000-4000-8000-000000000000"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateAdvancedWire(fields); err == nil {
				t.Fatal("invalid configured intent accepted")
			}
		})
	}
	for wire, options := range advancedEnums {
		for _, option := range options {
			if err := validateAdvancedWire(map[string]string{wire: option}); err != nil {
				t.Fatal(wire, option, err)
			}
		}
	}
	if err := validateAdvancedWire(map[string]string{"listenranges": "192.0.2.0/24,2001:db8::/64"}); err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{"1.46", "1.54_1", "2.0_4"} {
		if advancedSupported(frrCapability{Version: version}, "weight") {
			t.Fatal("inferred old package schema")
		}
		if !advancedSupported(frrCapability{Version: version, Fields: map[string]bool{"weight": true}}, "weight") {
			t.Fatal("ignored authoritative presence proof")
		}
	}
	payload := bgpNeighborAPIRequest{Weight: advancedInt64(context.Background(), types.Int64Value(0))}
	if _, err := prepareBGPRequest(context.Background(), nil, frrCapability{Version: "1.46"}, &payload); err == nil {
		t.Fatal("unsupported configured field passed prewrite gate")
	}
	for key := range advancedNeighborFields {
		if key == "password" {
			continue
		}
		if exactBGPPayload(map[string]json.RawMessage{}, map[string]string{key: ""}) {
			t.Fatal("missing advanced field matched", key)
		}
	}
	if exactBGPPayload(map[string]json.RawMessage{"password": json.RawMessage(`"same"`)}, map[string]string{"password": "same"}) {
		t.Fatal("secret used as ambiguous-create identity oracle")
	}
}

// Exercise the production request builder, including the supported-capability
// and serialization boundary, before asserting the wire map.
func marshalAdvancedRequest(t *testing.T, model BGPNeighborResourceModel) ([]byte, error) {
	t.Helper()
	payload, err := prepareBGPRequest(context.Background(), nil, frrCapability{Version: "2.0_1", Grouped: true, InterfaceOptions: map[string]bool{"wan": true}}, model.toAPI(context.Background()))
	if err != nil {
		return nil, err
	}
	return json.Marshal(payload)
}

func TestAdvancedTemplatePresenceMustBeAuthoritative(t *testing.T) {
	for _, tc := range []struct {
		wire, raw string
		valid     bool
	}{
		{"bfd", `null`, false}, {"bfd", `"2"`, false}, {"bfd", `"0"`, true}, {"password", `null`, false}, {"password", `""`, true}, {"weight", `null`, false}, {"capabilities", `null`, false}, {"capabilities", `[]`, true},
	} {
		if validAdvancedTemplateField(tc.wire, json.RawMessage(tc.raw)) != tc.valid {
			t.Fatal("malformed template accepted as authoritative capability", tc.wire)
		}
	}
	payload := bgpPeerGroupAPIRequest{ListenRanges: advancedSet(context.Background(), sliceToSet([]string{"192.0.2.0/24"}))}
	if _, err := prepareBGPRequest(context.Background(), nil, frrCapability{Version: "2.0_4", Grouped: true}, &payload); err == nil {
		t.Fatal("unpinned revision inferred peer listen-range capability")
	}
}

func TestOldPeerGroupRequiredFieldsRemainStrict(t *testing.T) {
	for _, missing := range []string{"remote_as_mode", "defaultoriginate"} {
		fields := map[string]any{"enabled": "1", "name": "fixture", "remote_as_mode": "", "remoteas": "64512", "family": "ipv4", "updatesource": "", "nexthopself": "0", "defaultoriginate": "0", "linkedPrefixlistIn": []string{}, "linkedPrefixlistOut": []string{}, "linkedRoutemapIn": []string{}, "linkedRoutemapOut": []string{}}
		delete(fields, missing)
		raw, _ := json.Marshal(fields)
		var response bgpPeerGroupAPIResponse
		if err := json.Unmarshal(raw, &response); err != nil {
			t.Fatal(err)
		}
		err := validateBGPResponse(&response, frrCapability{Version: "1.46"})
		if err == nil || !strings.Contains(err.Error(), missing) {
			t.Fatal("old peer-group field absence was exempted", missing, err)
		}
	}
}

func TestAdvancedNumericAndEnumProofDomains(t *testing.T) {
	for _, tc := range []struct{ wire, raw string }{{"weight", `"invalid"`}, {"weight", `"65536"`}, {"localas", `"0"`}, {"connecttimer", `"65001"`}, {"remote_as_mode", `"unsupported"`}, {"removeprivateas", `{"removeallas":{"selected":1}}`}, {"capabilities", `"dynamic,unsupported"`}} {
		if validAdvancedTemplateField(tc.wire, json.RawMessage(tc.raw)) {
			t.Fatal("invalid proof domain accepted", tc.wire)
		}
	}
}

func TestAdvancedInterfaceProofRejectsMetadata(t *testing.T) {
	for _, version := range []string{"2.0_1", "2.0_4"} {
		for _, metadata := range []string{`{"wan":{"value":"WAN"}}`, `{"wan":{"value":"WAN","selected":null}}`, `{"wan":{"value":"WAN","selected":2}}`, `{"wan":{"selected":0}}`, `{"wan":{"value":"WAN","selected":1},"lan":{"value":"LAN","selected":1}}`, `null`} {
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/core/firmware/info" {
					_ = json.NewEncoder(w).Encode(map[string]any{"package": []map[string]string{{"name": "os-frr", "version": version, "installed": "1"}}})
					return
				}
				_, _ = w.Write([]byte(`{"neighbor":{"family":{"ipv4":{"value":"IPv4","selected":1},"ipv6":{"value":"IPv6","selected":0}},"linklocalinterface":` + metadata + `}}`))
			}))
			_, err := discoverFRR(context.Background(), newTestClient(t, api.URL))
			api.Close()
			if err == nil {
				t.Fatal("invalid interface metadata admitted", version)
			}
		}
	}
}
