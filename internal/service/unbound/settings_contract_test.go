// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestGeneralSettingsPayloadOwnsOnlyFourBranches(t *testing.T) {
	t.Parallel()
	model := GeneralResourceModel{
		Enabled:           types.BoolValue(true),
		Port:              types.StringValue("5353"),
		HideIdentity:      types.BoolValue(true),
		PrivateDomains:    types.SetValueMust(types.StringType, []attr.Value{types.StringValue("corp.example")}),
		ACLDefaultAction:  types.StringValue("refuse"),
		ForwardingEnabled: types.BoolValue(true),
	}
	payload := generalSettingsPayload(context.Background(), model)
	if len(payload) != 4 {
		t.Fatalf("payload branches = %v, want exactly four owned branches", payload)
	}
	for _, name := range []string{"general", "advanced", "acls", "forwarding"} {
		if _, ok := payload[name]; !ok {
			t.Fatalf("payload missing %q branch", name)
		}
	}
	for _, forbidden := range []string{"dnsbl", "dots", "hosts", "aliases"} {
		if _, ok := payload[forbidden]; ok {
			t.Fatalf("payload must not include sibling branch %q", forbidden)
		}
	}
	advanced := payload["advanced"].(map[string]any)
	if advanced["privatedomain"] != "corp.example" {
		t.Fatalf("private domains = %#v", advanced["privatedomain"])
	}
}

func TestGeneralChangedPayloadEmitsOnlyChangedBranch(t *testing.T) {
	t.Parallel()
	base := GeneralResourceModel{
		Enabled: types.BoolValue(true), Port: types.StringValue("53"),
		HideIdentity: types.BoolValue(false), ACLDefaultAction: types.StringValue("allow"),
		ForwardingEnabled: types.BoolValue(false),
	}
	tests := map[string]func(*GeneralResourceModel){
		"general":    func(model *GeneralResourceModel) { model.Port = types.StringValue("5353") },
		"advanced":   func(model *GeneralResourceModel) { model.HideIdentity = types.BoolValue(true) },
		"acls":       func(model *GeneralResourceModel) { model.ACLDefaultAction = types.StringValue("refuse") },
		"forwarding": func(model *GeneralResourceModel) { model.ForwardingEnabled = types.BoolValue(true) },
	}
	for branch, mutate := range tests {
		branch, mutate := branch, mutate
		t.Run(branch, func(t *testing.T) {
			t.Parallel()
			plan := base
			mutate(&plan)
			payload := generalChangedPayload(context.Background(), plan, base)
			if len(payload) != 1 || payload[branch] == nil {
				t.Fatalf("payload = %#v, want only %q", payload, branch)
			}
		})
	}
}

func TestGeneralStrictNumericAndCacheDecoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		field     string
		value     any
		wantError bool
	}{
		{"numeric negative allowed", "serveexpiredttl", "-1", false},
		{"numeric lower boundary", "serveexpiredttl", "-99999999999999", false},
		{"numeric upper boundary", "serveexpiredttl", "99999999999999", false},
		{"numeric overflow", "serveexpiredttl", "100000000000000", true},
		{"numeric malformed", "serveexpiredttl", "12ms", true},
		{"cache valid uppercase", "msgcachesize", "4M", false},
		{"cache invalid suffix", "msgcachesize", "4mb", true},
		{"cache invalid sign", "msgcachesize", "-4m", true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			envelope := completeUnboundSettingsEnvelope(t, "53", "0", "0", nil)
			envelope.Advanced = replaceJSONFields(t, envelope.Advanced, map[string]any{test.field: test.value})
			_, err := flattenGeneralSettings(envelope)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, wantError=%t", err, test.wantError)
			}
		})
	}
}

func TestGeneralNewOptionalComputedAttributesHaveNoDefaults(t *testing.T) {
	t.Parallel()
	var response resource.SchemaResponse
	newGeneralResource().Schema(context.Background(), resource.SchemaRequest{}, &response)
	for _, name := range []string{
		"hide_identity", "hide_version", "enable_prefetch_key", "dnssec_stripped", "aggressive_nsec",
		"qname_min_strict", "enable_prefetch", "keep_probing_hosts", "extended_statistics", "log_queries",
		"log_replies", "tag_query_reply", "log_local_actions", "log_servfail", "serve_expired_enabled",
		"serve_expired_reset_ttl", "forwarding_enabled",
	} {
		attribute, ok := response.Schema.Attributes[name].(resourceschema.BoolAttribute)
		if !ok || !attribute.Optional || !attribute.Computed || attribute.Default != nil {
			t.Fatalf("%s must be Optional+Computed without a default, got %#v", name, response.Schema.Attributes[name])
		}
	}
	acl, ok := response.Schema.Attributes["acl_default_action"].(resourceschema.StringAttribute)
	if !ok || !acl.Optional || !acl.Computed || acl.Default != nil {
		t.Fatalf("acl_default_action must be Optional+Computed without a default, got %#v", response.Schema.Attributes["acl_default_action"])
	}
}

func TestGeneralPayloadOmittedVersusExplicitValues(t *testing.T) {
	t.Parallel()
	omitted := generalSettingsPayload(context.Background(), GeneralResourceModel{})
	if _, ok := omitted["advanced"].(map[string]any)["hideidentity"]; ok {
		t.Fatal("omitted hide_identity must not be encoded")
	}
	if _, ok := omitted["acls"].(map[string]any)["default_action"]; ok {
		t.Fatal("omitted acl_default_action must not be encoded")
	}
	explicit := generalSettingsPayload(context.Background(), GeneralResourceModel{
		HideIdentity:     types.BoolValue(false),
		PrivateDomains:   types.SetValueMust(types.StringType, nil),
		ACLDefaultAction: types.StringValue(""),
	})
	if got := explicit["advanced"].(map[string]any)["hideidentity"]; got != "0" {
		t.Fatalf("explicit false encoded as %#v", got)
	}
	if got := explicit["advanced"].(map[string]any)["privatedomain"]; got != "" {
		t.Fatalf("explicit empty set encoded as %#v", got)
	}
	if got := explicit["acls"].(map[string]any)["default_action"]; got != "" {
		t.Fatalf("explicit empty string encoded as %#v", got)
	}
}

func TestGeneralMutationUsesGeneralRouteOnceAndPreservesSiblings(t *testing.T) {
	t.Parallel()
	counts := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		counts[request.URL.Path]++
		switch request.URL.Path {
		case generalReqOpts.UpdateEndpoint:
			body, _ := io.ReadAll(request.Body)
			for _, forbidden := range []string{"dnsbl", "dots", "hosts", "aliases"} {
				if strings.Contains(string(body), `"`+forbidden+`"`) {
					t.Errorf("mutation body owns sibling %s: %s", forbidden, body)
				}
			}
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case generalReqOpts.ReconfigureEndpoint:
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		default:
			t.Errorf("unexpected route %s", request.URL.Path)
			http.NotFound(w, request)
		}
	}))
	t.Cleanup(server.Close)
	payload := generalSettingsPayload(context.Background(), GeneralResourceModel{HideIdentity: types.BoolValue(true)})
	if err := opnsense.UpdateSingleton(context.Background(), testClient(t, server.URL), generalReqOpts, &payload); err != nil {
		t.Fatal(err)
	}
	if counts[generalReqOpts.UpdateEndpoint] != 1 || counts[generalReqOpts.ReconfigureEndpoint] != 1 {
		t.Fatalf("route counts = %#v", counts)
	}
	if counts["/api/unbound/service/reconfigure"] != 0 {
		t.Fatalf("normal reconfigure route called: %#v", counts)
	}
}

func TestFlattenGeneralSettingsRejectsMissingBranch(t *testing.T) {
	t.Parallel()
	_, err := flattenGeneralSettings(&unboundSettingsEnvelope{
		General:    json.RawMessage(`{"port":"53"}`),
		Advanced:   json.RawMessage(`{}`),
		ACLs:       json.RawMessage(`{"default_action":{"allow":{"value":"allow","selected":1}}}`),
		Forwarding: json.RawMessage(`{"enabled":"0"}`),
	})
	if err == nil {
		t.Fatal("expected empty advanced branch to be diagnosed")
	}
}

func TestUnboundVersionCompatibilityIsStrict(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		actual string
		want   bool
		err    bool
	}{
		{"26.7.7", true, false},
		{"26.7.8", true, false},
		{"27.1", true, false},
		{"26.7.6", false, false},
		{"26.7.7_1", true, false},
		{"26.7.7-devel", false, true},
		{"not-a-version", false, true},
	} {
		got, err := unboundVersionAtLeast(test.actual, unboundParityMinimumVersion)
		if got != test.want || (err != nil) != test.err {
			t.Errorf("version %q: got=%t err=%v, want=%t err=%t", test.actual, got, err, test.want, test.err)
		}
	}
}

func TestGeneralCompatibilityFallbackIsVersionScoped(t *testing.T) {
	for _, test := range []struct {
		name       string
		version    string
		wantParity bool
		wantError  bool
	}{
		{"older appliance accepts legacy general shape", "26.7.6", false, false},
		{"current appliance rejects missing parity branches", "26.7.7", true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			envelope := completeUnboundSettingsEnvelope(t, "53", "0", "0", nil)
			body, err := json.Marshal(map[string]any{"unbound": map[string]any{"general": envelope.General}})
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case generalReqOpts.GetEndpoint:
					_, _ = w.Write(body)
				case "/api/core/firmware/status":
					_, _ = fmt.Fprintf(w, `{"product":{"product_name":"OPNsense","product_id":"opnsense","product_version":%q}}`, test.version)
				default:
					t.Fatalf("unexpected route %s", request.URL.Path)
				}
			}))
			defer server.Close()
			result, parity, err := getGeneralSettings(context.Background(), testClient(t, server.URL))
			if (err != nil) != test.wantError || parity != test.wantParity {
				t.Fatalf("result=%#v parity=%t err=%v", result, parity, err)
			}
			if !test.wantError && (result == nil || result.Port != "53") {
				t.Fatalf("legacy result=%#v", result)
			}
		})
	}
}

