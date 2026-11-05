// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func journalGuard(t *testing.T) (*ipsecLeaseGuard, string) {
	t.Helper()
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	hash := sha256.Sum256([]byte("identity"))
	g := &ipsecLeaseGuard{receipt: ipsecIsolationReceipt{TargetURI: "https://firewall.example.test", GuestID: "fixture-guest", ReservationID: "fixture-reservation", ReservedBy: "fixture-operator", ExpiresAt: now.Add(time.Hour), IdentityEndpoint: "/api/core/identity/get", IdentitySHA256: hex.EncodeToString(hash[:])}, now: func() time.Time { return now }, owned: map[string]string{}}
	directory := filepath.Join(t.TempDir(), "private")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(directory, "ownership.json")
	if err := g.initializeOwnershipJournal(filename); err != nil {
		t.Fatal(err)
	}
	return g, filename
}

func journalPost(g *ipsecLeaseGuard, base http.RoundTripper, endpoint string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, g.receipt.TargetURI+endpoint, nil)
	if err != nil {
		return err
	}
	response, err := (&ipsecGuardTransport{guard: g, base: base}).RoundTrip(req)
	if response != nil {
		_ = response.Body.Close()
	}
	return err
}

func journalResponse(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}
}

func TestIPsecJournalDispatchRecovery(t *testing.T) {
	for name, reply := range map[string]string{"confirmed": `{"result":"saved","uuid":"` + ipsecID + `"}`, "rejected": `{"result":"failed"}`, "missing UUID": `{"result":"saved"}`, "invalid UUID": `{"result":"saved","uuid":"unsafe/path"}`, "malformed": `{"result":`, "truncated": strings.Repeat("x", (4<<20)+1), "lost": ""} {
		t.Run(name, func(t *testing.T) {
			g, filename := journalGuard(t)
			posts := 0
			base := ipsecRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method == http.MethodGet {
					return journalResponse("identity"), nil
				}
				posts++
				// Reload at dispatch proves ambiguity was made durable before the POST.
				reload := &ipsecLeaseGuard{receipt: g.receipt, now: g.now}
				if err := reload.initializeOwnershipJournal(filename); err != nil {
					t.Fatal(err)
				}
				if len(reload.unresolvedAdds) != 1 || !reload.pendingActivation {
					t.Fatal("dispatch was not durably recorded")
				}
				if name == "lost" {
					return nil, fmt.Errorf("fixture connection lost")
				}
				return journalResponse(reply), nil
			})
			_ = journalPost(g, base, "/api/ipsec/connections/add_connection")
			reload := &ipsecLeaseGuard{receipt: g.receipt, now: g.now}
			if err := reload.initializeOwnershipJournal(filename); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "confirmed":
				if reload.owned[ipsecID] != "connection" || len(reload.unresolvedAdds) != 0 || !reload.pendingActivation {
					t.Fatal("confirmed ownership not recovered")
				}
			case "rejected":
				if len(reload.owned) != 0 || len(reload.unresolvedAdds) != 0 || reload.pendingActivation {
					t.Fatal("rejection retained uncertainty")
				}
			default:
				if len(reload.unresolvedAdds) != 1 {
					t.Fatal("uncertain add was forgotten")
				}
				if journalPost(g, base, "/api/ipsec/connections/add_connection") == nil || posts != 1 {
					t.Fatal("uncertain add retried")
				}
				g.owned[ipsecID] = "connection"
				if journalPost(g, base, "/api/ipsec/connections/del_connection/"+ipsecID) == nil || posts != 1 {
					t.Fatal("uncertain child/add allowed parent cascading cleanup")
				}
				if journalPost(g, base, "/api/ipsec/service/reconfigure") == nil || posts != 1 {
					t.Fatal("uncertain add allowed service activation")
				}

				client, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: g.receipt.TargetURI, APIKey: "fixture", APISecret: "fixture"})
				if err != nil {
					t.Fatal(err)
				}
				if cleanupIPsecOwned(context.Background(), client, reload) == nil {
					t.Fatal("uncertain ownership reported clean")
				}
			}
			info, err := os.Stat(filename)
			if err != nil || info.Mode().Perm() != 0o600 {
				t.Fatal("journal not private")
			}
			mismatch := &ipsecLeaseGuard{receipt: g.receipt, now: g.now}
			mismatch.receipt.ReservationID = "different"
			if mismatch.initializeOwnershipJournal(filename) == nil {
				t.Fatal("reservation binding ignored")
			}
		})
	}
}

func TestIPsecJournalFailureBlocksDispatch(t *testing.T) {
	g, filename := journalGuard(t)
	// Removing the reserved directory makes subsequent atomic persistence fail.
	if err := os.Remove(filename); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Dir(filename)); err != nil {
		t.Fatal(err)
	}
	posts := 0
	base := ipsecRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodPost {
			posts++
		}
		return journalResponse("identity"), nil
	})
	if journalPost(g, base, "/api/ipsec/connections/add_connection") == nil || posts != 0 {
		t.Fatal("POST dispatched without durable ownership")
	}
}

func TestIPsecCleanupHardExpiryAndIdentityLatency(t *testing.T) {
	for _, latency := range []bool{false, true} {
		t.Run(fmt.Sprintf("latency=%t", latency), func(t *testing.T) {
			g, _ := journalGuard(t)
			g.owned[ipsecID] = "connection"
			g.cleanupActivation = true
			now := g.receipt.ExpiresAt
			if latency {
				now = now.Add(-time.Second)
			}
			g.now = func() time.Time { return now }
			posts := 0
			base := ipsecRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method == http.MethodPost {
					posts++
				} else if latency {
					now = g.receipt.ExpiresAt
				}
				return journalResponse("identity"), nil
			})
			for _, endpoint := range []string{"/api/ipsec/connections/del_connection/" + ipsecID, "/api/ipsec/service/reconfigure"} {
				if latency {
					now = g.receipt.ExpiresAt.Add(-time.Second)
				}
				if journalPost(g, base, endpoint) == nil || posts != 0 {
					t.Fatal("cleanup dispatched after hard lease expiry")
				}
			}
		})
	}
}
