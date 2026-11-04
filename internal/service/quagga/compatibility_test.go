// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestFRRInstalledPackageAndReadOnlyCapability(t *testing.T) {
	for _, tc := range []struct {
		name, packages, template string
		status                   int
		good                     bool
	}{
		{"frr2_revision", `[{"name":"os-frr","version":"2.0_1","installed":"1"}]`, `{"family":{"ipv4":{"value":"IPv4","selected":1},"ipv6":{"value":"IPv6","selected":0}}}`, 200, true},
		{"frr2_revision_2", `[{"name":"os-frr","version":"2.0_2","installed":"1"}]`, `{"family":{"ipv4":{"value":"IPv4","selected":1},"ipv6":{"value":"IPv6","selected":0}}}`, 200, true},
		{"frr2_revision_999", `[{"name":"os-frr","version":"2.0_999","installed":"1"}]`, `{"family":{"ipv4":{"value":"IPv4","selected":1},"ipv6":{"value":"IPv6","selected":0}}}`, 200, true},
		{"frr2_revision_model_contradiction", `[{"name":"os-frr","version":"2.0_2","installed":"1"}]`, `{"multiprotocol":"0"}`, 200, false},
		{"legacy", `[{"name":"os-frr","version":"1.55_2","installed":"1"}]`, `{"multiprotocol":"0"}`, 200, true},
		{"missing", "[]", `{}`, 200, false},
		{"catalog_not_installed", `[{"name":"os-frr","version":"2.0","installed":"0"}]`, `{}`, 200, false},
		{"future", `[{"name":"os-frr","version":"2.1","installed":"1"}]`, `{}`, 200, false},
		{"malformed_version", `[{"name":"os-frr","version":"2.0-devel","installed":"1"}]`, `{}`, 200, false},
		{"duplicate", `[{"name":"os-frr","version":"2.0","installed":"1"},{"name":"os-frr","version":"1.55","installed":"1"}]`, `{}`, 200, false},
		{"contradiction", `[{"name":"os-frr","version":"2.0","installed":"1"}]`, `{"multiprotocol":"0"}`, 200, false},
		{"malformed_options", `[{"name":"os-frr","version":"2.0","installed":"1"}]`, `{"family":{"ipv4":{"value":"IPv4","selected":2},"ipv6":{"value":"IPv6","selected":0}}}`, 200, false},
		{"permission", `[]`, `{}`, 403, false},
		{"authentication", `[]`, `{}`, 401, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != "GET" {
					t.Errorf("discovery mutated: %s", r.Method)
				}
				w.WriteHeader(tc.status)
				switch r.URL.Path {
				case "/api/core/firmware/info":
					_, _ = fmt.Fprintf(w, `{"product_version":"99.1","package":%s}`, tc.packages)
				case "/api/quagga/bgp/get_neighbor":
					_, _ = fmt.Fprintf(w, `{"neighbor":%s}`, tc.template)
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
				}
			}))
			defer api.Close()
			got, err := discoverFRR(context.Background(), newTestClient(t, api.URL))
			if (err == nil) != tc.good {
				t.Fatalf("got capability %v err %v", got, err)
			}
			if tc.good && (got.Core != "99.1" || requests != 2) {
				t.Fatal("core was used as plugin authority or template omitted")
			}
		})
	}
}