func TestGeneralOlderApplianceCreateAndUpdateUseLegacyContract(t *testing.T) {
	remotePort := "53"
	setCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/core/firmware/status":
			_, _ = w.Write([]byte(`{"product":{"product_name":"OPNsense","product_id":"opnsense","product_version":"26.7.6"}}`))
		case legacyGeneralReqOpts.UpdateEndpoint:
			setCalls++
			var envelope struct {
				Unbound map[string]map[string]string `json:"unbound"`
			}
			if err := json.NewDecoder(request.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if len(envelope.Unbound) != 1 || envelope.Unbound["general"] == nil {
				t.Fatalf("legacy request branches = %#v", envelope.Unbound)
			}
			remotePort = envelope.Unbound["general"]["port"]
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case legacyGeneralReqOpts.ReconfigureEndpoint:
			_, _ = w.Write([]byte(`{}`))
		case legacyGeneralReqOpts.GetEndpoint:
			envelope := completeUnboundSettingsEnvelope(t, remotePort, "0", "0", nil)
			_, _ = fmt.Fprintf(w, `{"unbound":{"general":%s}}`, envelope.General)
		default:
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	owner := &generalResource{client: testClient(t, server.URL)}
	createModel := GeneralResourceModel{Port: types.StringValue("53"), ActiveInterface: types.SetNull(types.StringType), OutgoingInterface: types.SetNull(types.StringType), PrivateDomains: types.SetNull(types.StringType), PrivateAddresses: types.SetNull(types.StringType), InsecureDomains: types.SetNull(types.StringType)}
	createPlan := unboundResourceState(t, owner, &createModel)
	create := resource.CreateResponse{State: tfsdk.State{Schema: createPlan.Schema}}
	initializeUnboundPrivate(t, &create)
	owner.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config(createPlan), Plan: tfsdk.Plan(createPlan)}, &create)
	if create.Diagnostics.HasError() {
		t.Fatal(create.Diagnostics)
	}
	updateModel := createModel
	updateModel.Port = types.StringValue("5353")
	updatePlan := unboundResourceState(t, owner, &updateModel)
	update := resource.UpdateResponse{State: tfsdk.State{Schema: updatePlan.Schema}, Private: create.Private}
	owner.Update(context.Background(), resource.UpdateRequest{Config: tfsdk.Config(updatePlan), State: create.State, Plan: tfsdk.Plan(updatePlan), Private: create.Private}, &update)
	if update.Diagnostics.HasError() || setCalls != 2 || remotePort != "5353" {
		t.Fatalf("diagnostics=%v setCalls=%d remotePort=%q", update.Diagnostics, setCalls, remotePort)
	}
}

func TestQueryForwardOlderApplianceCreateAndUpdateUseLegacyContract(t *testing.T) {
	remoteServer, remoteDescription := "192.0.2.53", "legacy-create"
	addCalls, updateCalls := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/core/firmware/status":
			_, _ = w.Write([]byte(`{"product":{"product_name":"OPNsense","product_id":"opnsense","product_version":"26.7.6"}}`))
		case legacyDomainOverrideReqOpts().SearchEndpoint:
			_, _ = w.Write([]byte(`{"rows":[],"rowCount":0,"total":0,"current":1}`))
		case legacyDomainOverrideReqOpts().AddEndpoint:
			addCalls++
			var envelope struct {
				Dot struct {
					Enabled     string `json:"enabled"`
					Domain      string `json:"domain"`
					Type        string `json:"type"`
					Server      string `json:"server"`
					Port        string `json:"port"`
					Verify      string `json:"verify"`
					Description string `json:"description"`
				} `json:"dot"`
			}
			if err := json.NewDecoder(request.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Dot.Enabled != "1" || envelope.Dot.Domain != "example.test" || envelope.Dot.Type != "forward" || envelope.Dot.Server != "192.0.2.53" || envelope.Dot.Port != "53" || envelope.Dot.Verify != "" || envelope.Dot.Description != "legacy-create" {
				t.Fatalf("legacy create request = %#v", envelope.Dot)
			}
			_, _ = w.Write([]byte(`{"result":"saved","uuid":"legacy-forward"}`))
		case legacyDomainOverrideReqOpts().UpdateEndpoint + "/legacy-forward":
			updateCalls++
			var envelope struct {
				Dot struct {
					Enabled     string `json:"enabled"`
					Domain      string `json:"domain"`
					Type        string `json:"type"`
					Server      string `json:"server"`
					Port        string `json:"port"`
					Verify      string `json:"verify"`
					Description string `json:"description"`
				} `json:"dot"`
			}
			if err := json.NewDecoder(request.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Dot.Enabled != "1" || envelope.Dot.Domain != "example.test" || envelope.Dot.Type != "forward" || envelope.Dot.Port != "53" || envelope.Dot.Verify != "" {
				t.Fatalf("legacy update request = %#v", envelope.Dot)
			}
			remoteServer, remoteDescription = envelope.Dot.Server, envelope.Dot.Description
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case legacyDomainOverrideReqOpts().ReconfigureEndpoint:
			_, _ = w.Write([]byte(`{}`))
		case legacyDomainOverrideReqOpts().GetEndpoint + "/legacy-forward":
			_, _ = fmt.Fprintf(w, `{"dot":{"enabled":"1","domain":"example.test","server":%q,"description":%q}}`, remoteServer, remoteDescription)
		default:
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	owner := &domainOverrideResource{client: testClient(t, server.URL)}
	createModel := DomainOverrideResourceModel{Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue(remoteServer), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue(remoteDescription)}
	createPlan := unboundResourceState(t, owner, &createModel)
	create := resource.CreateResponse{State: tfsdk.State{Schema: createPlan.Schema}}
	initializeUnboundPrivate(t, &create)
	owner.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config(createPlan), Plan: tfsdk.Plan(createPlan)}, &create)
	if create.Diagnostics.HasError() {
		t.Fatal(create.Diagnostics)
	}
	updateModel := createModel
	updateModel.Server = types.StringValue("192.0.2.54")
	updateModel.Description = types.StringValue("legacy-update")
	updatePlan := unboundResourceState(t, owner, &updateModel)
	update := resource.UpdateResponse{State: tfsdk.State{Schema: updatePlan.Schema}, Private: create.Private}
	owner.Update(context.Background(), resource.UpdateRequest{Config: tfsdk.Config(updatePlan), State: create.State, Plan: tfsdk.Plan(updatePlan), Private: create.Private}, &update)
	if update.Diagnostics.HasError() || addCalls != 1 || updateCalls != 1 || remoteServer != "192.0.2.54" {
		t.Fatalf("diagnostics=%v addCalls=%d updateCalls=%d server=%q", update.Diagnostics, addCalls, updateCalls, remoteServer)
	}
}

func TestOlderApplianceRejectsConfiguredParityAndDoTSurfaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/core/firmware/status" {
			t.Fatalf("unexpected mutation route %s", request.URL.Path)
		}
		_, _ = w.Write([]byte(`{"product":{"product_name":"OPNsense","product_id":"opnsense","product_version":"26.7.6"}}`))
	}))
	t.Cleanup(server.Close)

	generalOwner := &generalResource{client: testClient(t, server.URL)}
	generalModel := GeneralResourceModel{Port: types.StringValue("53"), HideIdentity: types.BoolValue(true), ActiveInterface: types.SetNull(types.StringType), OutgoingInterface: types.SetNull(types.StringType), PrivateDomains: types.SetNull(types.StringType), PrivateAddresses: types.SetNull(types.StringType), InsecureDomains: types.SetNull(types.StringType)}
	generalPlan := unboundResourceState(t, generalOwner, &generalModel)
	generalResponse := resource.CreateResponse{State: tfsdk.State{Schema: generalPlan.Schema}}
	generalOwner.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config(generalPlan), Plan: tfsdk.Plan(generalPlan)}, &generalResponse)
	if !generalResponse.Diagnostics.HasError() || !strings.Contains(generalResponse.Diagnostics.Errors()[0].Summary(), "Unsupported OPNsense version") {
		t.Fatalf("expected parity rejection, got %v", generalResponse.Diagnostics)
	}

	forwardOwner := &domainOverrideResource{client: testClient(t, server.URL)}
	forwardModel := DomainOverrideResourceModel{Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("dot"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(853), VerifyCN: types.StringValue("dns.example.test"), Description: types.StringValue("dot")}
	forwardPlan := unboundResourceState(t, forwardOwner, &forwardModel)
	forwardResponse := resource.CreateResponse{State: tfsdk.State{Schema: forwardPlan.Schema}}
	forwardOwner.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config(forwardPlan), Plan: tfsdk.Plan(forwardPlan)}, &forwardResponse)
	if !forwardResponse.Diagnostics.HasError() || !strings.Contains(forwardResponse.Diagnostics.Errors()[0].Summary(), "Unsupported OPNsense version") {
		t.Fatalf("expected DoT rejection, got %v", forwardResponse.Diagnostics)
	}
}

func TestUnboundCompatibilityRejectsUnverifiedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/core/firmware/status" {
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
		_, _ = w.Write([]byte(`{"product":{"product_name":"Other","product_id":"other","product_version":"99.1"}}`))
	}))
	defer server.Close()
	if _, _, err := unboundParitySupport(context.Background(), testClient(t, server.URL)); err == nil {
		t.Fatal("non-OPNsense identity was accepted")
	}
}

func TestFlattenGeneralSettingsDecodesIndexedLogLevels(t *testing.T) {
	t.Parallel()
	envelope := completeUnboundSettingsEnvelope(t, "53", "0", "0", nil)
	envelope.Advanced = replaceJSONFields(t, envelope.Advanced, map[string]any{
		"logverbosity": []map[string]any{{"value": "Level 0", "selected": 0}, {"value": "Level 1", "selected": 0}, {"value": "Level 2", "selected": 1}},
		"valloglevel":  []map[string]any{{"value": "Level 0", "selected": 0}, {"value": "Level 1", "selected": 1}},
	})
	envelope.ACLs = replaceJSONFields(t, envelope.ACLs, map[string]any{"default_action": map[string]any{"refuse": map[string]any{"value": "Refuse", "selected": 1}}})
	result, err := flattenGeneralSettings(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if result.VerbosityLevel != "2" || result.ValidationLevel != "1" {
		t.Fatalf("log levels = verbosity %q, validation %q", result.VerbosityLevel, result.ValidationLevel)
	}
}

func TestNormalizeIndexedChoiceRejectsAmbiguousSelection(t *testing.T) {
	t.Parallel()
	_, err := normalizeIndexedChoice(json.RawMessage(`[{"selected":1},{"selected":1}]`))
	if err == nil || !strings.Contains(err.Error(), "multiple selections") {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateSelectedMapRejectsMalformedSelections(t *testing.T) {
	t.Parallel()
	for name, raw := range map[string]string{
		"invalid selection": `{"one":{"selected":2}}`,
		"ambiguous single":  `{"one":{"selected":1},"two":{"selected":1}}`,
		"missing single":    `{"one":{"selected":0}}`,
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateSelectedMap(json.RawMessage(raw), true); err == nil {
				t.Fatalf("validateSelectedMap(%s) succeeded", raw)
			}
		})
	}
	if err := validateSelectedMap(json.RawMessage(`{"one":{"selected":1},"two":{"selected":0}}`), true); err != nil {
		t.Fatalf("valid single selection: %v", err)
	}
	if err := validateSelectedMap(json.RawMessage(`{"one":{"selected":1},"two":{"selected":1}}`), false); err != nil {
		t.Fatalf("valid multiple selection: %v", err)
	}
}

func TestDNSBLSettingsItemsValueUsesNestedObjectType(t *testing.T) {
	t.Parallel()
	value := dnsblSettingsItemsValue([]dnsblSettingsItemModel{{
		UUID: types.StringValue("uuid"), Enabled: types.BoolValue(true),
		Type:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("ag")}),
		Lists: types.SetValueMust(types.StringType, nil), Allowlists: types.SetValueMust(types.StringType, nil),
		Blocklists: types.SetValueMust(types.StringType, nil), Wildcards: types.SetValueMust(types.StringType, nil),
		SourceNets: types.SetValueMust(types.StringType, nil), Address: types.StringValue("192.0.2.31"),
		NXDomain: types.BoolValue(true), CacheTTL: types.Int64Value(300), Description: types.StringValue("fixture"),
	}})
	if value.IsNull() || value.IsUnknown() || len(value.Elements()) != 1 {
		t.Fatalf("unexpected nested value: %s", value.String())
	}
	if _, ok := value.ElementType(context.Background()).(types.ObjectType); !ok {
		t.Fatalf("element type = %T, want types.ObjectType", value.ElementType(context.Background()))
	}
}

