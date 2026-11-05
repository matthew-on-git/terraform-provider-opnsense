// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestUnsetGeneralIntegers(t *testing.T) {
	for _, value := range []string{"", "0", "42"} {
		api := generalAPIResponse{IncomingNumTCP: value}
		var model GeneralResourceModel
		model.fromAPI(context.Background(), &api, generalID)
		normalizeUnsetGeneralIntegers(&model, &api)
		if model.IncomingNumTCP.IsNull() != (value == "") {
			t.Fatalf("wrong null state for %q", value)
		}
		fields := map[string]json.RawMessage{"incomingnumtcp": json.RawMessage(`"` + value + `"`)}
		if err := validateUnboundBranchFields("fixture", fields); err != nil {
			t.Fatal(err)
		}
		payload := generalSettingsPayload(context.Background(), model)["advanced"].(map[string]any)
		_, exists := payload["incomingnumtcp"]
		if exists != (value != "") {
			t.Fatalf("unset value must be omitted: %#v", payload)
		}
	}
	for _, value := range []string{"-1", "garbage"} {
		if err := validateUnboundBranchFields("fixture", map[string]json.RawMessage{"incomingnumtcp": json.RawMessage(`"` + value + `"`)}); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}

func TestUnsetGeneralCapturedResponse(t *testing.T) {
	data, err := os.ReadFile("testdata/unset_integers.json")
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Unbound unboundSettingsEnvelope `json:"unbound"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	api, err := flattenGeneralSettings(&response.Unbound)
	if err != nil {
		t.Fatal(err)
	}
	var state GeneralResourceModel
	state.fromAPI(context.Background(), api, generalID)
	normalizeUnsetGeneralIntegers(&state, api)
	if !state.IncomingNumTCP.IsNull() || !state.OutgoingRange.IsNull() {
		t.Fatal("appliance defaults were lost")
	}
}