func TestFRRFamilyMigrationAndLegacyTransportBoundary(t *testing.T) {
	for _, tc := range []struct {
		address string
		multi   bool
		want    string
	}{{"192.0.2.1", false, "ipv4"}, {"2001:db8::1", false, "ipv6"}, {"192.0.2.1", true, "ipv4,ipv6"}, {"2001:db8::1", true, "ipv4,ipv6"}} {
		got, err := legacyFamilies(tc.address, types.BoolValue(tc.multi))
		if err != nil || setToCSV(context.Background(), got) != tc.want {
			t.Fatalf("%+v %v %v", tc, got, err)
		}
	}
	unknown, err := legacyFamilies("192.0.2.1", types.BoolUnknown())
	if err != nil || !unknown.IsUnknown() {
		t.Fatal("unknown alias became a default")
	}
	for _, values := range [][]string{{}, {"IPv4"}, {"ipv5"}} {
		if _, err := validateFamilies(context.Background(), sliceToSet(values)); err == nil {
			t.Fatalf("accepted invalid family %v", values)
		}
	}
	payload := bgpNeighborAPIRequest{Address: "192.0.2.1", Family: "ipv6", MultiProtocol: "0"}
	if _, err := prepareBGPRequest(context.Background(), nil, frrCapability{}, &payload); err == nil {
		t.Fatal("legacy cross-family write accepted")
	}
	got, err := prepareBGPRequest(context.Background(), nil, frrCapability{Grouped: true}, &payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := (*got)["multiprotocol"]; exists || (*got)["family"] != "ipv6" {
		t.Fatal("FRR2 did not preserve cross-family intent")
	}
}

func TestFRRGroupedMembershipIdentityAndSingleSelect(t *testing.T) {
	members := policyMembership{"a": {Group: "policy"}, "b": {Group: "policy"}, "c": {Group: "other"}}
	if !equivalentPolicies([]string{"b"}, []string{"a"}, members) {
		t.Fatal("nonrepresentative sibling was lost")
	}
	if equivalentPolicies([]string{"stale"}, []string{"a"}, members) || equivalentPolicies([]string{"c"}, []string{"a"}, members) {
		t.Fatal("unproven semantic identity accepted")
	}
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/quagga/bgp/search_routemap":
			_, _ = fmt.Fprint(w, `{"rows":[{"uuid":"a"},{"uuid":"b"},{"uuid":"c"}],"current":1,"rowCount":3,"total":3}`)
		case "/api/quagga/bgp/get_routemap/a", "/api/quagga/bgp/get_routemap/b":
			_, _ = fmt.Fprint(w, `{"routemap":{"name":"policy","id":"20"}}`)
		case "/api/quagga/bgp/get_routemap/c":
			_, _ = fmt.Fprint(w, `{"routemap":{"name":"other","id":"10"}}`)
		default:
			t.Errorf("unexpected membership request %s", r.URL.Path)
		}
	}))
	defer api.Close()
	client := newTestClient(t, api.URL)
	for _, tc := range []struct {
		value string
		good  bool
	}{{"b", true}, {"a,b", true}, {"a,c", false}, {"stale", false}} {
		payload := map[string]string{"linkedRoutemapOut": tc.value}
		got, err := prepareBGPRequest(context.Background(), client, frrCapability{Grouped: true}, &payload)
		if (err == nil) != tc.good {
			t.Fatalf("value %s got %v err %v", tc.value, got, err)
		}
	}
	prior := BGPNeighborResourceModel{LinkedRoutemapOut: sliceToSet([]string{"b"}), LinkedPrefixlistIn: types.SetNull(types.StringType), LinkedPrefixlistOut: types.SetNull(types.StringType), LinkedRoutemapIn: types.SetNull(types.StringType)}
	current := prior
	current.LinkedRoutemapOut = sliceToSet([]string{"a"})
	// Empty/null selectors from read-back are concrete empty typed sets.
	current.LinkedRoutemapIn = sliceToSet(nil)
	current.LinkedPrefixlistIn = sliceToSet(nil)
	current.LinkedPrefixlistOut = sliceToSet(nil)
	if err := normalizeBGPRead(context.Background(), client, frrCapability{Grouped: true}, &prior, &current); err != nil {
		t.Fatal(err)
	}
	redistribution := BGPRedistributionResourceModel{RouteMap: types.StringValue("b")}
	normalized := redistribution
	normalized.RouteMap = types.StringValue("a")
	if err := normalizeBGPRead(context.Background(), client, frrCapability{Grouped: true}, &redistribution, &normalized); err != nil {
		t.Fatal(err)
	}
	if normalized.RouteMap.ValueString() != "b" {
		t.Fatal("scalar redistribution route-map sibling identity was lost")
	}
	if !current.LinkedRoutemapOut.Equal(prior.LinkedRoutemapOut) {
		t.Fatal("saved sibling identity changed")
	}
}