func TestDNSBLDiscoveryZeroOneAndMultiple(t *testing.T) {
	t.Parallel()
	for name, response := range map[string]string{
		"zero":     `{"rows":[],"rowCount":0,"total":0,"current":1}`,
		"one":      `{"rows":[{"uuid":"owned"}],"rowCount":1,"total":1,"current":1}`,
		"multiple": `{"rows":[{"uuid":"a"},{"uuid":"b"}],"rowCount":2,"total":2,"current":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.URL.Path != dnsblReqOpts.SearchEndpoint {
					t.Fatalf("unexpected route %s", request.URL.Path)
				}
				_, _ = w.Write([]byte(response))
			}))
			defer server.Close()
			uuid, err := discoverDNSBLUUID(context.Background(), testClient(t, server.URL), nil)
			switch name {
			case "zero":
				if err != nil || uuid != "" {
					t.Fatalf("zero result = %q, %v", uuid, err)
				}
			case "one":
				if err != nil || uuid != "owned" {
					t.Fatalf("one result = %q, %v", uuid, err)
				}
			case "multiple":
				if err == nil || !strings.Contains(err.Error(), "prior state is incomplete") {
					t.Fatalf("multiple error = %v", err)
				}
			}
		})
	}
}

func TestDNSBLAddAndUpdateUseItemContractsAndNormalReconfigure(t *testing.T) {
	t.Parallel()
	counts := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		counts[request.URL.Path]++
		switch request.URL.Path {
		case dnsblReqOpts.AddEndpoint:
			body, _ := io.ReadAll(request.Body)
			if !strings.Contains(string(body), `"blocklist"`) {
				t.Errorf("add body lacks blocklist monad: %s", body)
			}
			_, _ = w.Write([]byte(`{"result":"saved","uuid":"owned"}`))
		case dnsblReqOpts.UpdateEndpoint + "/owned":
			body, _ := io.ReadAll(request.Body)
			for _, preserved := range []string{`"source_nets":"192.0.2.0/24"`, `"cache_ttl":"123"`, `"description":"keep"`} {
				if !strings.Contains(string(body), preserved) {
					t.Errorf("update failed to preserve %s: %s", preserved, body)
				}
			}
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case dnsblReqOpts.ReconfigureEndpoint:
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		default:
			t.Errorf("unexpected route %s", request.URL.Path)
			http.NotFound(w, request)
		}
	}))
	t.Cleanup(server.Close)
	client := testClient(t, server.URL)
	model := DNSBLResourceModel{
		Enabled: types.BoolValue(true), Type: types.SetValueMust(types.StringType, nil),
		Lists: types.SetValueMust(types.StringType, nil), Whitelists: types.SetValueMust(types.StringType, nil),
		Blocklists: types.SetValueMust(types.StringType, nil), Wildcards: types.SetValueMust(types.StringType, nil),
		Address: types.StringValue("0.0.0.0"), NXDomain: types.BoolValue(false), SafeSearch: types.BoolValue(false),
	}
	if uuid, err := opnsense.Add(context.Background(), client, withoutReconfigure(dnsblReqOpts), dnsblItemRequest(context.Background(), model, nil)); err != nil || uuid != "owned" {
		t.Fatalf("add result = %q, %v", uuid, err)
	}
	if err := opnsense.RetryReconfigure(context.Background(), client, dnsblReqOpts); err != nil {
		t.Fatal(err)
	}
	existing := &dnsblItemAPIResponse{SourceNets: opnsense.SelectedMapList{"192.0.2.0/24"}, CacheTTL: "123", Description: "keep"}
	if err := opnsense.Update(context.Background(), client, withoutReconfigure(dnsblReqOpts), dnsblItemRequest(context.Background(), model, existing), "owned"); err != nil {
		t.Fatal(err)
	}
	if err := opnsense.RetryReconfigure(context.Background(), client, dnsblReqOpts); err != nil {
		t.Fatal(err)
	}
	if counts[dnsblReqOpts.AddEndpoint] != 1 || counts[dnsblReqOpts.UpdateEndpoint+"/owned"] != 1 || counts[dnsblReqOpts.ReconfigureEndpoint] != 2 {
		t.Fatalf("route counts = %#v", counts)
	}
}

func TestDecodeDNSBLItemsFieldCompleteAndLossless(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"blocklist":{"b":{"enabled":"0","type":{"t":{"value":"t","selected":1}},"lists":{},"allowlists":{},"blocklists":{},"wildcards":{},"source_nets":{},"address":"::","nxdomain":"1","cache_ttl":"42","description":"second"},"a":{"enabled":"1","type":{},"lists":{},"allowlists":{},"blocklists":{},"wildcards":{},"source_nets":{"n":{"value":"192.0.2.0/24","selected":1}},"address":"0.0.0.0","nxdomain":"0","cache_ttl":"72000","description":"first"}}}`)
	items, err := decodeDNSBLItems(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].UUID.ValueString() != "a" || items[1].UUID.ValueString() != "b" {
		t.Fatalf("items not complete and deterministically ordered: %#v", items)
	}
	if items[0].CacheTTL.ValueInt64() != 72000 || items[0].Description.ValueString() != "first" || items[0].SourceNets.IsNull() {
		t.Fatalf("first item fields lost: %#v", items[0])
	}
}

func TestDomainOverrideTypeRoutingAndWireMapping(t *testing.T) {
	t.Parallel()
	if got := domainOverrideMutationOpts("query").AddEndpoint; got != "/api/unbound/settings/add_forward" {
		t.Fatalf("query add endpoint = %q", got)
	}
	if got := domainOverrideMutationOpts("dot").UpdateEndpoint; got != "/api/unbound/settings/set_dot" {
		t.Fatalf("DoT update endpoint = %q", got)
	}
	model := DomainOverrideResourceModel{
		Enabled:     types.BoolValue(true),
		Domain:      types.StringValue("example.com"),
		Type:        types.StringValue("dot"),
		Server:      types.StringValue("192.0.2.53"),
		ServerPort:  types.Int64Value(853),
		VerifyCN:    types.StringValue(""),
		Description: types.StringValue("test"),
	}
	wire := model.toAPI(context.Background())
	if wire.Type != "dot" || wire.Port != "853" || wire.Verify != "" {
		t.Fatalf("unexpected wire mapping: %#v", wire)
	}
}

func TestDomainOverrideFromAPIRejectsMalformedTypeAndPort(t *testing.T) {
	t.Parallel()
	for name, response := range map[string]domainOverrideAPIResponse{
		"type":   {Enabled: "1", Domain: "example.test", Type: "bogus", Server: "192.0.2.1", Port: "53"},
		"port":   {Enabled: "1", Domain: "example.test", Type: "forward", Server: "192.0.2.1", Port: "not-a-port"},
		"server": {Enabled: "1", Domain: "example.test", Type: "forward", Port: "53"},
	} {
		t.Run(name, func(t *testing.T) {
			var model DomainOverrideResourceModel
			if err := model.fromAPI(context.Background(), &response, "id"); err == nil {
				t.Fatal("expected malformed response error")
			}
		})
	}
}

