// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// This receipt is operator-owned/private. It is never emitted in test output.
// The identity endpoint must be independently chosen and fingerprinted for the
// newly reserved guest, not copied from whatever target the credentials reach.
type ipsecIsolationReceipt struct {
	TargetURI                string    `json:"target_uri"`
	GuestID                  string    `json:"guest_id"`
	ReservationID            string    `json:"reservation_id"`
	ReservedBy               string    `json:"reserved_by"`
	Purpose                  string    `json:"purpose"`
	CreatedAt                time.Time `json:"created_at"`
	ExpiresAt                time.Time `json:"expires_at"`
	Isolated                 bool      `json:"isolated"`
	NewlyCreated             bool      `json:"newly_created"`
	Disposable               bool      `json:"disposable"`
	NoProductionConnectivity bool      `json:"no_production_connectivity"`
	AllowServiceReconfigure  bool      `json:"allow_service_reconfigure"`
	IdentityEndpoint         string    `json:"identity_endpoint"`
	IdentitySHA256           string    `json:"identity_sha256"`
}

func validateIPsecIsolation(receipt ipsecIsolationReceipt, uri string, now time.Time) error {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || strings.Trim(parsed.Path, "/") != "" {
		return fmt.Errorf("acceptance requires an exact HTTPS appliance origin")
	}
	if receipt.TargetURI != uri || receipt.GuestID == "" || receipt.ReservationID == "" || receipt.ReservedBy == "" || receipt.Purpose != "ipsec-disposable-acceptance" {
		return fmt.Errorf("receipt lacks exact target, reservation, owner or acceptance purpose")
	}
	if !receipt.Isolated || !receipt.NewlyCreated || !receipt.Disposable || !receipt.NoProductionConnectivity || !receipt.AllowServiceReconfigure {
		return fmt.Errorf("all isolation and service-wide activation attestations are mandatory")
	}
	if receipt.CreatedAt.After(now) || receipt.CreatedAt.Before(now.Add(-24*time.Hour)) || receipt.ExpiresAt.Before(now.Add(ipsecExecutionBudget+ipsecCleanupReserve)) || receipt.ExpiresAt.After(now.Add(2*time.Hour)) {
		return fmt.Errorf("reservation must be newly created and have an active bounded lease with at least 45 minutes execution plus 15 minutes cleanup")
	}
	identityURL, err := url.Parse(receipt.IdentityEndpoint)
	if err != nil || identityURL.IsAbs() || identityURL.Host != "" || identityURL.RawQuery != "" || identityURL.Fragment != "" || strings.Contains(receipt.IdentityEndpoint, "..") || (!strings.HasPrefix(receipt.IdentityEndpoint, "/api/core/") && !strings.HasPrefix(receipt.IdentityEndpoint, "/api/diagnostics/")) {
		return fmt.Errorf("identity proof must use a read-only core/diagnostics API path")
	}
	hash, err := hex.DecodeString(receipt.IdentitySHA256)
	if err != nil || len(hash) != sha256.Size {
		return fmt.Errorf("independent identity SHA256 is required")
	}
	return nil
}

func ipsecAcceptanceRead(ctx context.Context, client *opnsense.Client, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.BaseURL()+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid acceptance read")
	}
	response, err := client.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("acceptance read failed")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("acceptance read returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("acceptance read incomplete")
	}
	return body, nil
}

