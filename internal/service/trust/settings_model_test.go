// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSettingsStrictMaterialFieldsAndSelections(t *testing.T) {
	base := map[string]any{"store_intermediate_certs": "1", "install_crls": "0", "fetch_crls": "1", "enable_legacy_sect": "0", "enable_config_constraints": "1", "CipherString": map[string]any{"AES256-SHA": map[string]any{"value": "AES256-SHA", "selected": 1}, "AES128-SHA": map[string]any{"value": "AES128-SHA", "selected": 0}}}
	for _, field := range append(append([]string{}, settingsBools...), "CipherString") {
		for _, bad := range []any{nil, true, 1, []any{}, map[string]any{"selected": 1}} {
			t.Run(field, func(t *testing.T) {
				fields := map[string]any{}
				for k, v := range base {
					fields[k] = v
				}
				fields[field] = bad
				raw, _ := json.Marshal(map[string]any{"trust": fields})
				if _, e := decodeSettings(raw); e == nil {
					t.Fatal("malformed field accepted", field, bad)
				}
			})
		}
	}
	for _, bad := range []any{nil, "1", true, 2, -1} {
		fields := map[string]any{}
		for k, v := range base {
			fields[k] = v
		}
		fields["CipherString"] = map[string]any{"AES256-SHA": map[string]any{"value": "AES256-SHA", "selected": bad}}
		raw, _ := json.Marshal(map[string]any{"trust": fields})
		if _, e := decodeSettings(raw); e == nil {
			t.Fatal("malformed selector accepted", bad)
		}
	}
	raw, _ := json.Marshal(map[string]any{"trust": base})
	m, e := decodeSettings(raw)
	if e != nil {
		t.Fatal(e)
	}
	if m.values()["CipherString"] != "AES256-SHA" {
		t.Fatal("unselected choice included")
	}
	for _, cipher := range []any{"", map[string]any{}, "AES256-SHA,AES128-SHA"} {
		base["CipherString"] = cipher
		raw, _ := json.Marshal(map[string]any{"trust": base})
		if _, e := decodeSettings(raw); e != nil {
			t.Fatal("valid choices rejected", e)
		}
	}
	for _, raw := range []string{`{}`, `{"trust":null}`, `{"trust":{}}`, `{"trust":{},"trust":{}}`} {
		if _, e := decodeSettings([]byte(raw)); e == nil {
			t.Fatal("bad envelope accepted")
		}
	}
	for _, raw := range []string{`{"x":1,"x":2}`, `{"selected":0,"selected":1}`} {
		if _, e := settingsObject([]byte(raw), nil); e == nil {
			t.Fatal("duplicate accepted")
		}
	}
}

func TestSettingsUnknownMembersAndBooleanIntent(t *testing.T) {
	ctx := context.Background()
	set, _ := types.SetValue(types.StringType, []attr.Value{types.StringUnknown()})
	m := SettingsResourceModel{CipherString: set}
	if _, e := m.wireIntent(); e == nil {
		t.Fatal("unknown cipher member accepted")
	}
	m.CipherString = types.SetUnknown(types.StringType)
	if _, e := m.wireIntent(); e == nil {
		t.Fatal("unknown set accepted")
	}
	m = SettingsResourceModel{InstallCRLs: types.BoolUnknown()}
	if _, e := m.wireIntent(); e == nil {
		t.Fatal("unknown boolean accepted")
	}
	empty, _ := types.SetValueFrom(ctx, types.StringType, []string{})
	m = SettingsResourceModel{InstallCRLs: types.BoolValue(false), CipherString: empty}
	p, e := m.wireIntent()
	if e != nil || len(p) != 2 || p["install_crls"] != "0" || p["CipherString"] != "" {
		t.Fatal("explicit clear lost", e, p)
	}
}
