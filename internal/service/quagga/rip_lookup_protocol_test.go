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
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProtocolRIPLookupReadOnly(t *testing.T) {
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
		case "/api/quagga/rip/get":
			_, _ = io.WriteString(w, `{"rip":{"enabled":"0","version":"2","networks":{"":{"value":"","selected":1}},"redistribute":{"connected":{"selected":1},"static":{"selected":0}},"defaultmetric":"","passiveinterfaces":{"wan":{"selected":1}},"ignoredExtra":"not-lookup-state"}}`)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	server, schemas := bgpGlobalLookupServer(t, api.URL)
	schema := schemas.DataSourceSchemas["opnsense_quagga_rip"]
	if schema == nil {
		t.Fatal("RIP lookup is not registered")
	}
	for _, name := range []string{"password", "secret", "key", "passive_interfaces", "passiveinterfaces", "ignoredExtra"} {
		if bgpGlobalSchemaHas(schema, name) {
			t.Fatalf("excluded field %s is lookup state", name)
		}
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_quagga_rip",
		Config:   bgpGlobalLookupValue(t, schema, map[string]any{"id": "rip"}),
	})
	if err != nil || bgpGlobalLookupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := bgpGlobalLookupAttrs(t, schema, read.State)
	var got string
	var enabled bool
	if err := attrs["id"].As(&got); err != nil || got != "rip" {
		t.Fatalf("id = %q %v", got, err)
	}
	if err := attrs["enabled"].As(&enabled); err != nil || enabled {
		t.Fatalf("enabled = %v %v", enabled, err)
	}
	var version big.Float
	if err := attrs["version"].As(&version); err != nil || version.Cmp(big.NewFloat(2)) != 0 {
		t.Fatalf("version = %v %v", version.String(), err)
	}
	if !attrs["default_metric"].IsNull() {
		t.Fatal("blank metric was synthesized")
	}
	var networks []tftypes.Value
	if err := attrs["networks"].As(&networks); err != nil || len(networks) != 0 {
		t.Fatalf("empty networks = %v %v", networks, err)
	}
	var redistribute []tftypes.Value
	if err := attrs["redistribute"].As(&redistribute); err != nil || len(redistribute) != 1 {
		t.Fatalf("redistribute = %v %v", redistribute, err)
	}
	if _, ok := attrs["passiveinterfaces"]; ok {
		t.Fatal("passiveinterfaces is lookup state")
	}
	before := len(calls)
	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_quagga_rip",
		Config:   bgpGlobalLookupValue(t, schema, map[string]any{"id": "ospf"}),
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
