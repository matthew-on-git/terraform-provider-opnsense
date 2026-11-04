// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const (
	ipsecCleanupReserve  = 15 * time.Minute
	ipsecExecutionBudget = 45 * time.Minute
)

func ipsecDedicatedAcceptanceEnabled(tfAcc, ack string) (bool, error) {
	if ack == "" {
		return false, nil
	}
	if tfAcc != "1" || ack != "I_RESERVED_A_NEW_ISOLATED_GUEST" {
		return false, fmt.Errorf("dedicated acceptance requires TF_ACC=1 and the exact disposable acknowledgment")
	}
	return true, nil
}

type ipsecLeaseGuard struct {
	mu                sync.Mutex
	receipt           ipsecIsolationReceipt
	now               func() time.Time
	owned             map[string]string
	deleted           map[string]bool
	cleanupActivation bool
}
type ipsecGuardTransport struct {
	guard *ipsecLeaseGuard
	base  http.RoundTripper
}

func installIPsecGuard(client *opnsense.Client, guard *ipsecLeaseGuard) {
	// Wrap the underlying retry transport, so every retry also checks the lease.
	if retry, ok := client.HTTPClient().Transport.(*retryablehttp.RoundTripper); ok {
		retry.Client.HTTPClient.Transport = &ipsecGuardTransport{guard: guard, base: retry.Client.HTTPClient.Transport}
	} else {
		client.HTTPClient().Transport = &ipsecGuardTransport{guard: guard, base: client.HTTPClient().Transport}
	}
}

type guardedIPsecProvider struct {
	frameworkprovider.Provider
	guard *ipsecLeaseGuard
}

