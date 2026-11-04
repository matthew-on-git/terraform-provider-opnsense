// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const settingsRecoveryKey = "trust_settings_recovery_v1"

type settingsRecovery struct {
	ID      string          `json:"id"`
	Kind    string          `json:"kind"`
	Payload settingsPayload `json:"payload"`
}

func settingsMark(ctx context.Context, p settingsPrivate, kind string, payload settingsPayload) diag.Diagnostics {
	var d diag.Diagnostics
	if p == nil || reflect.ValueOf(p).IsNil() {
		d.AddError("Missing durable Trust settings recovery state", "Private state is required before a Trust settings mutation.")
		return d
	}
	if kind == "" {
		return p.SetKey(ctx, settingsRecoveryKey, nil)
	}
	raw, err := json.Marshal(settingsRecovery{ID: "trust_settings", Kind: kind, Payload: payload})
	if err != nil {
		d.AddError("Invalid Trust settings recovery state", err.Error())
		return d
	}
	return p.SetKey(ctx, settingsRecoveryKey, raw)
}

func settingsPending(ctx context.Context, p settingsPrivate, id string) (settingsRecovery, diag.Diagnostics) {
	var r settingsRecovery
	var d diag.Diagnostics
	if id != "trust_settings" {
		d.AddError("Invalid Trust settings identity", "State identity must be exactly trust_settings.")
		return r, d
	}
	if p == nil || reflect.ValueOf(p).IsNil() {
		return r, d
	}
	raw, diags := p.GetKey(ctx, settingsRecoveryKey)
	d.Append(diags...)
	if len(raw) == 0 {
		return r, d
	}
	err := json.Unmarshal(raw, &r)
	canonical, canonicalErr := json.Marshal(r)
	if err != nil || canonicalErr != nil || !bytes.Equal(raw, canonical) || r.ID != "trust_settings" || (r.Kind != settingsSaved && r.Kind != settingsSaveUnknown) || len(r.Payload) == 0 || validateSettingsPayload(r.Payload, false) != nil {
		d.AddError("Invalid Trust settings recovery state", "Noncanonical, corrupt, unsupported, identity-mismatched or invalid requested payload; no mutation attempted.")
	}
	return r, d
}

func requireSettings(c *opnsense.Client) error {
	if c == nil {
		return errors.New("trust settings client is not configured")
	}
	return nil
}

func readSettings(ctx context.Context, c *opnsense.Client) (SettingsResourceModel, error) {
	if err := c.AcquireRead(ctx); err != nil {
		return SettingsResourceModel{}, err
	}
	defer c.ReleaseRead()
	raw, err := settingsRequest(ctx, c, http.MethodGet, "/api/trust/settings/get", nil)
	if err != nil {
		return SettingsResourceModel{}, err
	}
	return decodeSettings(raw)
}

// Recovery always observes first and compares only the requested subset.
func recoverSettings(ctx context.Context, c *opnsense.Client, pending settingsRecovery, set func(SettingsResourceModel) diag.Diagnostics, p settingsPrivate) diag.Diagnostics {
	var d diag.Diagnostics
	remote, err := readSettings(ctx, c)
	if err != nil {
		d.AddError("Trust settings recovery read failed", err.Error())
		return d
	}
	d.Append(set(remote)...)
	if d.HasError() {
		return d
	}
	if pending.Kind != "" && settingsMatches(remote, pending.Payload) {
		if err := activateSettings(ctx, c); err != nil {
			d.AddError("Saved Trust settings require activation", err.Error())
			return d
		}
	} else if pending.Kind != "" {
		d.AddWarning("Trust settings save reconciled", "Authoritative fields differ from the requested save. No activation or repeated set; review a fresh plan.")
	}
	d.Append(settingsMark(context.WithoutCancel(ctx), p, "", nil)...)
	return d
}