func ipsecAcceptancePreflight(t *testing.T) (*opnsense.Client, *ipsecLeaseGuard) {
	t.Helper()
	if os.Getenv("TF_ACC") != "1" || os.Getenv("OPNSENSE_IPSEC_DISPOSABLE_ACK") != "I_RESERVED_A_NEW_ISOLATED_GUEST" {
		t.Fatal("explicit disposable acceptance opt-in is required")
	}
	acctest.PreCheck(t)
	receiptPath := os.Getenv("OPNSENSE_IPSEC_ISOLATION_RECEIPT")
	if receiptPath == "" {
		t.Fatal("private isolation receipt is required; no target fallback is permitted")
	}
	receiptFile, err := os.OpenInRoot(filepath.Dir(receiptPath), filepath.Base(receiptPath))
	if err != nil {
		t.Fatal("cannot open isolation receipt")
	}
	defer func() { _ = receiptFile.Close() }()
	raw, err := io.ReadAll(io.LimitReader(receiptFile, 65536))
	if err != nil {
		t.Fatal("cannot read isolation receipt")
	}
	var receipt ipsecIsolationReceipt
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&receipt) != nil {
		t.Fatal("invalid isolation receipt")
	}
	if err = validateIPsecIsolation(receipt, os.Getenv("OPNSENSE_URI"), time.Now()); err != nil {
		t.Fatal(err)
	}
	client := acctest.TestClient(t)
	identity, err := ipsecAcceptanceRead(context.Background(), client, receipt.IdentityEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(identity)
	if hex.EncodeToString(hash[:]) != strings.ToLower(receipt.IdentitySHA256) {
		t.Fatal("target identity does not match the independently reserved guest")
	}
	firmware, err := ipsecAcceptanceRead(context.Background(), client, "/api/core/firmware/status")
	if err != nil {
		t.Fatal(err)
	}
	if err = validateIPsecAcceptanceFirmware(firmware); err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{"connection", "child"} {
		body, readErr := ipsecAcceptanceRead(context.Background(), client, "/api/ipsec/connections/search_"+owner)
		if readErr != nil {
			t.Fatal(readErr)
		}
		var inventory struct {
			Total json.Number       `json:"total"`
			Rows  []json.RawMessage `json:"rows"`
		}
		if json.Unmarshal(body, &inventory) != nil || inventory.Total.String() != "0" || len(inventory.Rows) != 0 {
			t.Fatal("reserved guest must have an empty connection/child inventory before acceptance")
		}
	}
	guard := &ipsecLeaseGuard{receipt: receipt, now: time.Now, owned: map[string]string{}}
	if err = guard.initializeOwnershipJournal(os.Getenv("OPNSENSE_IPSEC_OWNERSHIP_JOURNAL")); err != nil {
		t.Fatal(err)
	}
	installIPsecGuard(client, guard)
	return client, guard
}

// Exact package attestation is independent of the base source-equivalence pin.
func validateIPsecAcceptanceFirmware(raw []byte) error {
	var version struct {
		Product struct {
			Name    string `json:"product_name"`
			Version string `json:"product_version"`
		} `json:"product"`
	}
	if json.Unmarshal(raw, &version) != nil || version.Product.Name != "OPNsense" || version.Product.Version != "26.7.8_2" {
		return fmt.Errorf("acceptance requires the attested released OPNsense 26.7.8_2 package")
	}
	return nil
}

func ipsecAbsent(client *opnsense.Client, owner, id string) error {
	if id == "" {
		return fmt.Errorf("cannot prove deletion without a recorded owned UUID")
	}
	raw, err := ipsecAcceptanceRead(context.Background(), client, "/api/ipsec/connections/get_"+owner+"/"+id)
	if err != nil {
		return err
	}
	if ipsecRawAbsent(raw, owner) {
		return nil
	}
	return fmt.Errorf("owned %s deletion is not independently proven", owner)
}