func (p *guardedIPsecProvider) Configure(ctx context.Context, req frameworkprovider.ConfigureRequest, resp *frameworkprovider.ConfigureResponse) {
	p.Provider.Configure(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	if client, ok := resp.ResourceData.(*opnsense.Client); ok {
		installIPsecGuard(client, p.guard)
	}
}

func (t *ipsecGuardTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	g := t.guard
	if req.URL.Scheme+"://"+req.URL.Host != strings.TrimRight(g.receipt.TargetURI, "/") {
		return nil, fmt.Errorf("guard refuses a different appliance origin")
	}
	if req.Method == http.MethodGet || req.Method == http.MethodHead {
		return t.base.RoundTrip(req)
	}
	if req.Method != http.MethodPost {
		return nil, fmt.Errorf("guard refuses unsupported mutation method")
	}
	endpoint := strings.TrimPrefix(req.URL.Path, "/api/ipsec/connections/")
	parts := strings.Split(endpoint, "/")
	action, owner, id := "", "", ""
	if len(parts) > 0 {
		for _, candidate := range []string{"connection", "child"} {
			for _, verb := range []string{"add", "set", "del"} {
				if parts[0] == verb+"_"+candidate {
					action, owner = verb, candidate
				}
			}
		}
	}
	if len(parts) == 2 {
		id = parts[1]
	}
	activation := req.URL.Path == "/api/ipsec/service/reconfigure"
	g.mu.Lock()
	known := id != "" && g.owned[id] == owner
	validAction := (action == "add" && len(parts) == 1) || ((action == "set" || action == "del") && len(parts) == 2 && known) || activation
	leaseActive := g.now().Before(g.receipt.ExpiresAt.Add(-ipsecCleanupReserve))
	cleanupAllowed := (action == "del" && known) || (activation && g.cleanupActivation)
	g.mu.Unlock()
	if !validAction || (!leaseActive && !cleanupAllowed) {
		return nil, fmt.Errorf("lease guard refuses mutation; only owned cleanup remains permitted")
	}
	// Independently pinned identity is rechecked before every mutation, even cleanup.
	identity, err := http.NewRequestWithContext(req.Context(), http.MethodGet, strings.TrimRight(g.receipt.TargetURI, "/")+g.receipt.IdentityEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid guarded identity request")
	}
	response, err := t.base.RoundTrip(identity)
	if err != nil {
		return nil, fmt.Errorf("guarded identity read failed")
	}
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	_ = response.Body.Close()
	hash := sha256.Sum256(raw)
	if readErr != nil || response.StatusCode != http.StatusOK || hex.EncodeToString(hash[:]) != strings.ToLower(g.receipt.IdentitySHA256) {
		return nil, fmt.Errorf("guarded target identity changed")
	}
	// Identity reads can take time; check the lease again at mutation dispatch.
	g.mu.Lock()
	leaseActive = g.now().Before(g.receipt.ExpiresAt.Add(-ipsecCleanupReserve))
	cleanupAllowed = (action == "del" && known) || (activation && g.cleanupActivation)
	g.mu.Unlock()
	if !leaseActive && !cleanupAllowed {
		return nil, fmt.Errorf("execution lease expired during identity verification")
	}
	response, err = t.base.RoundTrip(req)
	if err != nil {
		return response, err
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	_ = response.Body.Close()
	response.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return response, err
	}
	var result struct {
		Result string `json:"result"`
		UUID   string `json:"uuid"`
		Status string `json:"status"`
	}
	if response.StatusCode == http.StatusOK && json.Unmarshal(body, &result) == nil {
		g.mu.Lock()
		defer g.mu.Unlock()
		if action == "add" && result.Result == "saved" && result.UUID != "" {
			g.owned[result.UUID] = owner
		}
		if action == "del" && known && (result.Result == "deleted" || result.Result == "saved") {
			g.cleanupActivation = true
			if g.deleted == nil {
				g.deleted = map[string]bool{}
			}
			g.deleted[id] = true
		}
		if activation && result.Status == "ok" {
			g.cleanupActivation = false
		}
	}
	return response, nil
}

type ipsecRoundTripFunc func(*http.Request) (*http.Response, error)

func (f ipsecRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestIPsecAcceptanceLeaseTransport(t *testing.T) {
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	hash := sha256.Sum256([]byte("identity"))
	calls := 0
	g := &ipsecLeaseGuard{receipt: ipsecIsolationReceipt{TargetURI: "https://firewall.example.test", ExpiresAt: now.Add(time.Hour), IdentityEndpoint: "/api/core/identity/get", IdentitySHA256: hex.EncodeToString(hash[:])}, now: func() time.Time { return now }, owned: map[string]string{ipsecID: "connection"}}
	transport := &ipsecGuardTransport{guard: g, base: ipsecRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := "identity"
		if r.Method == http.MethodPost {
			calls++
			body = `{"result":"saved"}`
			if strings.Contains(r.URL.Path, "del_") {
				body = `{"result":"deleted"}`
			} else if strings.Contains(r.URL.Path, "reconfigure") {
				body = `{"status":"ok"}`
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	})}
	post := func(path string) error {
		r, err := http.NewRequestWithContext(context.Background(), http.MethodPost, g.receipt.TargetURI+path, nil)
		if err != nil {
			return err
		}
		resp, err := transport.RoundTrip(r)
		if resp != nil {
			_ = resp.Body.Close()
		}
		return err
	}
	if err := post("/api/ipsec/connections/set_connection/" + ipsecID); err != nil {
		t.Fatal(err)
	}
	now = now.Add(46 * time.Minute)
	before := calls
	if post("/api/ipsec/connections/set_connection/"+ipsecID) == nil || calls != before {
		t.Fatal("expired execution window allowed mutation")
	}
	if post("/api/ipsec/connections/del_connection/"+otherIPsecID) == nil || calls != before {
		t.Fatal("unowned cleanup allowed")
	}
	if post("/api/ipsec/service/reconfigure") == nil {
		t.Fatal("unrelated activation allowed after execution window")
	}
	if err := post("/api/ipsec/connections/del_connection/" + ipsecID); err != nil {
		t.Fatal("owned cleanup blocked", err)
	}
	if err := post("/api/ipsec/service/reconfigure"); err != nil {
		t.Fatal("owned cleanup activation blocked", err)
	}
	if post("/api/ipsec/service/reconfigure") == nil {
		t.Fatal("consumed cleanup activation replayed")
	}
	g.receipt.IdentitySHA256 = strings.Repeat("0", 64)
	if post("/api/ipsec/connections/del_connection/"+ipsecID) == nil {
		t.Fatal("cleanup allowed after identity changed")
	}
}

func TestIPsecDedicatedAcceptanceOptIn(t *testing.T) {
	if enabled, err := ipsecDedicatedAcceptanceEnabled("1", ""); enabled || err != nil {
		t.Fatal("general TF_ACC workflow must skip dedicated test")
	}
	if enabled, err := ipsecDedicatedAcceptanceEnabled("1", "wrong"); enabled || err == nil {
		t.Fatal("invalid explicit acknowledgment accepted")
	}
	if enabled, err := ipsecDedicatedAcceptanceEnabled("1", "I_RESERVED_A_NEW_ISOLATED_GUEST"); !enabled || err != nil {
		t.Fatal("valid dedicated acknowledgment rejected")
	}
}

// Independent of Terraform refresh/state: snapshot only UUIDs this run recorded
// from affirmative add replies. Never discover or delete arbitrary inventory.
func cleanupIPsecOwned(ctx context.Context, client *opnsense.Client, guard *ipsecLeaseGuard) error {
	guard.mu.Lock()
	owned := make(map[string]string, len(guard.owned))
	for id, owner := range guard.owned {
		owned[id] = owner
	}
	guard.mu.Unlock()
	ids := make([]string, 0, len(owned))
	for id := range owned {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var failures []error
	for _, owner := range []string{"child", "connection"} {
		for _, id := range ids {
			if owned[id] != owner {
				continue
			}
			identity, err := ipsecAcceptanceRead(ctx, client, guard.receipt.IdentityEndpoint)
			hash := sha256.Sum256(identity)
			if err != nil || hex.EncodeToString(hash[:]) != strings.ToLower(guard.receipt.IdentitySHA256) {
				failures = append(failures, fmt.Errorf("owned cleanup identity unproven"))
				continue
			}
			body, err := ipsecAcceptanceRead(ctx, client, "/api/ipsec/connections/get_"+owner+"/"+id)
			if err != nil {
				failures = append(failures, fmt.Errorf("owned %s initial read failed", owner))
				continue
			}
			opts := opnsense.ReqOpts{DeleteEndpoint: "/api/ipsec/connections/del_" + owner, ReconfigureEndpoint: "/api/ipsec/service/reconfigure", StrictDeleteResponse: true, StrictReconfigureResponse: true}
			if ipsecRawAbsent(body, owner) {
				guard.mu.Lock()
				confirmed, pending := guard.deleted[id], guard.cleanupActivation
				guard.mu.Unlock()
				if !confirmed {
					failures = append(failures, fmt.Errorf("owned %s absence lacks confirmed deletion", owner))
					continue
				}
				if pending {
					if err = opnsense.RetryReconfigure(ctx, client, opts); err != nil {
						failures = append(failures, fmt.Errorf("owned cleanup activation failed"))
						continue
					}
				}
			} else if err = opnsense.Delete(ctx, client, opts, id); err != nil {
				failures = append(failures, fmt.Errorf("owned %s delete/activation failed", owner))
				continue
			}
			body, err = ipsecAcceptanceRead(ctx, client, "/api/ipsec/connections/get_"+owner+"/"+id)
			if err != nil || !ipsecRawAbsent(body, owner) {
				failures = append(failures, fmt.Errorf("owned %s exact absence unproven", owner))
				continue
			}
			guard.mu.Lock()
			delete(guard.owned, id)
			delete(guard.deleted, id)
			guard.mu.Unlock()
		}
		// A failed child cleanup must not permit parent cascading over uncertainty.
		if owner == "child" && len(failures) > 0 {
			break
		}
	}
	return errors.Join(failures...)
}

func ipsecRawAbsent(body []byte, owner string) bool {
	var items []json.RawMessage
	if json.Unmarshal(body, &items) == nil && items != nil && len(items) == 0 {
		return true
	}
	var wrapped map[string]json.RawMessage
	return json.Unmarshal(body, &wrapped) == nil && json.Unmarshal(wrapped[owner], &items) == nil && items != nil && len(items) == 0
}

// Core refresh can fail before it reaches Delete. The independent runner must
// still clean only recorded UUIDs using a fresh context after execution expiry.
func TestIPsecExpiredPendingRefreshIndependentCleanup(t *testing.T) {
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	hash := sha256.Sum256([]byte("identity"))
	var mu sync.Mutex
	objects := map[string]map[string]string{otherIPsecID: ipsecFixture("connection")}
	owners := map[string]string{otherIPsecID: "connection"}
	activationFailed := true
	var deletions []string
	activations := 0
	failAbsence := false
	identityBad := false
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path == "/api/core/identity/get" {
			if identityBad {
				_, _ = fmt.Fprint(w, "different")
			} else {
				_, _ = fmt.Fprint(w, "identity")
			}
			return
		}
		if r.URL.Path == "/api/core/firmware/status" {
			_, _ = fmt.Fprint(w, `{"product":{"product_name":"OPNsense","product_version":"26.7.8"}}`)
			return
		}
		if r.URL.Path == "/api/ipsec/service/reconfigure" {
			activations++
			if activationFailed {
				_, _ = fmt.Fprint(w, `{"status":"failed"}`)
			} else {
				_, _ = fmt.Fprint(w, `{"status":"ok"}`)
			}
			return
		}
		endpoint := strings.TrimPrefix(r.URL.Path, "/api/ipsec/connections/")
		parts := strings.Split(endpoint, "/")
		for _, owner := range []string{"connection", "child"} {
			if parts[0] == "add_"+owner {
				id := ipsecID
				if owner == "child" {
					id = authIPsecID
				}
				var body map[string]map[string]string
				if json.NewDecoder(r.Body).Decode(&body) != nil {
					http.Error(w, "bad fixture", 400)
					return
				}
				objects[id] = ipsecDefaults(owner)
				owners[id] = owner
				for key, value := range body[owner] {
					objects[id][key] = value
				}
				_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"`+id+`"}`)
				return
			}
			if len(parts) != 2 {
				continue
			}
			id := parts[1]
			if parts[0] == "get_"+owner {
				if objects[id] == nil {
					if failAbsence && owner == "child" {
						http.Error(w, "fixture unavailable", 500)
						return
					}
					_, _ = fmt.Fprint(w, `[]`)
				} else {
					_ = json.NewEncoder(w).Encode(map[string]any{owner: ipsecWire(owner, objects[id])})
				}
				return
			}
			if parts[0] == "del_"+owner && owners[id] == owner {
				deletions = append(deletions, owner+"/"+id)
				delete(objects, id)
				_, _ = fmt.Fprint(w, `{"result":"deleted"}`)
				return
			}
		}
		_, _ = fmt.Fprint(w, `{}`)
	}))
	defer api.Close()
	guard := &ipsecLeaseGuard{receipt: ipsecIsolationReceipt{TargetURI: api.URL, ExpiresAt: now.Add(time.Hour), IdentityEndpoint: "/api/core/identity/get", IdentitySHA256: hex.EncodeToString(hash[:])}, now: func() time.Time { return now }, owned: map[string]string{}}
	base := providerserver.NewProtocol6(&guardedIPsecProvider{Provider: provider.New("cleanup-fixture")(), guard: guard})()
	schemaResponse, err := base.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	configured, err := base.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: ipsecValue(t, schemaResponse.Provider, map[string]any{"uri": api.URL, "api_key": "fixture-key", "api_secret": "fixture-secret"})})
	if err != nil || ipsecErrors(configured.Diagnostics) {
		t.Fatal("configure failed")
	}
	schema := schemaResponse.ResourceSchemas["opnsense_ipsec_connection"]
	config := ipsecValue(t, schema, ipsecConfig("connection"))
	empty := ipsecNull(t, schema)
	pending := ipsecApply(t, base, "connection", empty, config, ipsecPlan(t, base, "connection", empty, config, config))
	if !ipsecErrors(pending.Diagnostics) || !ipsecPending(pending.Private) {
		t.Fatal("saved create did not leave pending activation")
	}
	client, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: api.URL, APIKey: "fixture-key", APISecret: "fixture-secret", RetryMax: 1})
	if err != nil {
		t.Fatal(err)
	}
	installIPsecGuard(client, guard)
	child := map[string]string{"enabled": "0", "connection": ipsecID}
	childID, addErr := opnsense.Add(context.Background(), client, opnsense.ReqOpts{AddEndpoint: "/api/ipsec/connections/add_child", Monad: "child", ReconfigureEndpoint: "/api/ipsec/service/reconfigure", StrictReconfigureResponse: true}, &child)
	if addErr == nil || childID != authIPsecID {
		t.Fatal("owned child pending save not recorded")
	}
	now = now.Add(46 * time.Minute)
	blockedCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	blocked, err := base.ReadResource(blockedCtx, &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_connection", CurrentState: pending.NewState, Private: pending.Private})
	cancel()
	if err != nil || !ipsecErrors(blocked.Diagnostics) || !ipsecPending(blocked.Private) || len(deletions) != 0 || activations != 2 {
		t.Fatal("expired refresh should fail without activation or deletion")
	}
	mu.Lock()
	activationFailed = false
	failAbsence = true
	mu.Unlock()
	fresh, cancelCleanup := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCleanup()
	if cleanupIPsecOwned(fresh, client, guard) == nil {
		t.Fatal("absence transport failure must fail cleanup")
	}
	guard.mu.Lock()
	remaining := len(guard.owned)
	guard.mu.Unlock()
	if remaining != 2 || len(deletions) != 1 || !strings.HasPrefix(deletions[0], "child/") {
		t.Fatal("uncertain child cleanup lost ownership or cascaded parent")
	}
	mu.Lock()
	failAbsence = false
	identityBad = true
	mu.Unlock()
	if cleanupIPsecOwned(fresh, client, guard) == nil || len(deletions) != 1 {
		t.Fatal("identity mismatch allowed cleanup")
	}
	mu.Lock()
	identityBad = false
	mu.Unlock()
	if err = cleanupIPsecOwned(fresh, client, guard); err != nil {
		t.Fatal("fresh independent cleanup failed", err)
	}
	guard.mu.Lock()
	remaining = len(guard.owned)
	guard.mu.Unlock()
	if remaining != 0 || len(deletions) != 2 || !strings.HasPrefix(deletions[1], "connection/") || activations != 4 || objects[otherIPsecID] == nil {
		t.Fatal("independent cleanup repeated writes, lost order, or touched unowned object")
	}
	// Successful fallback does not alter the failed Core result or its private snapshot.
	if !ipsecErrors(blocked.Diagnostics) || !ipsecPending(blocked.Private) {
		t.Fatal("cleanup masked original Terraform failure")
	}
}
