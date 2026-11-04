// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SettingsResourceModel owns only the six audited Trust general policies.
type SettingsResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	StoreIntermediateCerts  types.Bool   `tfsdk:"store_intermediate_certs"`
	InstallCRLs             types.Bool   `tfsdk:"install_crls"`
	FetchCRLs               types.Bool   `tfsdk:"fetch_crls"`
	EnableLegacySect        types.Bool   `tfsdk:"enable_legacy_sect"`
	EnableConfigConstraints types.Bool   `tfsdk:"enable_config_constraints"`
	CipherString            types.Set    `tfsdk:"cipher_string"`
}
type settingsPayload map[string]string

var settingsBools = []string{"store_intermediate_certs", "install_crls", "fetch_crls", "enable_legacy_sect", "enable_config_constraints"}

func validateSettingsPayload(p settingsPayload, complete bool) error {
	for k, v := range p {
		if k == "CipherString" {
			if values, err := cipherCSV(v); err != nil || strings.Join(values, ",") != v {
				return fmt.Errorf("noncanonical CipherString payload")
			}
			continue
		}
		found := false
		for _, name := range settingsBools {
			if k == name {
				found = true
			}
		}
		if !found || (v != "0" && v != "1") {
			return fmt.Errorf("invalid Trust settings field %q", k)
		}
	}
	if complete && len(p) != 6 {
		return fmt.Errorf("incomplete Trust settings")
	}
	return nil
}

func cipherCSV(s string) ([]string, error) {
	if s == "" {
		return []string{}, nil
	}
	values := strings.Split(s, ",")
	seen := map[string]bool{}
	for _, v := range values {
		if strings.TrimSpace(v) != v || v == "" || seen[v] {
			return nil, fmt.Errorf("invalid or duplicate CipherString choice")
		}
		seen[v] = true
	}
	sort.Strings(values)
	return values, nil
}

func (m SettingsResourceModel) wireIntent() (settingsPayload, error) {
	p := settingsPayload{}
	for k, v := range map[string]types.Bool{"store_intermediate_certs": m.StoreIntermediateCerts, "install_crls": m.InstallCRLs, "fetch_crls": m.FetchCRLs, "enable_legacy_sect": m.EnableLegacySect, "enable_config_constraints": m.EnableConfigConstraints} {
		if v.IsUnknown() {
			return nil, fmt.Errorf("%s must be known", k)
		}
		if !v.IsNull() {
			p[k] = "0"
			if v.ValueBool() {
				p[k] = "1"
			}
		}
	}
	if m.CipherString.IsUnknown() {
		return nil, fmt.Errorf("cipher_string must be known")
	}
	if !m.CipherString.IsNull() {
		var values []string
		d := m.CipherString.ElementsAs(context.Background(), &values, false)
		if d.HasError() {
			return nil, fmt.Errorf("cipher_string members must be known strings")
		}
		sort.Strings(values)
		p["CipherString"] = strings.Join(values, ",")
		if len(values) > 0 && p["CipherString"] == "" {
			return nil, fmt.Errorf("empty cipher choice")
		}
		for _, v := range values {
			if v == "" || strings.Contains(v, ",") {
				return nil, fmt.Errorf("invalid cipher choice")
			}
		}
	}
	return p, validateSettingsPayload(p, false)
}

func settingsModel(p settingsPayload) SettingsResourceModel {
	values, _ := cipherCSV(p["CipherString"])
	set, _ := types.SetValueFrom(context.Background(), types.StringType, values)
	return SettingsResourceModel{ID: types.StringValue("trust_settings"), StoreIntermediateCerts: types.BoolValue(p["store_intermediate_certs"] == "1"), InstallCRLs: types.BoolValue(p["install_crls"] == "1"), FetchCRLs: types.BoolValue(p["fetch_crls"] == "1"), EnableLegacySect: types.BoolValue(p["enable_legacy_sect"] == "1"), EnableConfigConstraints: types.BoolValue(p["enable_config_constraints"] == "1"), CipherString: set}
}
func (m SettingsResourceModel) values() settingsPayload { p, _ := m.wireIntent(); return p }
func settingsMatches(m SettingsResourceModel, p settingsPayload) bool {
	actual := m.values()
	for k, v := range p {
		if actual[k] != v {
			return false
		}
	}
	return true
}

func decodeSettings(raw []byte) (SettingsResourceModel, error) {
	envelope, err := settingsObject(raw, []string{"trust"})
	if err != nil {
		return SettingsResourceModel{}, err
	}
	owned := append(append([]string{}, settingsBools...), "CipherString")
	fields, err := settingsObject(envelope["trust"], owned)
	if err != nil {
		return SettingsResourceModel{}, err
	}
	p := settingsPayload{}
	for _, k := range settingsBools {
		var v *string
		if json.Unmarshal(fields[k], &v) != nil || v == nil {
			return SettingsResourceModel{}, fmt.Errorf("missing or malformed %s", k)
		}
		p[k] = *v
	}
	raw = fields["CipherString"]
	var text *string
	if json.Unmarshal(raw, &text) == nil && text != nil {
		values, e := cipherCSV(*text)
		if e != nil {
			return SettingsResourceModel{}, e
		}
		p["CipherString"] = strings.Join(values, ",")
	} else {
		choices, e := settingsObject(raw, nil)
		if e != nil {
			return SettingsResourceModel{}, e
		}
		values := []string{}
		for key, value := range choices {
			if key == "" || strings.Contains(key, ",") {
				return SettingsResourceModel{}, fmt.Errorf("invalid cipher key")
			}
			entry, e := settingsObject(value, []string{"selected", "value"})
			if e != nil {
				return SettingsResourceModel{}, e
			}
			var selected *int
			if json.Unmarshal(entry["selected"], &selected) != nil || (selected == nil || (*selected != 0 && *selected != 1)) {
				return SettingsResourceModel{}, fmt.Errorf("invalid cipher selector")
			}
			var label *string
			if json.Unmarshal(entry["value"], &label) != nil || label == nil {
				return SettingsResourceModel{}, fmt.Errorf("invalid cipher label")
			}
			if *selected == 1 {
				values = append(values, key)
			}
		}
		sort.Strings(values)
		p["CipherString"] = strings.Join(values, ",")
	}
	if err := validateSettingsPayload(p, true); err != nil {
		return SettingsResourceModel{}, err
	}
	return settingsModel(p), nil
}

func settingsObject(raw []byte, owned []string) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("expected Trust settings JSON object")
	}
	result := map[string]json.RawMessage{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("invalid Trust settings object key")
		}
		for _, name := range append(owned, key) {
			if strings.EqualFold(key, name) {
				if key != name {
					return nil, fmt.Errorf("case-alias Trust settings key %q", key)
				}
				if _, exists := result[key]; exists {
					return nil, fmt.Errorf("duplicate Trust settings key %q", key)
				}
			}
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, fmt.Errorf("trailing Trust settings JSON data")
	}
	return result, nil
}