func TestDomainOverrideCurrentResponseRequiresPort(t *testing.T) {
	t.Parallel()
	for name, payload := range map[string]string{
		"missing": `{"dot":{"enabled":"1","domain":"example.test","type":{"forward":{"value":"forward","selected":1}},"server":"192.0.2.1","verify":"","description":"test"}}`,
		"empty":   `{"dot":{"enabled":"1","domain":"example.test","type":{"forward":{"value":"forward","selected":1}},"server":"192.0.2.1","port":"","verify":"","description":"test"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			var envelope struct {
				Dot domainOverrideAPIResponse `json:"dot"`
			}
			if err := json.Unmarshal([]byte(payload), &envelope); err == nil {
				t.Fatal("expected strict current response to reject absent or empty port")
			}
		})
	}
}

func TestDomainOverridePortConditionalDefault(t *testing.T) {
	t.Parallel()
	owner := &domainOverrideResource{}
	tests := []struct {
		name        string
		forwardType types.String
		configPort  types.Int64
		want        int64
	}{
		{"omitted query", types.StringValue("query"), types.Int64Null(), 53},
		{"omitted dot", types.StringValue("dot"), types.Int64Null(), 853},
		{"omitted type keeps query compatibility", types.StringNull(), types.Int64Null(), 53},
		{"explicit dot 53", types.StringValue("dot"), types.Int64Value(53), 53},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := DomainOverrideResourceModel{Type: test.forwardType, ServerPort: test.configPort}
			configState := unboundResourceState(t, owner, &model)
			planValue := types.Int64Unknown()
			if !test.configPort.IsNull() {
				planValue = test.configPort
			}
			request := planmodifier.Int64Request{Config: tfsdk.Config(configState), ConfigValue: test.configPort, PlanValue: planValue}
			response := planmodifier.Int64Response{PlanValue: request.PlanValue}
			domainOverridePortDefaultPlanModifier{}.PlanModifyInt64(context.Background(), request, &response)
			if response.Diagnostics.HasError() || response.PlanValue.ValueInt64() != test.want {
				t.Fatalf("plan=%v diagnostics=%v, want %d", response.PlanValue, response.Diagnostics, test.want)
			}
		})
	}
	var schemaResponse resource.SchemaResponse
	owner.Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	attribute := schemaResponse.Schema.Attributes["server_port"].(resourceschema.Int64Attribute)
	if attribute.Default != nil || len(attribute.PlanModifiers) != 1 {
		t.Fatalf("server_port must use only its conditional plan default, got %#v", attribute)
	}
}

func TestDomainOverrideLegacyReadDefaultsQueryAndPort(t *testing.T) {
	for _, test := range []struct {
		name      string
		version   string
		wantError bool
		wantGets  int
	}{
		{"older appliance", "26.7.6", false, 2},
		{"current appliance", "26.7.7", true, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			gets := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case domainOverrideReqOpts.GetEndpoint + "/legacy":
					gets++
					_, _ = w.Write([]byte(`{"dot":{"enabled":"1","domain":"example.test","server":"192.0.2.53","description":"legacy"}}`))
				case "/api/core/firmware/status":
					_, _ = fmt.Fprintf(w, `{"product":{"product_name":"OPNsense","product_id":"opnsense","product_version":%q}}`, test.version)
				default:
					t.Fatalf("unexpected route %s", request.URL.Path)
				}
			}))
			defer server.Close()
			result, err := getDomainOverride(context.Background(), testClient(t, server.URL), "legacy")
			if (err != nil) != test.wantError || gets != test.wantGets {
				t.Fatalf("result=%#v err=%v gets=%d", result, err, gets)
			}
			if !test.wantError && (string(result.Type) != "forward" || result.Port != "53") {
				t.Fatalf("legacy defaults = type %q port %q", result.Type, result.Port)
			}
		})
	}
}

func TestDNSBLCreateRequiresImportForExistingItem(t *testing.T) {
	t.Parallel()
	calls := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		calls = append(calls, request.URL.Path)
		if request.URL.Path != dnsblReqOpts.SearchEndpoint {
			t.Errorf("create attempted to overwrite existing item via %s", request.URL.Path)
			http.NotFound(w, request)
			return
		}
		_, _ = w.Write([]byte(`{"rows":[{"uuid":"existing"}],"rowCount":1,"total":1,"current":1}`))
	}))
	t.Cleanup(server.Close)
	owner := &dnsblResource{client: testClient(t, server.URL)}
	model := DNSBLResourceModel{Enabled: types.BoolValue(false), SafeSearch: types.BoolValue(false), Type: types.SetValueMust(types.StringType, nil), Lists: types.SetValueMust(types.StringType, nil), Whitelists: types.SetValueMust(types.StringType, nil), Blocklists: types.SetValueMust(types.StringType, nil), Wildcards: types.SetValueMust(types.StringType, nil), Address: types.StringValue("0.0.0.0"), NXDomain: types.BoolValue(false)}
	plan := unboundResourceState(t, owner, &model)
	response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	owner.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(plan)}, &response)
	if !response.Diagnostics.HasError() || !strings.Contains(response.Diagnostics.Errors()[0].Summary(), "requires import") {
		t.Fatalf("expected import-required diagnostic, got %v", response.Diagnostics)
	}
	if len(calls) != 1 || calls[0] != dnsblReqOpts.SearchEndpoint {
		t.Fatalf("create calls = %#v", calls)
	}
}

func TestDNSBLOlderApplianceFallsBackToLegacySingleton(t *testing.T) {
	setCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case dnsblReqOpts.SearchEndpoint:
			http.NotFound(w, request)
		case "/api/core/firmware/status":
			_, _ = w.Write([]byte(`{"product":{"product_name":"OPNsense","product_id":"opnsense","product_version":"26.7.6"}}`))
		case legacyDNSBLReqOpts.UpdateEndpoint:
			setCalls++
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if _, ok := body["unbound"].(map[string]any)["dnsbl"]; !ok {
				t.Fatalf("legacy request envelope = %#v", body)
			}
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case legacyDNSBLReqOpts.ReconfigureEndpoint:
			_, _ = w.Write([]byte(`{}`))
		case legacyDNSBLReqOpts.GetEndpoint:
			_, _ = w.Write([]byte(`{"unbound":{"dnsbl":{"enabled":"0","safesearch":"0","type":[],"lists":[],"whitelists":[],"blocklists":[],"wildcards":[],"address":"0.0.0.0","nxdomain":"0"}}}`))
		default:
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
	}))
	defer server.Close()
	owner := &dnsblResource{client: testClient(t, server.URL)}
	model := dnsblTestModel("0.0.0.0", false)
	plan := unboundResourceState(t, owner, &model)
	response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	initializeUnboundPrivate(t, &response)
	owner.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(plan)}, &response)
	if response.Diagnostics.HasError() || setCalls != 1 {
		t.Fatalf("diagnostics=%v setCalls=%d", response.Diagnostics, setCalls)
	}
}

func TestDomainOverrideTypeChangePreReadsThenRoutesDoT(t *testing.T) {
	t.Parallel()
	sequence := make([]string, 0, 4)
	getCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		sequence = append(sequence, request.URL.Path)
		switch request.URL.Path {
		case domainOverrideReqOpts.GetEndpoint + "/forward-id":
			getCalls++
			typ, server, port, verify := "forward", "192.0.2.53", "53", ""
			if getCalls > 1 {
				typ, server, port, verify = "dot", "192.0.2.54", "853", "dns.example.test"
			}
			_, _ = fmt.Fprintf(w, `{"dot":{"enabled":"1","domain":"example.test","type":{"%s":{"value":"%s","selected":1}},"server":"%s","port":"%s","verify":"%s","description":"test"}}`, typ, typ, server, port, verify)
		case domainOverrideDotReqOpts.UpdateEndpoint + "/forward-id":
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case domainOverrideReqOpts.ReconfigureEndpoint:
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
			http.NotFound(w, request)
		}
	}))
	t.Cleanup(server.Close)
	owner := &domainOverrideResource{client: testClient(t, server.URL)}
	stateModel := DomainOverrideResourceModel{ID: types.StringValue("forward-id"), Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue("test")}
	planModel := stateModel
	planModel.Type = types.StringValue("dot")
	planModel.Server = types.StringValue("192.0.2.54")
	planModel.ServerPort = types.Int64Value(853)
	planModel.VerifyCN = types.StringValue("dns.example.test")
	state := unboundResourceState(t, owner, &stateModel)
	plan := unboundResourceState(t, owner, &planModel)
	response := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}}
	owner.Update(context.Background(), resource.UpdateRequest{State: state, Plan: tfsdk.Plan(plan)}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	want := []string{domainOverrideReqOpts.GetEndpoint + "/forward-id", domainOverrideDotReqOpts.UpdateEndpoint + "/forward-id", domainOverrideReqOpts.ReconfigureEndpoint, domainOverrideReqOpts.GetEndpoint + "/forward-id"}
	if !reflect.DeepEqual(sequence, want) {
		t.Fatalf("request sequence = %#v, want %#v", sequence, want)
	}
}

type instrumentedCreateLocker struct {
	mu               sync.Mutex
	callsMu          sync.Mutex
	calls            int
	secondAttempting chan struct{}
}

func newInstrumentedCreateLocker() *instrumentedCreateLocker {
	return &instrumentedCreateLocker{secondAttempting: make(chan struct{})}
}

func (l *instrumentedCreateLocker) Lock() {
	l.callsMu.Lock()
	l.calls++
	if l.calls == 2 {
		close(l.secondAttempting)
	}
	l.callsMu.Unlock()
	l.mu.Lock()
}

func (l *instrumentedCreateLocker) Unlock() { l.mu.Unlock() }

func TestDomainOverrideCreateSerializesPreflightMutationAndRecovery(t *testing.T) {
	var mu sync.Mutex
	sequence := make([]string, 0, 8)
	searchCalls, addCalls := 0, 0
	recoveryEntered := make(chan struct{})
	releaseRecovery := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		mu.Lock()
		sequence = append(sequence, request.URL.Path)
		switch request.URL.Path {
		case domainOverrideReqOpts.SearchEndpoint:
			searchCalls++
			call := searchCalls
			mu.Unlock()
			if call == 2 {
				close(recoveryEntered)
				<-releaseRecovery
			}
			_, _ = w.Write([]byte(`{"rows":[],"rowCount":0,"total":0,"current":1}`))
		case domainOverrideReqOpts.AddEndpoint:
			addCalls++
			call := addCalls
			mu.Unlock()
			if call == 1 {
				_, _ = w.Write([]byte(`{}`))
				return
			}
			_, _ = w.Write([]byte(`{"result":"saved","uuid":"second"}`))
		case domainOverrideReqOpts.ReconfigureEndpoint:
			mu.Unlock()
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case domainOverrideReqOpts.GetEndpoint + "/second":
			mu.Unlock()
			_, _ = w.Write([]byte(`{"dot":{"enabled":"1","domain":"example.test","type":{"forward":{"value":"forward","selected":1}},"server":"192.0.2.53","port":"53","verify":"","description":"serialized"}}`))
		default:
			mu.Unlock()
			t.Errorf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	model := DomainOverrideResourceModel{Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue("serialized")}
	createLock := newInstrumentedCreateLocker()
	owner := &domainOverrideResource{client: testClient(t, server.URL), createLock: createLock}
	plan := unboundResourceState(t, owner, &model)
	responses := []*resource.CreateResponse{{State: tfsdk.State{Schema: plan.Schema}}, {State: tfsdk.State{Schema: plan.Schema}}}
	for _, response := range responses {
		initializeUnboundPrivate(t, response)
	}
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		owner.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(plan)}, responses[0])
	}()
	select {
	case <-recoveryEntered:
	case <-time.After(time.Second):
		close(releaseRecovery)
		t.Fatal("first create did not reach ambiguous-outcome recovery")
	}
	workers.Add(1)
	go func() {
		defer workers.Done()
		owner.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(plan)}, responses[1])
	}()
	select {
	case <-createLock.secondAttempting:
	case <-time.After(time.Second):
		close(releaseRecovery)
		t.Fatal("second create did not attempt the owner lock")
	}
	mu.Lock()
	blockedSearchCalls, blockedAddCalls := searchCalls, addCalls
	mu.Unlock()
	if blockedSearchCalls != 2 || blockedAddCalls != 1 {
		close(releaseRecovery)
		t.Fatalf("second create entered while first recovery held lock: searches=%d adds=%d", blockedSearchCalls, blockedAddCalls)
	}
	close(releaseRecovery)
	workers.Wait()

	mu.Lock()
	gotSequence := append([]string(nil), sequence...)
	gotAddCalls := addCalls
	mu.Unlock()
	want := []string{domainOverrideReqOpts.SearchEndpoint, domainOverrideReqOpts.AddEndpoint, domainOverrideReqOpts.SearchEndpoint, domainOverrideReqOpts.SearchEndpoint, domainOverrideReqOpts.AddEndpoint, domainOverrideReqOpts.ReconfigureEndpoint, domainOverrideReqOpts.GetEndpoint + "/second"}
	if !reflect.DeepEqual(gotSequence, want) {
		t.Fatalf("create critical sections overlapped: sequence=%v want=%v", gotSequence, want)
	}
	if gotAddCalls != 2 {
		t.Fatalf("add calls=%d, want 2 serialized attempts", gotAddCalls)
	}
	errorsSeen := 0
	for _, response := range responses {
		if response.Diagnostics.HasError() {
			errorsSeen++
		}
	}
	if errorsSeen != 1 {
		t.Fatalf("responses=%v, want exactly one ambiguous failure and one success", responses)
	}
}

func TestDNSBLCreateSerializesPreflightMutationAndRecovery(t *testing.T) {
	var mu sync.Mutex
	sequence := make([]string, 0, 5)
	searchCalls, addCalls := 0, 0
	recoveryEntered := make(chan struct{})
	releaseRecovery := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		mu.Lock()
		sequence = append(sequence, request.URL.Path)
		switch request.URL.Path {
		case dnsblReqOpts.SearchEndpoint:
			searchCalls++
			call := searchCalls
			mu.Unlock()
			if call == 2 {
				close(recoveryEntered)
				<-releaseRecovery
			}
			_, _ = w.Write([]byte(`{"rows":[],"rowCount":0,"total":0,"current":1}`))
		case dnsblReqOpts.AddEndpoint:
			addCalls++
			call := addCalls
			mu.Unlock()
			if call == 1 {
				_, _ = w.Write([]byte(`{}`))
				return
			}
			_, _ = w.Write([]byte(`{"result":"failed"}`))
		default:
			mu.Unlock()
			t.Errorf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	createLock := newInstrumentedCreateLocker()
	owner := &dnsblResource{client: testClient(t, server.URL), createLock: createLock}
	model := dnsblTestModel("192.0.2.31", false)
	plan := unboundResourceState(t, owner, &model)
	responses := []*resource.CreateResponse{{State: tfsdk.State{Schema: plan.Schema}}, {State: tfsdk.State{Schema: plan.Schema}}}
	for _, response := range responses {
		initializeUnboundPrivate(t, response)
	}
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		owner.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(plan)}, responses[0])
	}()
	select {
	case <-recoveryEntered:
	case <-time.After(time.Second):
		close(releaseRecovery)
		t.Fatal("first DNSBL create did not reach ambiguous-outcome recovery")
	}
	workers.Add(1)
	go func() {
		defer workers.Done()
		owner.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(plan)}, responses[1])
	}()
	select {
	case <-createLock.secondAttempting:
	case <-time.After(time.Second):
		close(releaseRecovery)
		t.Fatal("second DNSBL create did not attempt the owner lock")
	}
	mu.Lock()
	blockedSearchCalls, blockedAddCalls := searchCalls, addCalls
	mu.Unlock()
	if blockedSearchCalls != 2 || blockedAddCalls != 1 {
		close(releaseRecovery)
		t.Fatalf("second DNSBL create entered while first recovery held lock: searches=%d adds=%d", blockedSearchCalls, blockedAddCalls)
	}
	close(releaseRecovery)
	workers.Wait()

	mu.Lock()
	gotSequence := append([]string(nil), sequence...)
	gotSearchCalls, gotAddCalls := searchCalls, addCalls
	mu.Unlock()
	want := []string{dnsblReqOpts.SearchEndpoint, dnsblReqOpts.AddEndpoint, dnsblReqOpts.SearchEndpoint, dnsblReqOpts.SearchEndpoint, dnsblReqOpts.AddEndpoint}
	if !reflect.DeepEqual(gotSequence, want) || gotSearchCalls != 3 || gotAddCalls != 2 {
		t.Fatalf("DNSBL create critical sections overlapped: sequence=%v searches=%d adds=%d", gotSequence, gotSearchCalls, gotAddCalls)
	}
	for _, response := range responses {
		if !response.Diagnostics.HasError() {
			t.Fatalf("expected both injected DNSBL creates to fail: responses=%v", responses)
		}
	}
}

func TestDomainOverridePendingUpdateRetriesActivationWithoutMutation(t *testing.T) {
	t.Parallel()
	updateCalls, reconfigureCalls := 0, 0
	allowActivation := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case domainOverrideReqOpts.GetEndpoint + "/forward-id":
			_, _ = w.Write([]byte(`{"dot":{"enabled":"1","domain":"example.test","type":{"forward":{"value":"forward","selected":1}},"server":"192.0.2.54","port":"53","verify":"","description":"updated"}}`))
		case domainOverrideReqOpts.UpdateEndpoint + "/forward-id":
			updateCalls++
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case domainOverrideReqOpts.ReconfigureEndpoint:
			reconfigureCalls++
			if !allowActivation {
				http.Error(w, "injected", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		default:
			t.Errorf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	owner := &domainOverrideResource{client: testClient(t, server.URL)}
	stateModel := DomainOverrideResourceModel{ID: types.StringValue("forward-id"), Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue("old")}
	planModel := stateModel
	planModel.Server = types.StringValue("192.0.2.54")
	planModel.Description = types.StringValue("updated")
	state := unboundResourceState(t, owner, &stateModel)
	plan := unboundResourceState(t, owner, &planModel)
	first := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}}
	initializeUnboundPrivate(t, &first)
	owner.Update(context.Background(), resource.UpdateRequest{State: state, Plan: tfsdk.Plan(plan)}, &first)
	if !first.Diagnostics.HasError() {
		t.Fatal("expected activation failure")
	}
	if pending, diagnostics := getDomainOverridePending(context.Background(), first.Private); diagnostics.HasError() || pending != domainOverridePendingMutation {
		t.Fatalf("pending=%q diagnostics=%v operation=%v", pending, diagnostics, first.Diagnostics)
	}
	mutationsAfterFailure := updateCalls
	allowActivation = true
	second := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}, Private: first.Private}
	owner.Update(context.Background(), resource.UpdateRequest{State: first.State, Plan: tfsdk.Plan(plan), Private: first.Private}, &second)
	if second.Diagnostics.HasError() {
		t.Fatal(second.Diagnostics)
	}
	if updateCalls != mutationsAfterFailure || reconfigureCalls != 3 {
		t.Fatalf("mutation was repeated during pending recovery: before=%d after=%d reconfigure=%d", mutationsAfterFailure, updateCalls, reconfigureCalls)
	}
}

func TestUnboundSingletonDeleteIsStateOnly(t *testing.T) {
	t.Parallel()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	t.Cleanup(server.Close)
	owner := &dnsblResource{client: testClient(t, server.URL)}
	owner.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
	generalOwner := &generalResource{client: testClient(t, server.URL)}
	generalOwner.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
	if calls != 0 {
		t.Fatalf("state-only delete made %d appliance calls", calls)
	}
}

func TestUnboundSingletonDeleteRetriesPendingActivation(t *testing.T) {
	for _, test := range []struct {
		name string
		call func(resource.DeleteRequest, *resource.DeleteResponse)
		set  func(context.Context, unboundPrivateState, bool) diag.Diagnostics
	}{
		{
			name: "general",
			call: func(request resource.DeleteRequest, response *resource.DeleteResponse) {
				(&generalResource{client: pendingDeleteTestClient(t)}).Delete(context.Background(), request, response)
			},
			set: setUnboundPending,
		},
		{
			name: "dnsbl",
			call: func(request resource.DeleteRequest, response *resource.DeleteResponse) {
				(&dnsblResource{client: pendingDeleteTestClient(t)}).Delete(context.Background(), request, response)
			},
			set: setDNSBLPending,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := resource.DeleteResponse{}
			initializeUnboundPrivate(t, &response)
			response.Diagnostics.Append(test.set(context.Background(), response.Private, true)...)
			test.call(resource.DeleteRequest{Private: response.Private}, &response)
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
		})
	}
}

func pendingDeleteTestClient(t *testing.T) *opnsense.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/core/firmware/status" {
			writeCurrentFirmware(w)
			return
		}
		if request.URL.Path != generalReqOpts.ReconfigureEndpoint && request.URL.Path != dnsblReqOpts.ReconfigureEndpoint {
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(server.Close)
	return testClient(t, server.URL)
}

func TestDomainOverridePendingDeleteRetriesActivationWithoutDelete(t *testing.T) {
	t.Parallel()
	deleteCalls, reconfigureCalls := 0, 0
	allowActivation := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case domainOverrideReqOpts.DeleteEndpoint + "/forward-id":
			deleteCalls++
			_, _ = w.Write([]byte(`{"result":"deleted"}`))
		case domainOverrideReqOpts.ReconfigureEndpoint:
			reconfigureCalls++
			if !allowActivation {
				http.Error(w, "injected", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/core/firmware/status":
			writeCurrentFirmware(w)
		default:
			t.Errorf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	owner := &domainOverrideResource{client: testClient(t, server.URL)}
	model := DomainOverrideResourceModel{ID: types.StringValue("forward-id"), Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue("test")}
	state := unboundResourceState(t, owner, &model)
	first := resource.DeleteResponse{}
	initializeUnboundPrivate(t, &first)
	owner.Delete(context.Background(), resource.DeleteRequest{State: state}, &first)
	if !first.Diagnostics.HasError() {
		t.Fatal("expected activation failure")
	}
	if pending, diagnostics := getDomainOverridePending(context.Background(), first.Private); diagnostics.HasError() || pending != domainOverridePendingDelete {
		t.Fatalf("pending=%q diagnostics=%v operation=%v", pending, diagnostics, first.Diagnostics)
	}
	mutationsAfterFailure := deleteCalls
	allowActivation = true
	second := resource.DeleteResponse{Private: first.Private}
	owner.Delete(context.Background(), resource.DeleteRequest{State: state, Private: first.Private}, &second)
	if second.Diagnostics.HasError() {
		t.Fatal(second.Diagnostics)
	}
	if deleteCalls != mutationsAfterFailure || reconfigureCalls != 3 {
		t.Fatalf("delete repeated during pending recovery: before=%d after=%d reconfigure=%d", mutationsAfterFailure, deleteCalls, reconfigureCalls)
	}
}

func TestDomainOverrideLegacyDeleteUsesCompatibilityOptions(t *testing.T) {
	t.Parallel()
	deleteCalls, reconfigureCalls := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/core/firmware/status":
			_, _ = w.Write([]byte(`{"product":{"product_name":"OPNsense","product_id":"opnsense","product_version":"26.7.6"}}`))
		case domainOverrideReqOpts.DeleteEndpoint + "/forward-id":
			deleteCalls++
			_, _ = w.Write([]byte(`{"result":"deleted"}`))
		case domainOverrideReqOpts.ReconfigureEndpoint:
			reconfigureCalls++
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	owner := &domainOverrideResource{client: testClient(t, server.URL)}
	model := DomainOverrideResourceModel{ID: types.StringValue("forward-id"), Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue("legacy")}
	state := unboundResourceState(t, owner, &model)
	response := resource.DeleteResponse{}
	initializeUnboundPrivate(t, &response)
	owner.Delete(context.Background(), resource.DeleteRequest{State: state}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("legacy delete diagnostics: %v", response.Diagnostics)
	}
	if deleteCalls != 1 || reconfigureCalls != 1 {
		t.Fatalf("legacy delete calls=%d reconfigure=%d, want 1/1", deleteCalls, reconfigureCalls)
	}
}

func TestDomainOverrideUpdateAfterPendingDeleteRecreatesObject(t *testing.T) {
	sequence := make([]string, 0, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		sequence = append(sequence, request.URL.Path)
		switch request.URL.Path {
		case domainOverrideReqOpts.ReconfigureEndpoint:
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case domainOverrideReqOpts.SearchEndpoint:
			_, _ = w.Write([]byte(`{"rows":[],"rowCount":0,"total":0,"current":1}`))
		case domainOverrideReqOpts.AddEndpoint:
			_, _ = w.Write([]byte(`{"result":"saved","uuid":"replacement"}`))
		case domainOverrideReqOpts.GetEndpoint + "/replacement":
			_, _ = w.Write([]byte(`{"dot":{"enabled":"1","domain":"example.test","type":{"forward":{"value":"forward","selected":1}},"server":"192.0.2.54","port":"53","verify":"","description":"replacement"}}`))
		default:
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
	}))
	defer server.Close()
	owner := &domainOverrideResource{client: testClient(t, server.URL)}
	stateModel := DomainOverrideResourceModel{ID: types.StringValue("deleted"), Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue("old")}
	planModel := stateModel
	planModel.ID = types.StringNull()
	planModel.Server = types.StringValue("192.0.2.54")
	planModel.Description = types.StringValue("replacement")
	state := unboundResourceState(t, owner, &stateModel)
	plan := unboundResourceState(t, owner, &planModel)
	response := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}}
	initializeUnboundPrivate(t, &response)
	response.Diagnostics.Append(setDomainOverridePending(context.Background(), response.Private, domainOverridePendingDelete)...)
	owner.Update(context.Background(), resource.UpdateRequest{State: state, Plan: tfsdk.Plan(plan), Private: response.Private}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	want := []string{domainOverrideReqOpts.ReconfigureEndpoint, domainOverrideReqOpts.SearchEndpoint, domainOverrideReqOpts.AddEndpoint, domainOverrideReqOpts.ReconfigureEndpoint, domainOverrideReqOpts.GetEndpoint + "/replacement"}
	if !reflect.DeepEqual(sequence, want) {
		t.Fatalf("sequence=%#v, want %#v", sequence, want)
	}
}

func TestGeneralPendingActivationThenChangedPlanDoesNotReplayOldMutation(t *testing.T) {
	t.Parallel()
	remotePort, remoteHide := "53", "0"
	updatePorts := make([]string, 0, 2)
	reconfigureCalls := 0
	allowActivation := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case generalReqOpts.UpdateEndpoint:
			var envelope struct {
				Unbound struct {
					General  map[string]string `json:"general"`
					Advanced map[string]string `json:"advanced"`
				} `json:"unbound"`
			}
			if err := json.NewDecoder(request.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			remotePort = envelope.Unbound.General["port"]
			remoteHide = envelope.Unbound.Advanced["hideidentity"]
			updatePorts = append(updatePorts, remotePort)
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case generalReqOpts.ReconfigureEndpoint:
			reconfigureCalls++
			if !allowActivation {
				http.Error(w, "injected", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case generalReqOpts.GetEndpoint:
			_, _ = w.Write(completeUnboundSettingsResponse(t, remotePort, remoteHide, "0", nil))
		default:
			t.Errorf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	owner := &generalResource{client: testClient(t, server.URL)}
	stateModel := GeneralResourceModel{
		ID:                types.StringValue(generalID),
		Port:              types.StringValue("53"),
		HideIdentity:      types.BoolValue(false),
		ActiveInterface:   types.SetNull(types.StringType),
		OutgoingInterface: types.SetNull(types.StringType),
		PrivateDomains:    types.SetNull(types.StringType),
		PrivateAddresses:  types.SetNull(types.StringType),
		InsecureDomains:   types.SetNull(types.StringType),
	}
	firstPlan := stateModel
	firstPlan.Port = types.StringValue("5353")
	firstPlan.HideIdentity = types.BoolValue(true)
	secondPlan := firstPlan
	secondPlan.Port = types.StringValue("5354")
	secondPlan.HideIdentity = types.BoolValue(false)
	state := unboundResourceState(t, owner, &stateModel)
	firstPlanState := unboundResourceState(t, owner, &firstPlan)
	first := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}}
	initializeUnboundPrivate(t, &first)
	owner.Update(context.Background(), resource.UpdateRequest{State: state, Plan: tfsdk.Plan(firstPlanState), Private: first.Private}, &first)
	if !first.Diagnostics.HasError() {
		t.Fatal("expected reconfigure_general failure")
	}
	if pending, diagnostics := getUnboundPending(context.Background(), first.Private); diagnostics.HasError() || !pending {
		t.Fatalf("pending=%t diagnostics=%v operation=%v", pending, diagnostics, first.Diagnostics)
	}
	allowActivation = true
	secondPlanState := unboundResourceState(t, owner, &secondPlan)
	second := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}, Private: first.Private}
	owner.Update(context.Background(), resource.UpdateRequest{State: first.State, Plan: tfsdk.Plan(secondPlanState), Private: first.Private}, &second)
	if second.Diagnostics.HasError() {
		t.Fatal(second.Diagnostics)
	}
	if !reflect.DeepEqual(updatePorts, []string{"5353", "5354"}) {
		t.Fatalf("settings mutations = %#v; old mutation was replayed or changed plan was skipped", updatePorts)
	}
	if reconfigureCalls != 4 {
		t.Fatalf("reconfigure_general calls = %d, want two failed attempts, one retry, one changed-plan activation", reconfigureCalls)
	}
}

func TestDNSBLPendingActivationThenChangedPlanDoesNotReplayOldMutation(t *testing.T) {
	t.Parallel()
	remoteAddress, remoteSafeSearch := "192.0.2.30", "0"
	itemMutations, safeSearchMutations, reconfigureCalls := 0, 0, 0
	addresses := make([]string, 0, 2)
	allowActivation := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case dnsblReqOpts.GetEndpoint + "/owned":
			_, _ = fmt.Fprintf(w, `{"blocklist":{"enabled":"0","type":[],"lists":[],"allowlists":[],"blocklists":[],"wildcards":[],"source_nets":[],"address":%q,"nxdomain":"0","cache_ttl":"72000","description":"owned"}}`, remoteAddress)
		case dnsblReqOpts.UpdateEndpoint + "/owned":
			var envelope struct {
				Blocklist struct {
					Address string `json:"address"`
				} `json:"blocklist"`
			}
			if err := json.NewDecoder(request.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			remoteAddress = envelope.Blocklist.Address
			addresses = append(addresses, remoteAddress)
			itemMutations++
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case dnsblSettingsReqOpts.GetEndpoint:
			_, _ = fmt.Fprintf(w, `{"unbound":{"general":{"safesearch":%q}}}`, remoteSafeSearch)
		case dnsblSettingsReqOpts.UpdateEndpoint:
			var envelope struct {
				Unbound struct {
					General map[string]string `json:"general"`
				} `json:"unbound"`
			}
			if err := json.NewDecoder(request.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			remoteSafeSearch = envelope.Unbound.General["safesearch"]
			safeSearchMutations++
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case dnsblReqOpts.ReconfigureEndpoint:
			reconfigureCalls++
			if !allowActivation {
				http.Error(w, "injected", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/core/firmware/status":
			writeCurrentFirmware(w)
		default:
			t.Errorf("unexpected route %s", request.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	owner := &dnsblResource{client: testClient(t, server.URL)}
	stateModel := dnsblTestModel("192.0.2.30", false)
	firstPlan := dnsblTestModel("192.0.2.31", true)
	secondPlan := dnsblTestModel("192.0.2.32", false)
	state := unboundResourceState(t, owner, &stateModel)
	firstPlanState := unboundResourceState(t, owner, &firstPlan)
	first := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}}
	initializeUnboundPrivate(t, &first)
	first.Diagnostics.Append(setDNSBLUUID(context.Background(), first.Private, "owned")...)
	owner.Update(context.Background(), resource.UpdateRequest{State: state, Plan: tfsdk.Plan(firstPlanState), Private: first.Private}, &first)
	if !first.Diagnostics.HasError() {
		t.Fatal("expected normal Unbound reconfigure failure")
	}
	allowActivation = true
	secondPlanState := unboundResourceState(t, owner, &secondPlan)
	second := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}, Private: first.Private}
	owner.Update(context.Background(), resource.UpdateRequest{State: first.State, Plan: tfsdk.Plan(secondPlanState), Private: first.Private}, &second)
	if second.Diagnostics.HasError() {
		t.Fatal(second.Diagnostics)
	}
	if itemMutations != 2 || !reflect.DeepEqual(addresses, []string{"192.0.2.31", "192.0.2.32"}) {
		t.Fatalf("item mutations=%d addresses=%v; old mutation was replayed or changed plan was skipped", itemMutations, addresses)
	}
	if safeSearchMutations != 2 || remoteSafeSearch != "0" {
		t.Fatalf("SafeSearch mutations=%d final=%q", safeSearchMutations, remoteSafeSearch)
	}
	if reconfigureCalls != 4 {
		t.Fatalf("normal reconfigure calls=%d, want two failed attempts, one retry, one changed-plan activation", reconfigureCalls)
	}
}

func dnsblTestModel(address string, safeSearch bool) DNSBLResourceModel {
	return DNSBLResourceModel{
		ID: types.StringValue(dnsblID), Enabled: types.BoolValue(false), SafeSearch: types.BoolValue(safeSearch),
		Type: types.SetValueMust(types.StringType, nil), Lists: types.SetValueMust(types.StringType, nil),
		Whitelists: types.SetValueMust(types.StringType, nil), Blocklists: types.SetValueMust(types.StringType, nil),
		Wildcards: types.SetValueMust(types.StringType, nil), Address: types.StringValue(address), NXDomain: types.BoolValue(false),
	}
}

func unboundResourceState(t *testing.T, owner resource.Resource, model any) tfsdk.State {
	t.Helper()
	var response resource.SchemaResponse
	owner.Schema(context.Background(), resource.SchemaRequest{}, &response)
	state := tfsdk.State{Schema: response.Schema}
	if diagnostics := state.Set(context.Background(), model); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	return state
}

func initializeUnboundPrivate(t *testing.T, response any) {
	t.Helper()
	value := reflect.ValueOf(response).Elem().FieldByName("Private")
	value.Set(reflect.New(value.Type().Elem()))
}

func TestGeneralAmbiguousOutcomeComparesOnlyEmittedFields(t *testing.T) {
	remote := GeneralResourceModel{Port: types.StringValue("5353"), HideIdentity: types.BoolValue(true), ForwardingEnabled: types.BoolValue(true)}
	payload := generalSettingsPayload(context.Background(), GeneralResourceModel{Port: types.StringValue("5353")})
	if !generalPayloadMatchesRemote(context.Background(), payload, remote) {
		t.Fatal("omitted remote fields incorrectly affected mutation confirmation")
	}
	payload["general"].(map[string]any)["port"] = "53"
	if generalPayloadMatchesRemote(context.Background(), payload, remote) {
		t.Fatal("changed emitted field incorrectly matched")
	}
}

func TestStrictUnboundDecodersRejectMalformedAuthoritativeFields(t *testing.T) {
	for name, mutate := range map[string]func(*testing.T, *unboundSettingsEnvelope){
		"missing audited key": func(t *testing.T, envelope *unboundSettingsEnvelope) {
			envelope.Advanced = removeJSONField(t, envelope.Advanced, "serveexpiredclienttimeout")
		},
		"invalid boolean": func(t *testing.T, envelope *unboundSettingsEnvelope) {
			envelope.Advanced = replaceJSONFields(t, envelope.Advanced, map[string]any{"hideidentity": "yes"})
		},
		"negative integer": func(t *testing.T, envelope *unboundSettingsEnvelope) {
			envelope.Advanced = replaceJSONFields(t, envelope.Advanced, map[string]any{"outgoingnumtcp": "-1"})
		},
		"out of range log level": func(t *testing.T, envelope *unboundSettingsEnvelope) {
			envelope.Advanced = replaceJSONFields(t, envelope.Advanced, map[string]any{"logverbosity": "6"})
		},
		"invalid ACL choice": func(t *testing.T, envelope *unboundSettingsEnvelope) {
			envelope.ACLs = replaceJSONFields(t, envelope.ACLs, map[string]any{"default_action": "bogus"})
		},
	} {
		t.Run(name, func(t *testing.T) {
			envelope := completeUnboundSettingsEnvelope(t, "53", "0", "0", nil)
			mutate(t, envelope)
			if _, err := flattenGeneralSettings(envelope); err == nil {
				t.Fatal("malformed authoritative settings were accepted")
			}
		})
	}
	for name, raw := range map[string]string{
		"invalid boolean":  `{"blocklist":{"id":{"enabled":"yes","type":[],"lists":[],"allowlists":[],"blocklists":[],"wildcards":[],"source_nets":[],"address":"0.0.0.0","nxdomain":"0","cache_ttl":"1","description":"x"}}}`,
		"invalid selector": `{"blocklist":{"id":{"enabled":"1","type":{"ag":{"selected":2}},"lists":[],"allowlists":[],"blocklists":[],"wildcards":[],"source_nets":[],"address":"0.0.0.0","nxdomain":"0","cache_ttl":"1","description":"x"}}}`,
		"invalid ttl":      `{"blocklist":{"id":{"enabled":"1","type":[],"lists":[],"allowlists":[],"blocklists":[],"wildcards":[],"source_nets":[],"address":"0.0.0.0","nxdomain":"0","cache_ttl":"-1","description":"x"}}}`,
	} {
		t.Run("DNSBL "+name, func(t *testing.T) {
			if _, err := decodeDNSBLItems(json.RawMessage(raw)); err == nil {
				t.Fatal("malformed DNSBL item was accepted")
			}
		})
	}
	var forward domainOverrideAPIResponse
	if err := json.Unmarshal([]byte(`{"enabled":"1","domain":"example.test","type":{"forward":{"selected":1},"dot":{"selected":1}},"server":"192.0.2.1","port":"53","verify":"","description":"x"}`), &forward); err == nil {
		t.Fatal("ambiguous forward selector was accepted")
	}
}

func TestCanonicalJSONPreservesLargeNumbers(t *testing.T) {
	canonical, err := canonicalJSON(json.RawMessage(`{"extension":900719925474099312345}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(canonical) != `{"extension":900719925474099312345}` {
		t.Fatalf("large number lost precision: %s", canonical)
	}
}

func TestDNSBLLegacyStateSelectsExactlyOneMatchingItem(t *testing.T) {
	for name, addresses := range map[string][]string{"exact": {"192.0.2.31", "192.0.2.32"}, "zero": {"192.0.2.32", "192.0.2.33"}, "multiple": {"192.0.2.31", "192.0.2.31"}} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case dnsblReqOpts.SearchEndpoint:
					_, _ = w.Write([]byte(`{"rows":[{"uuid":"a"},{"uuid":"b"}],"rowCount":2,"total":2,"current":1}`))
				case dnsblReqOpts.GetEndpoint + "/a":
					_, _ = fmt.Fprintf(w, `{"blocklist":%s}`, dnsblItemJSON(addresses[0]))
				case dnsblReqOpts.GetEndpoint + "/b":
					_, _ = fmt.Fprintf(w, `{"blocklist":%s}`, dnsblItemJSON(addresses[1]))
				case dnsblSettingsReqOpts.GetEndpoint:
					_, _ = w.Write([]byte(`{"unbound":{"general":{"safesearch":"0"}}}`))
				default:
					t.Fatalf("unexpected route %s", request.URL.Path)
				}
			}))
			defer server.Close()
			prior := dnsblTestModel("192.0.2.31", false)
			uuid, err := discoverDNSBLUUID(context.Background(), testClient(t, server.URL), &prior)
			if name == "exact" {
				if err != nil || uuid != "a" {
					t.Fatalf("uuid=%q err=%v", uuid, err)
				}
			} else if err == nil {
				t.Fatalf("ambiguous %s result adopted %q", name, uuid)
			}
		})
	}
}

