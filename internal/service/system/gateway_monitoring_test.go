// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGatewayMonitoringSavedRoundTrip(t *testing.T) {
	for _, protocol := range []string{"inet", "inet6"} {
		t.Run(protocol, func(t *testing.T) {
			raw := `{"gateway":"","defaultgw":"0","ipprotocol":"` + protocol + `","monitor":"192.0.2.53","fargw":"1","monitor_noroute":"1","monitor_killstates":"1","monitor_killstates_priority":"1","force_down":"0","nosync":"1"}`
			var api gatewayAPIResponse
			if err := json.Unmarshal([]byte(raw), &api); err != nil {
				t.Fatal(err)
			}
			var model GatewayResourceModel
			model.fromAPI(context.Background(), &api, "fixture")
			request := model.toAPI(context.Background())
			payload, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			var values map[string]string
			if err := json.Unmarshal(payload, &values); err != nil {
				t.Fatal(err)
			}
			for key, want := range map[string]string{"gateway": "", "defaultgw": "0", "monitor": "192.0.2.53", "fargw": "1", "monitor_noroute": "1", "monitor_killstates": "1", "monitor_killstates_priority": "1", "force_down": "0", "nosync": "1"} {
				if got, ok := values[key]; !ok || got != want {
					t.Errorf("%s: got %q present=%v want %q", key, got, ok, want)
				}
			}
		})
	}
}

func TestGatewayMonitoringOmissionAndExplicitClear(t *testing.T) {
	for _, unknown := range []bool{false, true} {
		var model GatewayResourceModel
		if unknown {
			model.Monitor = types.StringUnknown()
			model.ForceDown = types.BoolUnknown()
		}
		encoded, err := json.Marshal(model.toAPI(context.Background()))
		if err != nil {
			t.Fatal(err)
		}
		var values map[string]string
		if err := json.Unmarshal(encoded, &values); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"monitor", "force_down", "fargw", "monitor_noroute", "monitor_killstates", "monitor_killstates_priority", "nosync"} {
			if _, ok := values[key]; ok {
				t.Errorf("omitted intent wrote %s", key)
			}
		}
	}
	var model GatewayResourceModel
	model.fromAPI(context.Background(), &gatewayAPIResponse{}, "fixture")
	if !model.Monitor.IsNull() || !model.ForceDown.IsNull() {
		t.Fatal("absent response fields must remain null")
	}
	model.Monitor = types.StringValue("")
	model.ForceDown = types.BoolValue(false)
	request := model.toAPI(context.Background())
	if request.Monitor == nil || *request.Monitor != "" || request.ForceDown == nil || *request.ForceDown != "0" {
		t.Fatal("explicit clear or false dropped")
	}
}

func TestGatewayMonitoringBooleanResponseContract(t *testing.T) {
	for _, tc := range []struct {
		name, payload        string
		null, value, invalid bool
	}{
		{name: "missing", payload: `{}`, null: true},
		{name: "null", payload: `{"force_down":null}`, null: true},
		{name: "empty", payload: `{"force_down":""}`},
		{name: "zero", payload: `{"force_down":"0"}`},
		{name: "one", payload: `{"force_down":"1"}`, value: true},
		{name: "malformed", payload: `{"force_down":"yes"}`, invalid: true},
		{name: "number", payload: `{"force_down":1}`, invalid: true},
		{name: "bool", payload: `{"force_down":true}`, invalid: true},
		{name: "object", payload: `{"force_down":{}}`, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var api gatewayAPIResponse
			err := json.Unmarshal([]byte(tc.payload), &api)
			if tc.invalid {
				if err == nil {
					t.Fatal("malformed boolean accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var model GatewayResourceModel
			model.fromAPI(context.Background(), &api, "fixture")
			if model.ForceDown.IsNull() != tc.null || model.ForceDown.ValueBool() != tc.value {
				t.Fatalf("unexpected boolean state: %v", model.ForceDown)
			}
			if tc.null && model.toAPI(context.Background()).ForceDown != nil {
				t.Fatal("missing/null boolean became a write")
			}
		})
	}
}

func TestGatewayMonitoringNumericResponseAndClear(t *testing.T) {
	for _, tc := range []struct {
		raw     string
		invalid bool
	}{
		{`{"latencylow":"","latencyhigh":"500","losslow":"1","losshigh":"100","interval":"1","time_period":"60","loss_interval":"4","data_length":"0"}`, false},
		{`{"latencylow":null,"data_length":null}`, false},
		{`{"interval":"0"}`, true},
		{`{"losslow":"100"}`, true},
		{`{"losshigh":"101"}`, true},
		{`{"data_length":"-1"}`, true},
		{`{"data_length":"1.5"}`, true},
		{`{"data_length":"01"}`, true},
		{`{"data_length":1}`, true},
	} {
		var api gatewayAPIResponse
		err := json.Unmarshal([]byte(tc.raw), &api)
		if (err != nil) != tc.invalid {
			t.Errorf("response %s error=%v", tc.raw, err)
		}
	}
	var api gatewayAPIResponse
	if err := json.Unmarshal([]byte(`{"latencylow":"","data_length":"0"}`), &api); err != nil {
		t.Fatal(err)
	}
	var model GatewayResourceModel
	model.fromAPI(context.Background(), &api, "fixture")
	if model.LatencyLow.IsNull() || model.LatencyLow.ValueString() != "" || model.DataLength.ValueString() != "0" || !model.Interval.IsNull() {
		t.Fatal("default/zero/missing distinctions lost")
	}
	model.DataLength = types.StringValue("")
	request := model.toAPI(context.Background())
	if request.DataLength == nil || *request.DataLength != "" {
		t.Fatal("numeric clear omitted")
	}
}