func TestAccIPsecAdvancedDisposable(t *testing.T) {
	enabled, err := ipsecDedicatedAcceptanceEnabled(os.Getenv("TF_ACC"), os.Getenv("OPNSENSE_IPSEC_DISPOSABLE_ACK"))
	if err != nil {
		t.Fatal(err)
	}
	if !enabled {
		t.Skip("dedicated live acceptance unexecuted: separate explicit acknowledgment and private isolation receipt required")
	}
	client, guard := ipsecAcceptancePreflight(t)
	// Runs even if Core refresh-before-destroy fails on a pending activation.
	// A cleanup error adds a failure; it never replaces/clears Terraform's failure.
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if cleanupErr := cleanupIPsecOwned(ctx, client, guard); cleanupErr != nil {
			t.Errorf("independent owned cleanup failed; retain reservation and ownership records: %s", cleanupErr)
		}
	})
	var connectionID, childID string
	var beforeOmission map[string]map[string]string
	var transferJournal []byte
	var transferJournalTime time.Time
	capture := func(state *terraform.State) error {
		if item := state.RootModule().Resources["opnsense_ipsec_connection.advanced"]; item != nil {
			connectionID = item.Primary.ID
		}
		if item := state.RootModule().Resources["opnsense_ipsec_child.advanced"]; item != nil {
			childID = item.Primary.ID
		}
		if connectionID == "" || childID == "" {
			return fmt.Errorf("owned identities missing")
		}
		if err := verifyIPsecAcceptanceLookups(state, false); err != nil {
			return err
		}
		return verifyIPsecAcceptanceFields(client, connectionID, childID, false)
	}
	check := func(cleared bool) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			if err := verifyIPsecAcceptanceLookups(state, cleared); err != nil {
				return err
			}
			return verifyIPsecAcceptanceFields(client, connectionID, childID, cleared)
		}
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"opnsense": providerserver.NewProtocol6WithError(&guardedIPsecProvider{Provider: provider.New("isolated-acceptance")(), guard: guard})},
		CheckDestroy: func(_ *terraform.State) error {
			if err := ipsecAbsent(client, "child", childID); err != nil {
				return err
			}
			return ipsecAbsent(client, "connection", connectionID)
		},
		Steps: []resource.TestStep{
			{Config: strings.ReplaceAll(strings.ReplaceAll(ipsecAcceptanceConfig(true, false), "\"advanced\"", "\"seed\""), ".advanced.id", ".seed.id"), Check: func(state *terraform.State) error {
				connectionID = state.RootModule().Resources["opnsense_ipsec_connection.seed"].Primary.ID
				childID = state.RootModule().Resources["opnsense_ipsec_child.seed"].Primary.ID
				t.Setenv("TF_VAR_ipsec_connection_import_id", connectionID)
				t.Setenv("TF_VAR_ipsec_child_import_id", childID)
				if err := verifyIPsecAcceptanceLookupsAt(state, "seed", false); err != nil {
					return err
				}
				return verifyIPsecAcceptanceFields(client, connectionID, childID, false)
			}},
			// Import both owners into fresh addresses in one Core apply. Sequential
			// ImportState steps normally use disposable state, not the next apply's state.
			{PreConfig: func() {
				var err error
				transferJournal, err = os.ReadFile(guard.journalPath)
				if err != nil {
					t.Fatal("cannot read ownership journal before import")
				}
				info, err := os.Stat(guard.journalPath)
				if err != nil {
					t.Fatal("cannot stat ownership journal before import")
				}
				transferJournalTime = info.ModTime()
			}, Config: ipsecAcceptanceImportConfig(), Check: func(state *terraform.State) error {
				if state.RootModule().Resources["opnsense_ipsec_connection.advanced"].Primary.ID != connectionID || state.RootModule().Resources["opnsense_ipsec_child.advanced"].Primary.ID != childID {
					return fmt.Errorf("import transfer changed owned UUIDs")
				}
				if state.RootModule().Resources["opnsense_ipsec_connection.seed"] != nil || state.RootModule().Resources["opnsense_ipsec_child.seed"] != nil {
					return fmt.Errorf("import transfer retained seed owners")
				}
				journal, err := os.ReadFile(guard.journalPath)
				if err != nil || !bytes.Equal(transferJournal, journal) {
					return fmt.Errorf("import changed ownership journal")
				}
				info, err := os.Stat(guard.journalPath)
				if err != nil || !info.ModTime().Equal(transferJournalTime) {
					return fmt.Errorf("import dispatched a journaled mutation")
				}
				return capture(state)
			}},
			{Config: ipsecAcceptanceConfig(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{PreConfig: func() {
				var readErr error
				beforeOmission, readErr = ipsecSavedAdvancedSnapshot(client, connectionID, childID)
				if readErr != nil {
					t.Fatal(readErr)
				}
			}, Config: ipsecAcceptanceOmittedConfig(), Check: func(state *terraform.State) error {
				after, readErr := ipsecSavedAdvancedSnapshot(client, connectionID, childID)
				if readErr != nil {
					return readErr
				}
				for _, owner := range []string{"connection", "child"} {
					if !maps.Equal(beforeOmission[owner], after[owner]) {
						return fmt.Errorf("omitted imported %s policy changed", owner)
					}
				}
				return check(false)(state)
			}},
			{Config: ipsecAcceptanceOmittedConfig(), PlanOnly: true, ExpectNonEmptyPlan: false},
			{PreConfig: func() {
				payload := map[string]string{"dpd_timeout": "121"}
				if err := opnsense.Update(context.Background(), client, opnsense.ReqOpts{UpdateEndpoint: "/api/ipsec/connections/set_connection", Monad: "connection", ReconfigureEndpoint: "/api/ipsec/service/reconfigure"}, &payload, connectionID); err != nil {
					t.Fatal("owned connection drift injection failed")
				}
			}, Config: ipsecAcceptanceConfig(true, false), PlanOnly: true, ExpectNonEmptyPlan: true},
			{Config: ipsecAcceptanceConfig(true, false), Check: check(false)},
			{PreConfig: func() {
				payload := map[string]string{"reqid": "124"}
				if err := opnsense.Update(context.Background(), client, opnsense.ReqOpts{UpdateEndpoint: "/api/ipsec/connections/set_child", Monad: "child", ReconfigureEndpoint: "/api/ipsec/service/reconfigure"}, &payload, childID); err != nil {
					t.Fatal("owned child drift injection failed")
				}
			}, Config: ipsecAcceptanceConfig(true, false), PlanOnly: true, ExpectNonEmptyPlan: true},
			{Config: ipsecAcceptanceConfig(true, false), Check: check(false)},
			{Config: ipsecAcceptanceConfig(true, true), Check: check(true)},
			{Config: ipsecAcceptanceConfig(false, true), Check: func(_ *terraform.State) error { return ipsecAbsent(client, "child", childID) }},
		},
	})
	t.Log("disposable connection/child create, import, no-change, drift/repair, clears, independent child deletion and parent cleanup completed")
}