func TestDNSBLSoleLegacyItemRequiresExactPriorMatch(t *testing.T) {
	for _, test := range []struct {
		name       string
		remote     string
		safeSearch bool
		wantUUID   string
		wantError  bool
	}{
		{"exact ignores SafeSearch", "192.0.2.31", true, "only", false},
		{"mismatch rejected", "192.0.2.32", false, "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case dnsblReqOpts.SearchEndpoint:
					_, _ = w.Write([]byte(`{"rows":[{"uuid":"only"}],"rowCount":1,"total":1,"current":1}`))
				case dnsblReqOpts.GetEndpoint + "/only":
					_, _ = fmt.Fprintf(w, `{"blocklist":%s}`, dnsblItemJSON(test.remote))
				default:
					t.Fatalf("unexpected route %s", request.URL.Path)
				}
			}))
			defer server.Close()
			prior := dnsblTestModel("192.0.2.31", test.safeSearch)
			uuid, err := discoverDNSBLUUID(context.Background(), testClient(t, server.URL), &prior)
			if (err != nil) != test.wantError || uuid != test.wantUUID {
				t.Fatalf("uuid=%q err=%v, want uuid=%q error=%t", uuid, err, test.wantUUID, test.wantError)
			}
		})
	}
}

func TestDomainOverridePendingMarkerClosedSetAndNilSafe(t *testing.T) {
	if diagnostics := setDomainOverridePending(context.Background(), nil, domainOverridePendingMutation); diagnostics.HasError() {
		t.Fatalf("nil private state should be a no-op: %v", diagnostics)
	}
	if diagnostics := setDomainOverridePending(context.Background(), nil, "unknown"); !diagnostics.HasError() {
		t.Fatal("unknown marker was accepted for nil private state")
	}
	response := resource.ReadResponse{}
	initializeUnboundPrivate(t, &response)
	if diagnostics := setDomainOverridePending(context.Background(), response.Private, "unknown"); !diagnostics.HasError() {
		t.Fatal("unknown marker was accepted")
	}
	if diagnostics := response.Private.SetKey(context.Background(), domainOverridePendingKey, []byte(`"unknown"`)); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if _, diagnostics := getDomainOverridePending(context.Background(), response.Private); !diagnostics.HasError() {
		t.Fatal("unknown persisted marker was accepted")
	}
}

