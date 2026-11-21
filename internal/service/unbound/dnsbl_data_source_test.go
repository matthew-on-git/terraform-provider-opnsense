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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDNSBLDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newDNSBLDataSource())
	if len(schema.Attributes) != len(dnsblLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(dnsblLookupFields))
	}
	for _, name := range dnsblLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlist field %s is not in the schema", name)
		}
	}
	for _, excluded := range []string{"source_nets", "cache_ttl", "description", "uuid"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is in the schema", excluded)
		}
	}
}

func TestDNSBLDataSource_readSoleItem(t *testing.T) {
	t.Parallel()

	const id = "00000000-0000-4000-8000-000000000001"
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == dnsblReqOpts.SearchEndpoint:
			_, _ = w.Write([]byte(`{"rows":[{"uuid":"` + id + `"}],"rowCount":1,"total":1,"current":1}`))
		case r.Method == http.MethodGet && r.URL.Path == dnsblReqOpts.GetEndpoint+"/"+id:
			_, _ = w.Write([]byte(`{"blocklist":` + dnsblFixtureItem(t, map[string]any{
				"enabled":     " 1 ",
				"address":     " 192.0.2.31 ",
				"nxdomain":    "0",
				"note":        "ignored-extra",
				"cache_ttl":   "72000",
				"description": "fixture",
			}) + `}`))
		case r.Method == http.MethodGet && r.URL.Path == dnsblSettingsReqOpts.GetEndpoint:
			_, _ = w.Write([]byte(`{"unbound":{"general":{"safesearch":" 1 "},"dnsbl":{"blocklist":{"decoy":{"address":"198.51.100.9"}}}}}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	state := readDNSBLLookup(t, server.URL, dnsblID)
	if !state.Enabled.ValueBool() || !state.SafeSearch.ValueBool() || state.NXDomain.ValueBool() {
		t.Fatalf("flags = enabled %v safe %v nx %v", state.Enabled, state.SafeSearch, state.NXDomain)
	}
	if state.Address.ValueString() != "192.0.2.31" {
		t.Fatalf("address = %q", state.Address.ValueString())
	}
	if got := setValues(t, state.Type); len(got) != 1 || got[0] != "ag" {
		t.Fatalf("type = %#v", got)
	}
	if got := setValues(t, state.Whitelists); len(got) != 0 {
		t.Fatalf("whitelists = %#v", got)
	}
	joined := strings.Join(calls, "\n")
	for _, forbidden := range []string{"set_dnsbl", "add_dnsbl", "del_dnsbl", "reconfigure", "/settings/set"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("lookup mutated via %s in %s", forbidden, joined)
		}
	}
	if !strings.Contains(joined, "POST "+dnsblReqOpts.SearchEndpoint) || !strings.Contains(joined, "GET "+dnsblReqOpts.GetEndpoint+"/"+id) {
		t.Fatalf("calls = %s", joined)
	}
}

func TestDNSBLDataSource_wrongIDDoesNotRead(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(server.Close)

	_, diags := readDNSBLLookupRaw(t, server.URL, "not-the-id")
	if !diags.HasError() || !strings.Contains(diagText(diags), "Invalid DNSBL ID") {
		t.Fatalf("diagnostics = %s", diagText(diags))
	}
}

func TestDNSBLDataSource_zeroItemsNotFound(t *testing.T) {
	t.Parallel()

	server := dnsblSearchServer(t, `{"rows":[],"rowCount":0,"total":0,"current":1}`)
	t.Cleanup(server.Close)
	_, diags := readDNSBLLookupRaw(t, server.URL, dnsblID)
	if diagText(diags) != dnsblNotFoundSummary+"\nThe stable blocklist search returned no items. An empty result is not found and is not replaced with a new item or model defaults." {
		t.Fatalf("diagnostics = %q", diagText(diags))
	}
}

func TestDNSBLDataSource_manyItemsFailClosed(t *testing.T) {
	t.Parallel()

	body := `{"rows":[{"uuid":"00000000-0000-4000-8000-000000000001"},{"uuid":"00000000-0000-4000-8000-000000000002"}],"rowCount":2,"total":2,"current":1}`
	server := dnsblSearchServer(t, body)
	t.Cleanup(server.Close)
	_, diags := readDNSBLLookupRaw(t, server.URL, dnsblID)
	if !strings.Contains(diagText(diags), "Ambiguous DNSBL blocklist ownership") || strings.Contains(diagText(diags), dnsblNotFoundSummary) {
		t.Fatalf("diagnostics = %s", diagText(diags))
	}
}

func TestDNSBLDataSource_searchMissingRowsIsNotNotFound(t *testing.T) {
	t.Parallel()

	server := dnsblSearchServer(t, `{"rowCount":0,"total":0,"current":1}`)
	t.Cleanup(server.Close)
	_, diags := readDNSBLLookupRaw(t, server.URL, dnsblID)
	text := diagText(diags)
	if strings.Contains(text, dnsblNotFoundSummary) || !strings.Contains(text, "missing rows") {
		t.Fatalf("diagnostics = %s", text)
	}
}

func TestDNSBLDataSource_pluginMissingFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != dnsblReqOpts.SearchEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	_, diags := readDNSBLLookupRaw(t, server.URL, dnsblID)
	text := diagText(diags)
	if !strings.Contains(text, "plugin 'unbound' is not installed on OPNsense") || strings.Contains(text, dnsblNotFoundSummary) {
		t.Fatalf("diagnostics = %s", text)
	}
}

func TestDNSBLDataSource_itemShapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		item    map[string]any
		general string
		wantErr string
		check   func(*testing.T, dnsblLookupModel)
	}{
		{
			name:    "blank enabled",
			item:    map[string]any{"enabled": " "},
			general: `"0"`,
			wantErr: "enabled is missing, blank, or not 0/1",
		},
		{
			name:    "disabled",
			item:    map[string]any{"enabled": "0"},
			general: `"0"`,
			check: func(t *testing.T, state dnsblLookupModel) {
				t.Helper()
				if state.Enabled.ValueBool() || state.SafeSearch.ValueBool() {
					t.Fatalf("expected false flags, got %#v %#v", state.Enabled, state.SafeSearch)
				}
			},
		},
		{
			name:    "blank nxdomain",
			item:    map[string]any{"nxdomain": " "},
			general: `null`,
			check: func(t *testing.T, state dnsblLookupModel) {
				t.Helper()
				if !state.NXDomain.IsNull() || !state.SafeSearch.IsNull() {
					t.Fatalf("expected null optional flags, nx=%#v safe=%#v", state.NXDomain, state.SafeSearch)
				}
			},
		},
		{
			name:    "bad nxdomain",
			item:    map[string]any{"nxdomain": "2"},
			general: `"1"`,
			wantErr: "nxdomain token",
		},
		{
			name:    "padded type key",
			item:    map[string]any{"type": map[string]any{" ag": map[string]any{"selected": 1}}},
			general: `"1"`,
			wantErr: "outside the released option list",
		},
		{
			name:    "unselected type",
			item:    map[string]any{"type": map[string]any{"ag": map[string]any{"selected": 0}}},
			general: `"1"`,
			check: func(t *testing.T, state dnsblLookupModel) {
				t.Helper()
				if got := setValues(t, state.Type); len(got) != 0 {
					t.Fatalf("type = %#v", got)
				}
			},
		},
		{
			name:    "empty address",
			item:    map[string]any{"address": ""},
			general: `"0"`,
			check: func(t *testing.T, state dnsblLookupModel) {
				t.Helper()
				if state.Address.IsNull() || state.Address.ValueString() != "" {
					t.Fatalf("address = %#v", state.Address)
				}
			},
		},
		{
			name:    "masked address",
			item:    map[string]any{"address": "192.0.2.31/32"},
			general: `"0"`,
			wantErr: "dotted IPv4",
		},
		{
			name:    "mapped address",
			item:    map[string]any{"address": "::ffff:192.0.2.1"},
			general: `"0"`,
			wantErr: "dotted IPv4",
		},
		{
			name:    "leading zero address",
			item:    map[string]any{"address": "192.0.2.01"},
			general: `"0"`,
			wantErr: "dotted IPv4",
		},
		{
			name:    "bad safesearch",
			item:    map[string]any{},
			general: `"2"`,
			wantErr: "not 0 or 1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := dnsblItemServer(t, test.item, test.general)
			t.Cleanup(server.Close)
			state, diags := readDNSBLLookupRaw(t, server.URL, dnsblID)
			text := diagText(diags)
			if test.wantErr != "" {
				if !strings.Contains(text, test.wantErr) || strings.Contains(text, dnsblNotFoundSummary) {
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

func TestDNSBLDataSource_paddedEmptyItemIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case dnsblReqOpts.SearchEndpoint:
			_, _ = w.Write([]byte(`{"rows":[{"uuid":"00000000-0000-4000-8000-000000000001"}],"rowCount":1,"total":1,"current":1}`))
		case dnsblReqOpts.GetEndpoint + "/00000000-0000-4000-8000-000000000001":
			_, _ = w.Write([]byte("{\"blocklist\": { }}"))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	_, diags := readDNSBLLookupRaw(t, server.URL, dnsblID)
	text := diagText(diags)
	if strings.Contains(text, dnsblNotFoundSummary) || !strings.Contains(text, "missing ") {
		t.Fatalf("diagnostics = %s", text)
	}
}

func TestDNSBLDataSource_emptySettingsObjectUsesHelperText(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case dnsblReqOpts.SearchEndpoint:
			_, _ = w.Write([]byte(`{"rows":[{"uuid":"00000000-0000-4000-8000-000000000001"}],"rowCount":1,"total":1,"current":1}`))
		case dnsblReqOpts.GetEndpoint + "/00000000-0000-4000-8000-000000000001":
			_, _ = w.Write([]byte(`{"blocklist":` + dnsblFixtureItem(t, nil) + `}`))
		case dnsblSettingsReqOpts.GetEndpoint:
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	_, diags := readDNSBLLookupRaw(t, server.URL, dnsblID)
	text := diagText(diags)
	if !strings.Contains(text, "Error reading Unbound SafeSearch") || !strings.Contains(text, "resource not found") || strings.Contains(text, dnsblNotFoundSummary) {
		t.Fatalf("diagnostics = %s", text)
	}
}

func dnsblFixtureItem(t *testing.T, overrides map[string]any) string {
	t.Helper()
	item := map[string]any{
		"enabled":     "1",
		"type":        map[string]any{"ag": map[string]any{"value": "AdGuard List", "selected": 1}, "el": map[string]any{"value": "EasyList", "selected": 0}},
		"lists":       []any{},
		"allowlists":  []any{},
		"blocklists":  []any{},
		"wildcards":   []any{},
		"source_nets": map[string]any{"192.0.2.0/24": map[string]any{"selected": 1}},
		"address":     "192.0.2.31",
		"nxdomain":    "0",
		"cache_ttl":   "72000",
		"description": "fixture",
		"note":        "ignored-extra",
	}
	for key, value := range overrides {
		item[key] = value
	}
	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func dnsblSearchServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != dnsblReqOpts.SearchEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
}

func dnsblItemServer(t *testing.T, item map[string]any, safeSearch string) *httptest.Server {
	t.Helper()
	const id = "00000000-0000-4000-8000-000000000001"
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case dnsblReqOpts.SearchEndpoint:
			_, _ = w.Write([]byte(`{"rows":[{"uuid":"` + id + `"}],"rowCount":1,"total":1,"current":1}`))
		case dnsblReqOpts.GetEndpoint + "/" + id:
			_, _ = w.Write([]byte(`{"blocklist":` + dnsblFixtureItem(t, item) + `}`))
		case dnsblSettingsReqOpts.GetEndpoint:
			_, _ = w.Write([]byte(`{"unbound":{"general":{"safesearch":` + safeSearch + `}}}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
}

func readDNSBLLookup(t *testing.T, baseURL, id string) dnsblLookupModel {
	t.Helper()
	state, diags := readDNSBLLookupRaw(t, baseURL, id)
	if diags.HasError() {
		t.Fatalf("diagnostics = %s", diagText(diags))
	}
	return state
}

func readDNSBLLookupRaw(t *testing.T, baseURL, id string) (dnsblLookupModel, diag.Diagnostics) {
	t.Helper()
	ds := newDNSBLDataSource()
	configureDataSource(t, ds, testClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := dnsblLookupModel{
		ID:         types.StringValue(id),
		Type:       types.SetNull(types.StringType),
		Lists:      types.SetNull(types.StringType),
		Whitelists: types.SetNull(types.StringType),
		Blocklists: types.SetNull(types.StringType),
		Wildcards:  types.SetNull(types.StringType),
	}
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &config)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		return dnsblLookupModel{}, resp.Diagnostics
	}
	var state dnsblLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state, resp.Diagnostics
}

func diagText(diags diag.Diagnostics) string {
	parts := make([]string, 0, len(diags))
	for _, item := range diags.Errors() {
		parts = append(parts, item.Summary()+"\n"+item.Detail())
	}
	return strings.Join(parts, "\n")
}

func setValues(t *testing.T, value types.Set) []string {
	t.Helper()
	var got []string
	if diags := value.ElementsAs(context.Background(), &got, false); diags.HasError() {
		t.Fatalf("set diagnostics: %v", diags)
	}
	return got
}
