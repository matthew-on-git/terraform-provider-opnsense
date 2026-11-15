// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestProtocolGeneralLookupReadOnly(t *testing.T) {
	ctx := context.Background()
	var calls []string
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls = append(calls, req.Method+" "+req.URL.Path)
		if req.Method != http.MethodGet {
			t.Errorf("lookup mutated appliance: %s %s", req.Method, req.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		switch req.URL.Path {
		case "/api/core/menu/search":
			_, _ = io.WriteString(w, `{}`)
		case "/api/quagga/general/get":
			_, _ = io.WriteString(w, `{"general":{"enabled":"0","profile":{"traditional":{"selected":1},"datacenter":{"selected":0}},"enablecarp":"0","enablesyslog":"0","enablesnmp":"1","sysloglevel":"warnings","fwrules":"0","manual_config":"1","daemons":{"bgp":{"selected":1}},"ignoredExtra":"not-lookup-state"}}`)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	server, schemas := bgpGlobalLookupServer(t, api.URL)
	schema := schemas.DataSourceSchemas["opnsense_quagga_general"]
	if schema == nil {
		t.Fatal("FRR general lookup is not registered")
	}
	for _, name := range []string{"password", "secret", "key", "manual_config", "daemons", "ignoredExtra"} {
		if bgpGlobalSchemaHas(schema, name) {
			t.Fatalf("excluded field %s is lookup state", name)
		}
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_quagga_general",
		Config:   bgpGlobalLookupValue(t, schema, map[string]any{"id": "general"}),
	})
	if err != nil || bgpGlobalLookupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := bgpGlobalLookupAttrs(t, schema, read.State)
	var got string
	var enabled bool
	if err := attrs["id"].As(&got); err != nil || got != "general" {
		t.Fatalf("id = %q %v", got, err)
	}
	if err := attrs["enabled"].As(&enabled); err != nil || enabled {
		t.Fatalf("enabled = %v %v", enabled, err)
	}
	if err := attrs["profile"].As(&got); err != nil || got != "traditional" {
		t.Fatalf("profile = %q %v", got, err)
	}
	if err := attrs["syslog_level"].As(&got); err != nil || got != "warnings" {
		t.Fatalf("syslog_level = %q %v", got, err)
	}
	var syslog bool
	if err := attrs["enable_syslog"].As(&syslog); err != nil || syslog {
		t.Fatalf("disabled syslog was synthesized: %v %v", syslog, err)
	}
	if _, ok := attrs["manual_config"]; ok {
		t.Fatal("manual_config is lookup state")
	}
	if _, ok := attrs["daemons"]; ok {
		t.Fatal("daemons is lookup state")
	}
	before := len(calls)
	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_quagga_general",
		Config:   bgpGlobalLookupValue(t, schema, map[string]any{"id": "not-general"}),
	})
	if err != nil || !bgpGlobalLookupErrors(rejected.Diagnostics) || len(calls) != before {
		t.Fatalf("wrong id err=%v diagnostics=%v calls=%d/%d", err, rejected.Diagnostics, len(calls), before)
	}
	for _, call := range calls {
		if !strings.HasPrefix(call, "GET ") {
			t.Fatalf("mutation call %s", call)
		}
	}
}