func TestFRRStrictOwnerSelectorDecoding(t *testing.T) {
	for _, raw := range []string{`null`, `["id"]`, `{"id":{"selected":2}}`, `{"id":{}}`, `12`, `true`} {
		var v frrSelectedMapList
		if json.Unmarshal([]byte(raw), &v) == nil {
			t.Fatalf("accepted malformed selector %s", raw)
		}
	}
	var single frrSelectedMap
	if json.Unmarshal([]byte(`{"a":{"selected":1},"b":{"selected":1}}`), &single) == nil {
		t.Fatal("accepted multiple single-select values")
	}
	var values frrSelectedMapList
	if err := json.Unmarshal([]byte(`{"b":{"selected":"1"},"a":{"selected":1}}`), &values); err != nil || len(values) != 2 || values[0] != "a" {
		t.Fatal(values, err)
	}
}

func TestFRRNeighborPasswordPreservedOutsidePublicModel(t *testing.T) {
	const secret = "unreadable-existing-password"
	var api bgpNeighborAPIResponse
	if err := json.Unmarshal([]byte(`{"enabled":"1","address":"192.0.2.1","remoteas":"64512","password":"`+secret+`","family":"ipv4","nexthopself":"0","peergroup":"","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}`), &api); err != nil {
		t.Fatal(err)
	}
	var model BGPNeighborResourceModel
	model.fromAPI(context.Background(), &api, "existing-id")
	encoded, err := marshalAdvancedRequest(t, model)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if json.Unmarshal(encoded, &payload) != nil {
		t.Fatal("payload malformed")
	}
	if _, exists := payload["password"]; exists {
		t.Fatal("provider would overwrite an unowned existing credential")
	}
	stateJSON, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(stateJSON), secret) || strings.Contains(string(encoded), secret) {
		t.Fatal("unowned password escaped into model or payload")
	}
}

func TestFRRSourceRevisionAndMalformedCapabilityBoundaries(t *testing.T) {
	for _, v := range []string{"2.1", "2.1_1", "2.0.1", "2.0_2-devel", "2.0_2_1", "2.0_-1", "2.0_2,1", "2.00", "1.56", "1.55-devel", "2.0_", ""} {
		if frrPackageVersion.MatchString(v) {
			t.Fatalf("unsupported plugin version accepted: %s", v)
		}
	}
	if _, err := legacyFamilies("bad-address", types.BoolValue(false)); err == nil {
		t.Fatal("invalid single-stack transport silently defaulted")
	}
	for _, raw := range []string{`{"enabled":"maybe"}`, `{"enabled":"1","family":[]}`, `{"enabled":"1","family":"IPv6"}`} {
		var decoded bgpNeighborAPIResponse
		if err := json.Unmarshal([]byte(raw), &decoded); err == nil && validateBGPResponse(&decoded, frrCapability{Grouped: true}) == nil {
			t.Fatal("malformed API family/flag accepted")
		}
	}
}

func TestFRRMixedFamilyPrefixGroupsRejectOnlySelectedGroup(t *testing.T) {
	members := policyMembership{"a": {Group: "mixed", Family: "IPv4"}, "b": {Group: "mixed", Family: "IPv6"}, "c": {Group: "coherent", Family: "IPv4"}}
	if _, err := policyGroups([]string{"a"}, members); err == nil {
		t.Fatal("selected mixed-family logical group was silently collapsed")
	}
	if _, err := policyGroups([]string{"c"}, members); err != nil {
		t.Fatal("unrelated mixed-family group blocked a coherent relation", err)
	}
}

func TestFRRDiagnosticsNeverEchoUntrustedCredentialValues(t *testing.T) {
	const secret = "private-response-credential-value"
	for _, err := range []error{fmt.Errorf("ambiguous mutation response: got %q", secret), fmt.Errorf("reconfigure: unexpected status %q", secret), &opnsense.ValidationError{Fields: map[string]string{"neighbor.password": secret}}, &opnsense.MutationReconfigureError{Cause: fmt.Errorf("status %q", secret)}} {
		if strings.Contains(frrDiagnostic(err), secret) {
			t.Fatal("untrusted response credential escaped into diagnostic")
		}
	}
}