func saveSettings(ctx context.Context, c *opnsense.Client, config, plan SettingsResourceModel, pending settingsRecovery, set func(SettingsResourceModel) diag.Diagnostics, p settingsPrivate) diag.Diagnostics {
	var d diag.Diagnostics
	payload, err := config.wireIntent()
	if err != nil {
		d.AddError("Invalid Trust settings", err.Error())
		return d
	}
	if c == nil {
		d.AddError("Unconfigured Trust settings client", "The provider client is unavailable.")
		return d
	}
	if err := c.LockMutex(ctx); err != nil {
		d.AddError("Trust settings mutation canceled", err.Error())
		return d
	}
	defer c.UnlockMutex()
	if err := requireSettings(c); err != nil {
		d.AddError("Trust settings unavailable", err.Error())
		return d
	}
	if pending.Kind != "" {
		d.Append(recoverSettings(ctx, c, pending, set, p)...)
		if !d.HasError() {
			d.AddError("Trust settings recovered; fresh plan required", "Recovery did not repeat the saved mutation. Run a fresh plan and apply.")
		}
		return d
	}
	before, err := readSettings(ctx, c)
	if err != nil {
		d.AddError("Cannot verify Trust settings baseline", err.Error())
		return d
	}
	if field := settingsStaleOmission(config, plan, before); field != "" {
		d.AddError("Stale Trust settings plan", fmt.Sprintf("Omitted planned field %s changed on the appliance. No settings saved; refresh and create a fresh plan.", field))
		return d
	}
	if len(payload) == 0 {
		d.Append(set(before)...)
		return d
	}
	// Prove durable private storage works before issuing a non-idempotent write.
	d.Append(settingsMark(context.WithoutCancel(ctx), p, settingsSaveUnknown, payload)...)
	if d.HasError() {
		return d
	}
	merged := before.values()
	for k, v := range payload {
		merged[k] = v
	}
	raw, err := settingsRequest(ctx, c, http.MethodPost, "/api/trust/settings/set", map[string]any{"trust": payload})
	if err == nil {
		err = parseSettingsSave(raw)
	}
	persist := context.WithoutCancel(ctx)
	if err != nil {
		var validation *opnsense.ValidationError
		if errors.As(err, &validation) {
			d.Append(settingsMark(persist, p, "", nil)...)
			d.AddError("Trust settings validation failed", err.Error())
			return d
		}
		d.Append(set(before)...)
		d.Append(settingsMark(persist, p, settingsSaveUnknown, payload)...)
		d.AddError("Trust settings save outcome uncertain", fmt.Sprintf("%v; fixed identity and requested fields retained. Refresh to reconcile without repeating set.", err))
		return d
	}
	d.Append(set(settingsModel(merged))...)
	d.Append(settingsMark(persist, p, settingsSaved, payload)...)
	if d.HasError() {
		return d
	}
	var observed SettingsResourceModel
	d.Append(recoverSettings(ctx, c, settingsRecovery{ID: "trust_settings", Kind: settingsSaved, Payload: payload}, func(remote SettingsResourceModel) diag.Diagnostics {
		observed = remote
		return set(remote)
	}, p)...)
	if !d.HasError() && !settingsMatches(observed, payload) {
		d.AddError("Saved Trust settings require reconciliation", "Authoritative requested fields differ; pending intent was cleared without activation. Review a fresh plan.")
	}
	return d
}

// Only omitted, known planned values constrain fresh baseline adoption.
func settingsStaleOmission(config, plan, current SettingsResourceModel) string {
	configs := []types.Bool{config.StoreIntermediateCerts, config.InstallCRLs, config.FetchCRLs, config.EnableLegacySect, config.EnableConfigConstraints}
	plans := []types.Bool{plan.StoreIntermediateCerts, plan.InstallCRLs, plan.FetchCRLs, plan.EnableLegacySect, plan.EnableConfigConstraints}
	actual := []types.Bool{current.StoreIntermediateCerts, current.InstallCRLs, current.FetchCRLs, current.EnableLegacySect, current.EnableConfigConstraints}
	for i, v := range configs {
		if v.IsNull() && !plans[i].IsNull() && !plans[i].IsUnknown() && !plans[i].Equal(actual[i]) {
			return settingsBools[i]
		}
	}
	if config.CipherString.IsNull() && !plan.CipherString.IsNull() && !plan.CipherString.IsUnknown() && !plan.CipherString.Equal(current.CipherString) {
		return "cipher_string"
	}
	return ""
}

func parseSettingsSave(raw []byte) error {
	fields, err := settingsObject(raw, []string{"result", "validations", "status"})
	if err != nil {
		return err
	}
	var result *string
	if json.Unmarshal(fields["result"], &result) != nil || result == nil {
		return errors.New("ambiguous settings save result")
	}
	if statusRaw, ok := fields["status"]; ok {
		var status *string
		if json.Unmarshal(statusRaw, &status) != nil || status == nil || (*result == "saved" && *status != "ok") || (*result == "failed" && *status != "failed") {
			return errors.New("contradictory settings save status")
		}
	}
	if *result == "saved" {
		if v, ok := fields["validations"]; ok && string(v) != "null" && string(v) != "{}" {
			return errors.New("saved settings response contains validations")
		}
	}
	_, err = opnsense.ParseMutationResponse(raw)
	return err
}
