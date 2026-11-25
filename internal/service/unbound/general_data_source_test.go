// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGeneralDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newGeneralDataSource())
	if len(schema.Attributes) != len(generalLookupFields) {
		t.Fatalf("schema has %d attributes, allowlist has %d", len(schema.Attributes), len(generalLookupFields))
	}
	for _, name := range generalLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing from the schema", name)
		}
	}
	for _, excluded := range []string{"safesearch", "belownxdomain", "password", "secret"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is published", excluded)
		}
	}
	id, ok := schema.Attributes["id"].(dsschema.StringAttribute)
	if !ok || !id.Required {
		t.Fatal("id must be a required string")
	}
}

func TestGeneralDataSource_read(t *testing.T) {
	t.Parallel()

	var paths []string
	server := generalLookupServer(t, generalFixture(t, nil), &paths)
	t.Cleanup(server.Close)

	state := readGeneralLookup(t, server.URL, generalID)
	if state.ID.ValueString() != generalID || state.Enabled.ValueBool() {
		t.Fatalf("identity or enabled = %#v %#v", state.ID, state.Enabled)
	}
	if state.Port.ValueString() != "5353" || state.LocalZoneType.ValueString() != "refuse" {
		t.Fatalf("port or zone copied a default: %#v %#v", state.Port, state.LocalZoneType)
	}
	if state.AggressiveNSEC.ValueBool() || state.VerbosityLevel.ValueInt64() != 0 || state.ValidationLevel.ValueInt64() != 0 {
		t.Fatalf("required zero values were replaced: aggressive=%#v verbosity=%#v validation=%#v", state.AggressiveNSEC, state.VerbosityLevel, state.ValidationLevel)
	}
	if !state.Stats.IsNull() || !state.OutgoingNumTCP.IsNull() || !state.ForwardingEnabled.IsNull() {
		t.Fatalf("blank optional values were synthesized: stats=%#v tcp=%#v forwarding=%#v", state.Stats, state.OutgoingNumTCP, state.ForwardingEnabled)
	}
	if state.IncomingNumTCP.ValueInt64() != 1 {
		t.Fatalf("leading-zero integer = %#v", state.IncomingNumTCP)
	}
	if got := setValues(t, state.PrivateAddresses); len(got) != 1 || got[0] != "192.0.2.0/24" {
		t.Fatalf("private addresses = %#v", got)
	}
	if got := setValues(t, state.ActiveInterface); len(got) != 1 || got[0] != "lan" {
		t.Fatalf("active interfaces = %#v", got)
	}
	if state.DNS64Prefix.ValueString() != "" || state.ACLDefaultAction.ValueString() != "deny" {
		t.Fatalf("prefix or action = %#v %#v", state.DNS64Prefix, state.ACLDefaultAction)
	}
	for _, path := range paths {
		if path != generalReqOpts.GetEndpoint {
			t.Fatalf("unexpected request path %s", path)
		}
	}
}