func TestDomainOverrideAmbiguousCreateRecoveryRequiresOneExactNewMatch(t *testing.T) {
	plan := DomainOverrideResourceModel{Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue("wanted")}
	for name, rows := range map[string]string{
		"exact":    `[{"uuid":"old","enabled":"1","domain":"example.test","server":"192.0.2.53","port":"53","verify":"","description":"wanted"},{"uuid":"new","enabled":"1","domain":"example.test","server":"192.0.2.53","port":"53","verify":"","description":"wanted"}]`,
		"zero":     `[{"uuid":"new","enabled":"1","domain":"example.test","server":"192.0.2.54","port":"53","verify":"","description":"wanted"}]`,
		"multiple": `[{"uuid":"new1","enabled":"1","domain":"example.test","server":"192.0.2.53","port":"53","verify":"","description":"wanted"},{"uuid":"new2","enabled":"1","domain":"example.test","server":"192.0.2.53","port":"53","verify":"","description":"wanted"}]`,
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				var decoded []json.RawMessage
				if err := json.Unmarshal([]byte(rows), &decoded); err != nil {
					t.Fatal(err)
				}
				_, _ = fmt.Fprintf(w, `{"rows":%s,"rowCount":%d,"total":%d,"current":1}`, rows, len(decoded), len(decoded))
			}))
			defer server.Close()
			uuid, err := recoverDomainOverrideID(context.Background(), testClient(t, server.URL), plan, map[string]struct{}{"old": {}})
			switch name {
			case "exact":
				if err != nil || uuid != "new" {
					t.Fatalf("uuid=%q err=%v", uuid, err)
				}
			case "zero":
				if err != nil || uuid != "" {
					t.Fatalf("uuid=%q err=%v", uuid, err)
				}
			case "multiple":
				if err == nil {
					t.Fatal("multiple exact matches accepted")
				}
			}
		})
	}
}

