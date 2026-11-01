// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Emulate DNat's actual asymmetry: writes consume names from category; reads
// expose UUIDs through categories. A write to categories is silently ignored.
func TestNatPortForwardCategoriesLifecycle(t *testing.T) {
	ctx := context.Background()
	var selected []string
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/firewall/category/getItem/cat-a":
			_, _ = io.WriteString(w, `{"category":{"name":"DNS"}}`)
		case "/api/firewall/category/getItem/cat-b":
			_, _ = io.WriteString(w, `{"category":{"name":"dNAT"}}`)
		case "/api/firewall/d_nat/add_rule", "/api/firewall/d_nat/set_rule/id":
			var body struct {
				Rule map[string]any `json:"rule"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			writes++
			if _, present := body.Rule["categories"]; present {
				t.Error("wrote volatile categories field")
			}
			if value, present := body.Rule["category"]; present {
				selected = nil
				if value != "" {
					for _, name := range strings.Split(value.(string), ",") {
						switch name {
						case "DNS":
							selected = append(selected, "cat-a")
						case "dNAT":
							selected = append(selected, "cat-b")
						default:
							t.Errorf("expected category name, got %q", name)
						}
					}
				}
			}
			_, _ = io.WriteString(w, `{"result":"saved","uuid":"id"}`)
		case "/api/firewall/d_nat/apply":
			_, _ = io.WriteString(w, `{"status":"ok"}`)
		case "/api/firewall/d_nat/get_rule/id":
			rule := map[string]any{
				"disabled": "1", "sequence": "10", "interface": map[string]any{"lan": map[string]int{"selected": 1}, "wan": map[string]int{"selected": 1}},
				"ipprotocol": map[string]any{"inet": map[string]int{"selected": 1}}, "protocol": map[string]any{"tcp": map[string]int{"selected": 1}},
				"source": map[string]string{"network": "any", "port": "", "not": "0"}, "destination": map[string]string{"network": "any", "port": "443", "not": "0"},
				"target": "192.0.2.1", "local-port": "443", "log": "0", "descr": "test", "natreflection": []string{}, "categories": strings.Join(selected, ","),
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"rule": rule})
		default:
			t.Errorf("unexpected endpoint %s", req.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	r := &natPortForwardResource{}
	configureResource(t, r, server.URL)
	model := normalizeNAT(NatPortForwardResourceModel{Enabled: types.BoolValue(false), Sequence: types.Int64Value(10), Interface: types.StringValue("wan"), Interfaces: stringSetValue([]string{"lan", "wan"}), IPProtocol: types.StringValue("inet"), Protocol: types.StringValue("tcp"), SourceNet: types.StringValue("any"), SourcePort: types.StringValue(""), SourceNot: types.BoolValue(false), DestinationNet: types.StringValue("any"), DestinationPort: types.StringValue("443"), DestinationNot: types.BoolValue(false), Target: types.StringValue("192.0.2.1"), LocalPort: types.StringValue("443"), Log: types.BoolValue(false), Description: types.StringValue("test"), Reflection: types.StringValue(""), Categories: stringSetValue([]string{"cat-b", "cat-a"})})
	plan := resourceState(t, r, model)
	created := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan(plan)}, &created)
	if created.Diagnostics.HasError() {
		t.Fatal(created.Diagnostics)
	}
	check := func(state tfsdk.State, want types.Set) {
		t.Helper()
		var got NatPortForwardResourceModel
		if d := state.Get(ctx, &got); d.HasError() {
			t.Fatal(d)
		}
		if !got.Categories.Equal(want) {
			t.Fatalf("categories = %s, want %s", got.Categories, want)
		}
		if !got.Interfaces.Equal(model.Interfaces) || got.Target.ValueString() != "192.0.2.1" {
			t.Fatal("unrelated rule fields changed")
		}
	}
	check(created.State, model.Categories)
	prior := created.State
	for _, step := range []struct {
		name        string
		value, want types.Set
	}{
		{"preserve omitted", types.SetUnknown(types.StringType), model.Categories},
		{"replace", stringSetValue([]string{"cat-b"}), stringSetValue([]string{"cat-b"})},
		{"clear", stringSetValue(nil), stringSetValue(nil)},
	} {
		t.Run(step.name, func(t *testing.T) {
			model.Categories = step.value
			plan = resourceState(t, r, model)
			updated := resource.UpdateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.Update(ctx, resource.UpdateRequest{Plan: tfsdk.Plan(plan), State: prior}, &updated)
			if updated.Diagnostics.HasError() {
				t.Fatal(updated.Diagnostics)
			}
			check(updated.State, step.want)
			read := resource.ReadResponse{State: updated.State}
			r.Read(ctx, resource.ReadRequest{State: updated.State}, &read)
			if read.Diagnostics.HasError() {
				t.Fatal(read.Diagnostics)
			}
			check(read.State, step.want)
			prior = read.State
		})
	}
	if writes != 4 {
		t.Fatalf("writes = %d, want 4", writes)
	}
}

func TestNatPortForwardCategoriesFailBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		status         int
	}{
		{"missing", `{}`, 404},
		{"permission denied", `{}`, 403},
		{"empty name", `{"category":{"name":""}}`, 200},
		{"invalid name", `{"category":{"name":"one,two"}}`, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet {
					writes++
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.response)
			}))
			defer server.Close()
			r := &natPortForwardResource{}
			configureResource(t, r, server.URL)
			model := normalizeNAT(NatPortForwardResourceModel{ID: types.StringValue("id"), Categories: stringSetValue([]string{"missing"})})
			state := resourceState(t, r, model)
			create := resource.CreateResponse{State: tfsdk.State{Schema: state.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(state)}, &create)
			if !create.Diagnostics.HasError() {
				t.Fatal("expected create error")
			}
			update := resource.UpdateResponse{State: tfsdk.State{Schema: state.Schema}}
			r.Update(context.Background(), resource.UpdateRequest{Plan: tfsdk.Plan(state), State: state}, &update)
			if !update.Diagnostics.HasError() {
				t.Fatal("expected update error")
			}
			if writes != 0 {
				t.Fatalf("%d writes after category lookup failure", writes)
			}
		})
	}
}