func ipsecAcceptanceConfig(child, clearValues bool) string {
	timeout, request, rekey := "120", "123", "3600"
	policies, enabledFlags := "true", "true"
	if clearValues {
		timeout = ""
		request = ""
		rekey = "0"
		policies = "false"
		enabledFlags = "false"
	}
	config := fmt.Sprintf(`
resource "opnsense_ipsec_connection" "advanced" {
 enabled = false
 description = "tf-disposable-advanced-ipsec"
 remote_addrs = "203.0.113.1"
 proposals = "default"
 version = "1"
 aggressive = %s
 dpd_delay = "0"
 dpd_timeout = %q
 ike_lifetime = "30"
 ip_pools = ""
 keying_tries = "0"
 local_addresses = "192.0.2.1"
 local_port = ""
 mobike = false
 reauthentication_time = ""
 rekey_time = "3600"
 remote_port = "4500"
 send_certificate = "never"
 send_certificate_request = false
 udp_encapsulation = true
}
data "opnsense_ipsec_connection" "advanced" { id = opnsense_ipsec_connection.advanced.id }
`, enabledFlags, timeout)
	if child {
		config += fmt.Sprintf(`
resource "opnsense_ipsec_child" "advanced" {
 enabled = false
 connection_id = opnsense_ipsec_connection.advanced.id
 description = "tf-disposable-advanced-child"
 local_ts = "192.0.2.0/24"
 remote_ts = "198.51.100.0/24"
 esp_proposals = "default"
 start_action = "trap|start"
 close_action = "none"
 dpd_action = "clear"
 install_policies = %s
 request_id = %q
 sha256_96 = %s
 rekey_time = %q
}
data "opnsense_ipsec_child" "advanced" { id = opnsense_ipsec_child.advanced.id }
`, policies, request, enabledFlags, rekey)
	}
	return config
}

func TestIPsecDisposableIsolationGuard(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	good := ipsecIsolationReceipt{TargetURI: "https://firewall.example.test", GuestID: "reserved-guest", ReservationID: "reservation-receipt", ReservedBy: "acceptance-operator", Purpose: "ipsec-disposable-acceptance", CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), Isolated: true, NewlyCreated: true, Disposable: true, NoProductionConnectivity: true, AllowServiceReconfigure: true, IdentityEndpoint: "/api/diagnostics/system/systemInformation", IdentitySHA256: strings.Repeat("a", 64)}
	if err := validateIPsecIsolation(good, good.TargetURI, now); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*ipsecIsolationReceipt){"wrong target": func(r *ipsecIsolationReceipt) { r.TargetURI = "https://other.example.test" }, "shared target": func(r *ipsecIsolationReceipt) { r.Isolated = false }, "existing guest": func(r *ipsecIsolationReceipt) { r.NewlyCreated = false }, "production connectivity": func(r *ipsecIsolationReceipt) { r.NoProductionConnectivity = false }, "no activation consent": func(r *ipsecIsolationReceipt) { r.AllowServiceReconfigure = false }, "expired lease": func(r *ipsecIsolationReceipt) { r.ExpiresAt = now }, "short lease": func(r *ipsecIsolationReceipt) { r.ExpiresAt = now.Add(59 * time.Minute) }, "unbounded lease": func(r *ipsecIsolationReceipt) { r.ExpiresAt = now.Add(24 * time.Hour) }, "missing reservation": func(r *ipsecIsolationReceipt) { r.ReservationID = "" }, "missing identity": func(r *ipsecIsolationReceipt) { r.IdentitySHA256 = "" }, "mutation identity path": func(r *ipsecIsolationReceipt) { r.IdentityEndpoint = "/api/ipsec/connections/add_connection" }}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			bad := good
			change(&bad)
			if validateIPsecIsolation(bad, good.TargetURI, now) == nil {
				t.Fatal("unsafe attestation accepted")
			}
		})
	}
}