func TestGeneralDataSource_wrongIDDoesNotRead(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		http.Error(w, "called", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	_, diags := readGeneralLookupRaw(t, server.URL, "other")
	text := diagText(diags)
	if !strings.Contains(text, "Invalid Unbound general settings ID") || strings.Contains(text, generalNotFoundSummary) {
		t.Fatalf("diagnostics = %s", text)
	}
}

func TestGeneralDataSource_emptyObjectIsNotFound(t *testing.T) {
	t.Parallel()

	for _, body := range []string{`{}`, `{"unbound":{}}`, `{"unbound":null}`, "  {}  ", "{ }"} {
		server := generalBodyServer(t, body)
		_, diags := readGeneralLookupRaw(t, server.URL, generalID)
		server.Close()
		text := diagText(diags)
		if !strings.Contains(text, generalNotFoundSummary) || !strings.Contains(text, "resource not found") {
			t.Fatalf("body %q diagnostics = %s", body, text)
		}
	}
}

func TestGeneralDataSource_invalidJSONIsMalformed(t *testing.T) {
	t.Parallel()

	server := generalBodyServer(t, "{")
	t.Cleanup(server.Close)
	_, diags := readGeneralLookupRaw(t, server.URL, generalID)
	text := diagText(diags)
	if strings.Contains(text, generalNotFoundSummary) || strings.Contains(text, "resource not found") {
		t.Fatalf("invalid JSON was classified as not found: %s", text)
	}
	if !diags.HasError() {
		t.Fatal("expected a malformed response")
	}
}

func TestGeneralDataSource_shapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		wantErr string
		check   func(*testing.T, GeneralResourceModel)
	}{
		{
			name:    "blank enabled",
			mutate:  func(body map[string]any) { generalBranch(body, "general")["enabled"] = " " },
			wantErr: "enabled is missing or blank",
		},
		{
			name:   "disabled stays false",
			mutate: func(body map[string]any) { generalBranch(body, "general")["enabled"] = "0" },
			check: func(t *testing.T, state GeneralResourceModel) {
				t.Helper()
				if state.Enabled.ValueBool() {
					t.Fatal("disabled flag became true")
				}
			},
		},
		{
			name:    "padded bool",
			mutate:  func(body map[string]any) { generalBranch(body, "general")["stats"] = " 2 " },
			wantErr: "stats token",
		},
		{
			name:    "padded option key",
			mutate:  func(body map[string]any) { generalBranch(body, "general")["local_zone_type"] = selected(" refuse") },
			wantErr: "outside the released option list",
		},
		{
			name:   "blank optional bool stays null",
			mutate: func(body map[string]any) { generalBranch(body, "advanced")["hideidentity"] = "" },
			check: func(t *testing.T, state GeneralResourceModel) {
				t.Helper()
				if !state.HideIdentity.IsNull() {
					t.Fatalf("hide identity = %#v", state.HideIdentity)
				}
			},
		},
		{
			name:    "empty private addresses",
			mutate:  func(body map[string]any) { generalBranch(body, "advanced")["privateaddress"] = selected("") },
			wantErr: "privateaddress is missing or empty",
		},
		{
			name:   "literal zero is kept",
			mutate: func(body map[string]any) { generalBranch(body, "advanced")["cacheminttl"] = "0" },
			check: func(t *testing.T, state GeneralResourceModel) {
				t.Helper()
				if state.MinTTL.IsNull() || state.MinTTL.ValueInt64() != 0 {
					t.Fatalf("min ttl = %#v", state.MinTTL)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			body := generalFixtureMap()
			test.mutate(body)
			server := generalLookupServer(t, mustJSON(t, map[string]any{"unbound": body}), nil)
			t.Cleanup(server.Close)
			state, diags := readGeneralLookupRaw(t, server.URL, generalID)
			text := diagText(diags)
			if test.wantErr != "" {
				if !strings.Contains(text, test.wantErr) {
					t.Fatalf("diagnostics = %s", text)
				}
				return
			}
			if diags.HasError() {
				t.Fatalf("diagnostics = %s", text)
			}
			test.check(t, state)
		})
	}
}

func TestGeneralDataSource_pluginMissingFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	_, diags := readGeneralLookupRaw(t, server.URL, generalID)
	text := diagText(diags)
	if !strings.Contains(text, "plugin 'unbound' is not installed on OPNsense") || strings.Contains(text, generalNotFoundSummary) {
		t.Fatalf("diagnostics = %s", text)
	}
}

func TestGeneralDataSource_modelRejectsCopiedDefault(t *testing.T) {
	t.Parallel()

	settings := &unboundSettingsEnvelope{General: json.RawMessage(`{}`)}
	_, err := generalLookupFromEnvelope(settings)
	if err == nil || !strings.Contains(err.Error(), "enabled is missing or blank") {
		t.Fatalf("error = %v", err)
	}
}

func generalLookupServer(t *testing.T, body string, paths *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if paths != nil {
			*paths = append(*paths, r.URL.Path)
		}
		if r.Method != http.MethodGet || r.URL.Path != generalReqOpts.GetEndpoint || strings.Contains(r.URL.Path, "reconfigure") || strings.HasSuffix(r.URL.Path, "/set") {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
}

func generalBodyServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return generalLookupServer(t, body, nil)
}

