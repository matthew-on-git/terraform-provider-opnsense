// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settingsPrivateTest struct {
	raw          []byte
	failNonempty bool
}

func (p *settingsPrivateTest) GetKey(context.Context, string) ([]byte, diag.Diagnostics) {
	return p.raw, nil
}

func (p *settingsPrivateTest) SetKey(ctx context.Context, _ string, raw []byte) diag.Diagnostics {
	var d diag.Diagnostics
	if ctx.Err() != nil {
		d.AddError("Canceled private storage", ctx.Err().Error())
		return d
	}
	if p.failNonempty && len(raw) > 0 {
		d.AddError("Injected private storage failure", "Nonempty intent cannot be persisted.")
		return d
	}
	p.raw = raw
	return nil
}

func TestSettingsPreflightStorageFailureStopsBeforeSet(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/trust/settings/get" {
			t.Error("mutation despite failed intent storage", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"trust":{"store_intermediate_certs":"0","install_crls":"1","fetch_crls":"0","enable_legacy_sect":"1","enable_config_constraints":"0","CipherString":""}}`))
	}))
	defer server.Close()
	p := &settingsPrivateTest{failNonempty: true}
	config := SettingsResourceModel{InstallCRLs: types.BoolValue(false)}
	d := saveSettings(context.Background(), testClient(t, server.URL), config, SettingsResourceModel{}, settingsRecovery{}, func(SettingsResourceModel) diag.Diagnostics { return nil }, p)
	if !d.HasError() || calls != 1 {
		t.Fatal("failed private storage did not stop save", d, calls)
	}
}

func TestSettingsDeleteNilClientCanceledContext(t *testing.T) {
	r := &settingsResource{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var sc resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &sc)
	resp := resource.DeleteResponse{State: tfsdk.State{Schema: sc.Schema}}
	if d := resp.State.Set(context.Background(), settingsModel(settingsPayload{"store_intermediate_certs": "0", "install_crls": "1", "fetch_crls": "0", "enable_legacy_sect": "1", "enable_config_constraints": "0", "CipherString": ""})); d.HasError() {
		t.Fatal(d)
	}
	r.Delete(ctx, resource.DeleteRequest{}, &resp)
	if resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
		t.Fatal("nil-client canceled delete failed", resp)
	}
}

func TestSettingsMalformedPrivateAndIdentityFailClosed(t *testing.T) {
	valid, _ := json.Marshal(settingsRecovery{ID: "trust_settings", Kind: settingsSaved, Payload: settingsPayload{"install_crls": "0"}})
	for _, raw := range [][]byte{[]byte(`{`), []byte(`{"id":"wrong","kind":"saved","payload":{"install_crls":"0"}}`), append([]byte(" "), valid...), []byte(`{"id":"trust_settings","kind":"saved","payload":{"install_crls":"0","install_crls":"1"}}`), []byte(`{"id":"trust_settings","kind":"saved","payload":{}}`)} {
		p := &settingsPrivateTest{raw: raw}
		_, d := settingsPending(context.Background(), p, "trust_settings")
		if !d.HasError() || string(p.raw) != string(raw) {
			t.Fatal("malformed private marker accepted or changed", d)
		}
	}
	p := &settingsPrivateTest{raw: valid}
	_, d := settingsPending(context.Background(), p, "wrong")
	if !d.HasError() || string(p.raw) != string(valid) {
		t.Fatal("corrupt identity accepted or changed", d)
	}
}
