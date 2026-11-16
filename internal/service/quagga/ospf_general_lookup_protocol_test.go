// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"context"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestProtocolOSPFGeneralLookupReadOnly(t *testing.T) {
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
		case "/api/quagga/ospfsettings/get":
			_, _ = io.WriteString(w, `{"ospf":{"enabled":"0","routerid":"","costreference":"","logadjacencychanges":"1","originate":"0","originatealways":"1","originatemetric":"0","carp_demote":"0","areas":{"area":{}},"ignoredExtra":"not-lookup-state"}}`)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	server, schemas := bgpGlobalLookupServer(t, api.URL)
	schema := schemas.DataSourceSchemas["opnsense_quagga_ospf_general"]
	if schema == nil {
		t.Fatal("OSPF general lookup is not registered")
	}
	for _, name := range []string{"password", "secret", "key", "passive_interfaces", "areas", "networks", "neighbors", "interfaces", "prefixlists", "routemaps", "redistributions", "ignoredExtra"} {
		if bgpGlobalSchemaHas(schema, name) {
			t.Fatalf("excluded field %s is lookup state", name)
		}
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_quagga_ospf_general",
		Config:   bgpGlobalLookupValue(t, schema, map[string]any{"id": "ospf"}),
	})
	if err != nil || bgpGlobalLookupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := bgpGlobalLookupAttrs(t, schema, read.State)
	var got string
	var enabled bool
	if err := attrs["id"].As(&got); err != nil || got != "ospf" {
		t.Fatalf("id = %q %v", got, err)
	}
	if err := attrs["enabled"].As(&enabled); err != nil || enabled {
		t.Fatalf("enabled = %v %v", enabled, err)
	}
	if !attrs["router_id"].IsNull() || !attrs["cost_reference"].IsNull() {
		t.Fatal("blank router ID or cost reference was synthesized")
	}
	var always bool
	var logged bool
	if err := attrs["originate_default_always"].As(&always); err != nil || !always {
		t.Fatalf("originate_default_always = %v %v", always, err)
	}
	if err := attrs["log_adjacency_changes"].As(&logged); err != nil || !logged {
		t.Fatalf("log_adjacency_changes = %v %v", logged, err)
	}
	var metric big.Float
	if err := attrs["originate_default_metric"].As(&metric); err != nil || metric.Cmp(big.NewFloat(0)) != 0 {
		t.Fatalf("literal metric 0 changed: %v %v", metric.String(), err)
	}
	if _, ok := attrs["areas"]; ok {
		t.Fatal("areas is lookup state")
	}
	if _, ok := attrs["passive_interfaces"]; ok {
		t.Fatal("passive_interfaces is lookup state")
	}
	before := len(calls)
	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_quagga_ospf_general",
		Config:   bgpGlobalLookupValue(t, schema, map[string]any{"id": "general"}),
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