func generalFixture(t *testing.T, mutate func(map[string]any)) string {
	t.Helper()
	body := generalFixtureMap()
	if mutate != nil {
		mutate(body)
	}
	return mustJSON(t, map[string]any{"unbound": body})
}

func generalFixtureMap() map[string]any {
	return map[string]any{
		"general": map[string]any{
			"enabled": "0", "port": "5353", "stats": "", "active_interface": selected("lan", "wan"),
			"dnssec": "1", "dns64": "", "dns64prefix": "", "noarecords": "0", "regdhcp": "0",
			"regdhcpdomain": "", "regdhcpstatic": "0", "noreglladdr6": "0", "noregrecords": "0",
			"txtsupport": "0", "cacheflush": "0", "safesearch": "1", "local_zone_type": selected("refuse"),
			"outgoing_interface": []any{}, "enable_wpad": "0", "ignored_note": "not-a-secret",
		},
		"advanced": map[string]any{
			"hideidentity": "", "hideversion": "0", "prefetch": "0", "prefetchkey": "0",
			"dnssecstripped": "0", "belownxdomain": "1", "aggressivensec": "0", "serveexpired": "0",
			"serveexpiredreplyttl": "", "serveexpiredttl": "", "serveexpiredttlreset": "",
			"serveexpiredclienttimeout": "", "qnameminstrict": "0", "extendedstatistics": "0",
			"logqueries": "0", "logreplies": "0", "logtagqueryreply": "0", "logservfail": "0",
			"loglocalactions": "0", "logverbosity": "0", "valloglevel": "0", "privatedomain": []any{},
			"privateaddress": selected("192.0.2.0/24"), "insecuredomain": []any{}, "msgcachesize": "",
			"rrsetcachesize": "", "outgoingnumtcp": "", "incomingnumtcp": "01", "numqueriesperthread": "0",
			"outgoingrange": "0", "jostletimeout": "0", "discardtimeout": "0", "cachemaxttl": "0",
			"cachemaxnegativettl": "0", "cacheminttl": "", "infrahostttl": "0", "infrakeepprobing": "0",
			"infracachenumhosts": "0", "unwantedreplythreshold": "0",
		},
		"acls":       map[string]any{"default_action": selected("deny"), "acl": map[string]any{"unrelated": map[string]any{}}},
		"forwarding": map[string]any{"enabled": ""},
		"dnsbl":      map[string]any{"blocklist": map[string]any{"unrelated": map[string]any{}}},
		"dots":       map[string]any{"unrelated": map[string]any{}},
		"hosts":      map[string]any{"unrelated": map[string]any{}},
		"aliases":    map[string]any{"unrelated": map[string]any{}},
	}
}

func generalBranch(body map[string]any, name string) map[string]any {
	return body[name].(map[string]any)
}

func selected(keys ...string) map[string]any {
	result := map[string]any{}
	for _, key := range keys {
		selectedValue := 1
		if key == "wan" {
			selectedValue = 0
		}
		result[key] = map[string]any{"value": key, "selected": selectedValue}
	}
	return result
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func readGeneralLookup(t *testing.T, baseURL, id string) GeneralResourceModel {
	t.Helper()
	state, diags := readGeneralLookupRaw(t, baseURL, id)
	if diags.HasError() {
		t.Fatalf("diagnostics = %s", diagText(diags))
	}
	return state
}

func readGeneralLookupRaw(t *testing.T, baseURL, id string) (GeneralResourceModel, diag.Diagnostics) {
	t.Helper()
	ds := newGeneralDataSource()
	configureDataSource(t, ds, testClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := GeneralResourceModel{
		ID:                types.StringValue(id),
		ActiveInterface:   types.SetNull(types.StringType),
		OutgoingInterface: types.SetNull(types.StringType),
		PrivateDomains:    types.SetNull(types.StringType),
		PrivateAddresses:  types.SetNull(types.StringType),
		InsecureDomains:   types.SetNull(types.StringType),
	}
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &config)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		return GeneralResourceModel{}, resp.Diagnostics
	}
	var state GeneralResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state, resp.Diagnostics
}