func TestDNSBLCreateReadBackFailureRetainsUsableIdentityState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case dnsblReqOpts.SearchEndpoint:
			_, _ = w.Write([]byte(`{"rows":[],"rowCount":0,"total":0,"current":1}`))
		case dnsblReqOpts.AddEndpoint:
			_, _ = w.Write([]byte(`{"result":"saved","uuid":"created"}`))
		case dnsblSettingsReqOpts.GetEndpoint:
			_, _ = w.Write([]byte(`{"unbound":{"general":{"safesearch":"0"}}}`))
		case dnsblReqOpts.ReconfigureEndpoint:
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case dnsblReqOpts.GetEndpoint + "/created":
			http.Error(w, "injected read failure", http.StatusInternalServerError)
		default:
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
	}))
	defer server.Close()
	owner := &dnsblResource{client: testClient(t, server.URL)}
	model := dnsblTestModel("192.0.2.31", false)
	plan := unboundResourceState(t, owner, &model)
	response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	initializeUnboundPrivate(t, &response)
	owner.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(plan)}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("expected post-create read diagnostic")
	}
	var state DNSBLResourceModel
	if diagnostics := response.State.Get(context.Background(), &state); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if state.ID.ValueString() != dnsblID {
		t.Fatalf("public identity was not retained: %#v", state)
	}
	if uuid, diagnostics := getDNSBLUUID(context.Background(), response.Private); diagnostics.HasError() || uuid != "created" {
		t.Fatalf("private UUID=%q diagnostics=%v", uuid, diagnostics)
	}
}