// Independent raw saved-field observations, not provider model round trips.
func expectedIPsecAcceptanceFields(cleared bool) map[string]map[string]string {
	expected := map[string]map[string]string{
		"connection": {"enabled": "0", "aggressive": "1", "dpd_delay": "0", "dpd_timeout": "120", "over_time": "30", "pools": "", "keyingtries": "0", "local_addrs": "192.0.2.1", "local_port": "", "mobike": "0", "reauth_time": "", "rekey_time": "3600", "remote_port": "4500", "send_cert": "never", "send_certreq": "0", "encap": "1"},
		"child":      {"enabled": "0", "close_action": "none", "dpd_action": "clear", "policies": "1", "reqid": "123", "sha256_96": "1", "rekey_time": "3600"},
	}
	if cleared {
		expected["connection"]["aggressive"] = "0"
		expected["child"]["sha256_96"] = "0"
		expected["connection"]["dpd_timeout"] = ""
		expected["child"]["policies"] = "0"
		expected["child"]["reqid"] = ""
		expected["child"]["rekey_time"] = "0"
	}
	return expected
}

func verifyIPsecAcceptanceFields(client *opnsense.Client, connectionID, childID string, cleared bool) error {
	expected := expectedIPsecAcceptanceFields(cleared)
	for _, owner := range []string{"connection", "child"} {
		id := connectionID
		if owner == "child" {
			id = childID
		}
		body, err := ipsecAcceptanceRead(context.Background(), client, "/api/ipsec/connections/get_"+owner+"/"+id)
		if err != nil {
			return err
		}
		var wrapped map[string]map[string]json.RawMessage
		if json.Unmarshal(body, &wrapped) != nil || wrapped[owner] == nil {
			return fmt.Errorf("independent saved %s read is malformed", owner)
		}
		for key, want := range expected[owner] {
			raw, exists := wrapped[owner][key]
			if !exists {
				return fmt.Errorf("independent %s read omits %s", owner, key)
			}
			value, readErr := ipsecAcceptanceScalar(raw)
			if readErr != nil || value != want {
				return fmt.Errorf("independent saved %s.%s differs from expected value", owner, key)
			}
		}
	}
	return nil
}

func ipsecAcceptanceScalar(raw []byte) (string, error) {
	var scalar string
	if json.Unmarshal(raw, &scalar) == nil && string(raw) != "null" {
		return scalar, nil
	}
	if strings.TrimSpace(string(raw)) == "[]" {
		return "", nil
	}
	var options map[string]struct {
		Selected json.Number `json:"selected"`
	}
	if json.Unmarshal(raw, &options) != nil || options == nil {
		return "", fmt.Errorf("invalid saved field shape")
	}
	values := []string{}
	for key, option := range options {
		switch option.Selected.String() {
		case "1":
			values = append(values, key)
		case "0":
		default:
			return "", fmt.Errorf("invalid selection")
		}
	}
	sort.Strings(values)
	return strings.Join(values, ","), nil
}