// A missing selector is not an authoritative empty selection: refresh must
// retain prior state and fail, rather than silently deleting a linked policy.
func TestFRRReadPresenceAndRouteMapSiblingFields(t *testing.T) {
	var redistribution bgpRedistributionAPIResponse
	if err := json.Unmarshal([]byte(`{"enabled":"0","description":"fixture","redistribute":"connected"}`), &redistribution); err != nil {
		t.Fatal(err)
	}
	if validateBGPResponse(&redistribution, frrCapability{Grouped: true}) == nil {
		t.Fatal("missing route-map selector was inferred as an empty clear")
	}
	var route routeMapAPIResponse
	if err := json.Unmarshal([]byte(`{"enabled":"1","description":"fixture","name":"policy","action":"permit","id":"20","match":"as-id","match2":"prefix-id","match3":"community-id","set":"local-preference 123"}`), &route); err != nil {
		t.Fatal(err)
	}
	if err := validateBGPResponse(&route, frrCapability{Grouped: true}); err != nil {
		t.Fatal(err)
	}
	var model RouteMapResourceModel
	model.fromAPI(context.Background(), &route, "entry-uuid")
	request := model.toAPI(context.Background())
	if model.ID.ValueString() != "entry-uuid" || request.Order != "20" || request.MatchASPath != "as-id" || request.MatchPrefix != "prefix-id" || request.MatchCommunity != "community-id" || request.Set != "local-preference 123" {
		t.Fatal("route-map sequence or sibling policy fields disappeared during roundtrip")
	}
}

type frrReadTransport func(*http.Request) (*http.Response, error)

func (f frrReadTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFRRReadOnlyDiscoveryBoundsBackendAndPreservesEarlierDeadline(t *testing.T) {
	for _, shorter := range []bool{false, true} {
		client := newTestClient(t, "http://fixture.invalid")
		client.HTTPClient().Transport = frrReadTransport(func(r *http.Request) (*http.Response, error) {
			deadline, bounded := r.Context().Deadline()
			if !bounded || time.Until(deadline) > 30*time.Second {
				t.Error("firmware backend request lacks bounded discovery deadline")
			}
			if shorter && time.Until(deadline) > time.Second {
				t.Error("earlier caller deadline was extended")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("{}"))}, nil
		})
		ctx := context.Background()
		cancel := func() {}
		if shorter {
			ctx, cancel = context.WithTimeout(ctx, time.Second)
		}
		var result map[string]any
		err := frrRead(ctx, client, "/api/core/firmware/info", &result)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestFRROwnedNullScalarsAndRecoverySelectorIdentity(t *testing.T) {
	for _, raw := range []string{`{"description":null}`, `{"keepalive":null}`, `{"holddown":null}`} {
		var response bgpNeighborAPIResponse
		if json.Unmarshal([]byte(raw), &response) == nil {
			t.Fatalf("owned null accepted: %s", raw)
		}
	}
	var valid bgpNeighborAPIResponse
	if err := json.Unmarshal([]byte(`{"description":"","keepalive":"","linkedRoutemapOut":{"":{"selected":1}}}`), &valid); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		key, got, want string
		ok             bool
	}{
		{"description", `"a,b"`, "b,a", false},
		{"family", `"ipv6,ipv4"`, "ipv4,ipv6", true},
		{"match2", `"b,a"`, "a,b", true},
		{"linkedRoutemapOut", `"b,a"`, "a,b", true},
		{"linkedRoutemap", `"b,a"`, "a,b", false},
		{"description", `null`, "", false},
	} {
		if exactBGPPayload(map[string]json.RawMessage{tc.key: json.RawMessage(tc.got)}, map[string]string{tc.key: tc.want}) != tc.ok {
			t.Fatalf("incorrect recovery identity for %s", tc.key)
		}
	}
}