func TestGeneralReadRetriesPendingActivationAndRefreshesAuthoritativeState(t *testing.T) {
	reconfigureCalls, getCalls := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case generalReqOpts.ReconfigureEndpoint:
			reconfigureCalls++
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case generalReqOpts.GetEndpoint:
			getCalls++
			_, _ = w.Write(completeUnboundSettingsResponse(t, "5353", "1", "0", nil))
		case "/api/core/firmware/status":
			writeCurrentFirmware(w)
		default:
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
	}))
	defer server.Close()
	owner := &generalResource{client: testClient(t, server.URL)}
	api, _ := flattenGeneralSettings(completeUnboundSettingsEnvelope(t, "53", "0", "0", nil))
	var initial GeneralResourceModel
	initial.fromAPI(context.Background(), api, generalID)
	state := unboundResourceState(t, owner, &initial)
	response := resource.ReadResponse{State: tfsdk.State{Schema: state.Schema}}
	initializeUnboundPrivate(t, &response)
	response.Diagnostics.Append(setUnboundPending(context.Background(), response.Private, true)...)
	owner.Read(context.Background(), resource.ReadRequest{State: state, Private: response.Private}, &response)
	if response.Diagnostics.HasError() || reconfigureCalls != 1 || getCalls != 1 {
		t.Fatalf("diagnostics=%v reconfigure=%d get=%d", response.Diagnostics, reconfigureCalls, getCalls)
	}
}

func TestDNSBLReadRetriesPendingActivationWithoutMutation(t *testing.T) {
	reconfigureCalls, getCalls := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case dnsblReqOpts.ReconfigureEndpoint:
			reconfigureCalls++
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case dnsblReqOpts.GetEndpoint + "/owned":
			getCalls++
			_, _ = fmt.Fprintf(w, `{"blocklist":%s}`, dnsblItemJSON("192.0.2.31"))
		case dnsblSettingsReqOpts.GetEndpoint:
			_, _ = w.Write([]byte(`{"unbound":{"general":{"safesearch":"1"}}}`))
		case "/api/core/firmware/status":
			writeCurrentFirmware(w)
		default:
			t.Fatalf("unexpected route %s", request.URL.Path)
		}
	}))
	defer server.Close()
	owner := &dnsblResource{client: testClient(t, server.URL)}
	initial := dnsblTestModel("192.0.2.30", false)
	state := unboundResourceState(t, owner, &initial)
	response := resource.ReadResponse{State: tfsdk.State{Schema: state.Schema}}
	initializeUnboundPrivate(t, &response)
	response.Diagnostics.Append(setDNSBLUUID(context.Background(), response.Private, "owned")...)
	response.Diagnostics.Append(setDNSBLPending(context.Background(), response.Private, true)...)
	owner.Read(context.Background(), resource.ReadRequest{State: state, Private: response.Private}, &response)
	if response.Diagnostics.HasError() || reconfigureCalls != 1 || getCalls != 1 {
		t.Fatalf("diagnostics=%v reconfigure=%d get=%d", response.Diagnostics, reconfigureCalls, getCalls)
	}
}

func TestDomainOverrideReadRecoversPendingMutationAndDelete(t *testing.T) {
	for _, pending := range []string{domainOverridePendingMutation, domainOverridePendingDelete} {
		t.Run(pending, func(t *testing.T) {
			reconfigureCalls, getCalls := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case domainOverrideReqOpts.ReconfigureEndpoint:
					reconfigureCalls++
					_, _ = w.Write([]byte(`{"status":"ok"}`))
				case domainOverrideReqOpts.GetEndpoint + "/forward-id":
					getCalls++
					_, _ = w.Write([]byte(`{"dot":{"enabled":"1","domain":"example.test","type":"forward","server":"192.0.2.54","port":"53","verify":"","description":"remote"}}`))
				case "/api/core/firmware/status":
					writeCurrentFirmware(w)
				default:
					t.Fatalf("unexpected route %s", request.URL.Path)
				}
			}))
			defer server.Close()
			owner := &domainOverrideResource{client: testClient(t, server.URL)}
			model := DomainOverrideResourceModel{ID: types.StringValue("forward-id"), Enabled: types.BoolValue(true), Domain: types.StringValue("example.test"), Type: types.StringValue("query"), Server: types.StringValue("192.0.2.53"), ServerPort: types.Int64Value(53), VerifyCN: types.StringValue(""), Description: types.StringValue("old")}
			state := unboundResourceState(t, owner, &model)
			response := resource.ReadResponse{State: tfsdk.State{Schema: state.Schema}}
			initializeUnboundPrivate(t, &response)
			response.Diagnostics.Append(setDomainOverridePending(context.Background(), response.Private, pending)...)
			owner.Read(context.Background(), resource.ReadRequest{State: state, Private: response.Private}, &response)
			if response.Diagnostics.HasError() || reconfigureCalls != 1 {
				t.Fatalf("diagnostics=%v reconfigure=%d get=%d", response.Diagnostics, reconfigureCalls, getCalls)
			}
			if pending == domainOverridePendingDelete && getCalls != 0 {
				t.Fatalf("pending delete performed %d GETs", getCalls)
			}
			if pending == domainOverridePendingMutation && getCalls != 1 {
				t.Fatalf("pending mutation GETs=%d", getCalls)
			}
		})
	}
}

func TestPrivateStateDiagnosticsStopDNSBLReadAndUpdate(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer server.Close()
	owner := &dnsblResource{client: testClient(t, server.URL)}
	model := dnsblTestModel("192.0.2.31", false)
	state := unboundResourceState(t, owner, &model)
	t.Run("read", func(t *testing.T) {
		response := resource.ReadResponse{State: tfsdk.State{Schema: state.Schema}}
		initializeUnboundPrivate(t, &response)
		response.Diagnostics.Append(response.Private.SetKey(context.Background(), dnsblItemUUIDKey, []byte("not-json"))...)
		owner.Read(context.Background(), resource.ReadRequest{State: state, Private: response.Private}, &response)
		if !response.Diagnostics.HasError() || calls != 0 {
			t.Fatalf("diagnostics=%v calls=%d", response.Diagnostics, calls)
		}
	})
	t.Run("update", func(t *testing.T) {
		response := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}}
		initializeUnboundPrivate(t, &response)
		response.Diagnostics.Append(response.Private.SetKey(context.Background(), dnsblItemUUIDKey, []byte("not-json"))...)
		owner.Update(context.Background(), resource.UpdateRequest{State: state, Plan: tfsdk.Plan(state), Private: response.Private}, &response)
		if !response.Diagnostics.HasError() || calls != 0 {
			t.Fatalf("diagnostics=%v calls=%d", response.Diagnostics, calls)
		}
	})
}

func completeUnboundSettingsEnvelope(t *testing.T, port, hideIdentity, safeSearch string, dnsbl any) *unboundSettingsEnvelope {
	t.Helper()
	marshal := func(value any) json.RawMessage {
		body, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return body
	}
	if dnsbl == nil {
		dnsbl = map[string]any{"blocklist": map[string]any{}}
	}
	return &unboundSettingsEnvelope{
		General: marshal(map[string]any{
			"enabled": "1", "port": port, "stats": "0", "active_interface": []any{}, "dnssec": "1", "dns64": "0", "dns64prefix": "", "noarecords": "0", "regdhcp": "0", "regdhcpdomain": "", "regdhcpstatic": "0", "noreglladdr6": "0", "noregrecords": "0", "txtsupport": "0", "cacheflush": "0", "safesearch": safeSearch, "local_zone_type": "transparent", "outgoing_interface": []any{}, "enable_wpad": "0",
		}),
		Advanced: marshal(map[string]any{
			"hideidentity": hideIdentity, "hideversion": "0", "prefetchkey": "0", "dnssecstripped": "0", "aggressivensec": "0", "qnameminstrict": "0", "outgoingnumtcp": "10", "incomingnumtcp": "10", "numqueriesperthread": "512", "outgoingrange": "4096", "jostletimeout": "200", "discardtimeout": "1900", "privatedomain": []any{}, "privateaddress": []any{}, "insecuredomain": []any{}, "prefetch": "0", "unwantedreplythreshold": "10000000", "msgcachesize": "4m", "rrsetcachesize": "8m", "cachemaxttl": "86400", "cachemaxnegativettl": "3600", "cacheminttl": "0", "infrahostttl": "900", "infrakeepprobing": "0", "infracachenumhosts": "10000", "extendedstatistics": "0", "logqueries": "0", "logreplies": "0", "logtagqueryreply": "0", "loglocalactions": "0", "logservfail": "0", "logverbosity": "1", "valloglevel": "0", "serveexpired": "0", "serveexpiredreplyttl": "30", "serveexpiredttl": "86400", "serveexpiredttlreset": "0", "serveexpiredclienttimeout": "180",
		}),
		ACLs:       marshal(map[string]any{"default_action": "allow"}),
		Forwarding: marshal(map[string]any{"enabled": "0"}),
		DNSBL:      marshal(dnsbl),
	}
}

func completeUnboundSettingsResponse(t *testing.T, port, hideIdentity, safeSearch string, dnsbl any) []byte {
	t.Helper()
	envelope := completeUnboundSettingsEnvelope(t, port, hideIdentity, safeSearch, dnsbl)
	body, err := json.Marshal(map[string]any{"unbound": envelope})
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func writeCurrentFirmware(w http.ResponseWriter) {
	_, _ = w.Write([]byte(`{"product":{"product_name":"OPNsense","product_id":"opnsense","product_version":"26.7.7"}}`))
}

func replaceJSONFields(t *testing.T, raw json.RawMessage, replacements map[string]any) json.RawMessage {
	t.Helper()
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for key, value := range replacements {
		fields[key] = value
	}
	result, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func removeJSONField(t *testing.T, raw json.RawMessage, key string) json.RawMessage {
	t.Helper()
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	delete(fields, key)
	result, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func dnsblItemJSON(address string) string {
	return fmt.Sprintf(`{"enabled":"0","type":[],"lists":[],"allowlists":[],"blocklists":[],"wildcards":[],"source_nets":[],"address":%q,"nxdomain":"0","cache_ttl":"72000","description":"owned"}`, address)
}