func ipsecAcceptanceOmittedConfig() string {
	return `resource "opnsense_ipsec_connection" "advanced" {
 enabled=false
 version="1"
 description="tf-disposable-advanced-ipsec-omission"
 remote_addrs="203.0.113.1"
 proposals="default"
}
resource "opnsense_ipsec_child" "advanced" {
 enabled=false
 connection_id=opnsense_ipsec_connection.advanced.id
 description="tf-disposable-advanced-child-omission"
 local_ts="192.0.2.0/24"
 remote_ts="198.51.100.0/24"
 esp_proposals="default"
 start_action="trap|start"
}
data "opnsense_ipsec_connection" "advanced" { id=opnsense_ipsec_connection.advanced.id }
data "opnsense_ipsec_child" "advanced" { id=opnsense_ipsec_child.advanced.id }
`
}

func ipsecSavedAdvancedSnapshot(client *opnsense.Client, connectionID, childID string) (map[string]map[string]string, error) {
	snapshot := map[string]map[string]string{}
	for _, owner := range []string{"connection", "child"} {
		id := connectionID
		if owner == "child" {
			id = childID
		}
		body, err := ipsecAcceptanceRead(context.Background(), client, "/api/ipsec/connections/get_"+owner+"/"+id)
		if err != nil {
			return nil, err
		}
		var wrapped map[string]map[string]json.RawMessage
		if json.Unmarshal(body, &wrapped) != nil || wrapped[owner] == nil {
			return nil, fmt.Errorf("invalid independent snapshot")
		}
		fields := []string{"aggressive", "dpd_delay", "dpd_timeout", "over_time", "pools", "keyingtries", "local_addrs", "local_port", "mobike", "reauth_time", "rekey_time", "remote_port", "send_cert", "send_certreq", "encap"}
		if owner == "child" {
			fields = []string{"close_action", "dpd_action", "policies", "reqid", "sha256_96", "rekey_time"}
		}
		snapshot[owner] = map[string]string{}
		for _, field := range fields {
			raw, present := wrapped[owner][field]
			if !present {
				return nil, fmt.Errorf("snapshot omits %s.%s", owner, field)
			}
			value, decodeErr := ipsecAcceptanceScalar(raw)
			if decodeErr != nil {
				return nil, decodeErr
			}
			snapshot[owner][field] = value
		}
	}
	return snapshot, nil
}

func ipsecAcceptanceImportConfig() string {
	return ipsecAcceptanceConfig(true, false) + `
variable "ipsec_connection_import_id" { type = string }
variable "ipsec_child_import_id" { type = string }
removed {
 from = opnsense_ipsec_connection.seed
 lifecycle { destroy = false }
}
removed {
 from = opnsense_ipsec_child.seed
 lifecycle { destroy = false }
}
import {
 to = opnsense_ipsec_connection.advanced
 id = var.ipsec_connection_import_id
}
import {
 to = opnsense_ipsec_child.advanced
 id = var.ipsec_child_import_id
}
`
}

func verifyIPsecAcceptanceLookups(state *terraform.State, cleared bool) error {
	return verifyIPsecAcceptanceLookupsAt(state, "advanced", cleared)
}

func verifyIPsecAcceptanceLookupsAt(state *terraform.State, name string, cleared bool) error {
	// Explicit independent public-name translations, not provider field metadata.
	translated := map[string]string{"over_time": "ike_lifetime", "pools": "ip_pools", "keyingtries": "keying_tries", "local_addrs": "local_addresses", "reauth_time": "reauthentication_time", "send_cert": "send_certificate", "send_certreq": "send_certificate_request", "encap": "udp_encapsulation", "policies": "install_policies", "reqid": "request_id"}
	flags := map[string]bool{"aggressive": true, "mobike": true, "send_certreq": true, "encap": true, "policies": true, "sha256_96": true}
	for owner, expected := range expectedIPsecAcceptanceFields(cleared) {
		item := state.RootModule().Resources["data.opnsense_ipsec_"+owner+"."+name]
		if item == nil || item.Primary == nil {
			return fmt.Errorf("%s lookup state missing", owner)
		}
		for wire, want := range expected {
			if wire == "enabled" {
				continue
			}
			attribute := wire
			if public, ok := translated[wire]; ok {
				attribute = public
			}
			if flags[wire] {
				if want == "1" {
					want = "true"
				} else {
					want = "false"
				}
			}
			got, present := item.Primary.Attributes[attribute]
			if !present || got != want {
				return fmt.Errorf("%s lookup %s differs from independent expected value", owner, attribute)
			}
		}
	}
	return nil
}
