// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var frrUUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func bgpSnapshotIDs(ctx context.Context, c *opnsense.Client, opts opnsense.ReqOpts) (map[string]bool, error) {
	rows, err := opnsense.Search[struct {
		UUID string `json:"uuid"`
	}](ctx, c, opts, opnsense.SearchParams{})
	if err != nil {
		return nil, err
	}
	ids := map[string]bool{}
	for _, row := range rows {
		if !frrUUID.MatchString(row.UUID) || ids[row.UUID] {
			return nil, fmt.Errorf("FRR identity search returned empty/duplicate UUID")
		}
		ids[row.UUID] = true
	}
	return ids, nil
}

func exactBGPPayload(remote map[string]json.RawMessage, want map[string]string) bool {
	for key, expected := range want {
		raw, exists := remote[key]
		if !exists {
			return false
		}
		if key == "password" {
			return false
		} // Authentication cannot be reconciled from remote responses.
		list := key == "capabilities" || key == "listenranges" || key == "family" || key == "match" || key == "match2" || key == "match3" || key == "networks" || strings.HasPrefix(key, "linkedPrefixlist") || (strings.HasPrefix(key, "linkedRoutemap") && key != "linkedRoutemap")
		selector := key == "linklocalinterface" || key == "removeprivateas" || key == "allowas_in" || key == "attributeunchanged" || list || strings.HasPrefix(key, "linkedPrefixlist") || strings.HasPrefix(key, "linkedRoutemap") || key == "peergroup" || key == "updatesource" || key == "action" || key == "version" || key == "remote_as_mode" || key == "redistribute" || key == "profile" || key == "sysloglevel"
		if !selector {
			var scalar *string
			if json.Unmarshal(raw, &scalar) != nil || scalar == nil || *scalar != expected {
				return false
			}
			continue
		}

		got, err := decodeFRRSelection(raw)
		if err != nil {
			return false
		}
		expectedMembers := []string{}
		if expected != "" {
			expectedMembers = strings.Split(expected, ",")
		}
		if !list && (len(got) > 1 || len(expectedMembers) > 1) {
			return false
		}
		sort.Strings(got)
		sort.Strings(expectedMembers)
		if strings.Join(got, ",") != strings.Join(expectedMembers, ",") {
			return false
		}
	}
	return true
}

func createBGPConfirmed(ctx context.Context, c *opnsense.Client, opts opnsense.ReqOpts, payload *map[string]string) (string, error) {
	before, err := bgpSnapshotIDs(ctx, c, opts)
	if err != nil {
		return "", err
	}
	id, mutationErr := opnsense.Add(ctx, c, opts, payload)
	if id != "" && !frrUUID.MatchString(id) {
		id = ""
		mutationErr = fmt.Errorf("FRR create response did not provide a valid UUID; authoritative recovery required")
	}
	if id != "" || mutationErr == nil {
		return id, mutationErr
	}
	var validation *opnsense.ValidationError
	if errors.As(mutationErr, &validation) {
		return "", mutationErr
	}
	after, err := bgpSnapshotIDs(ctx, c, opts)
	if err != nil {
		return "", fmt.Errorf("FRR create outcome ambiguous; authoritative recovery search failed")
	}
	matches := []string{}
	for candidate := range after {
		if before[candidate] {
			continue
		}
		remote, err := opnsense.Get[map[string]json.RawMessage](ctx, c, opts, candidate)
		if err != nil {
			return "", fmt.Errorf("FRR create outcome ambiguous; candidate read failed")
		}
		if exactBGPPayload(*remote, *payload) {
			matches = append(matches, candidate)
		}
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("FRR create outcome ambiguous: %d exact new-object matches; reconcile appliance objects before retrying; no identity was adopted", len(matches))
	}
	// The failed response does not prove activation; retain the exact, newly
	// committed object's UUID and a durable pending marker, never replay add.
	return matches[0], fmt.Errorf("FRR saved object recovered by unique full-payload match; activation must be retried")
}

// A failed save acknowledgement is not evidence that the singleton was not
// committed. Adopt desired state only after exact authoritative read-back.
func saveFRRSingleton[K any](ctx context.Context, c *opnsense.Client, opts opnsense.ReqOpts, payload *K) error {
	err := opnsense.UpdateSingleton(ctx, c, strictBGPOpts(opts), payload)
	if err == nil {
		return nil
	}
	var saved *opnsense.MutationReconfigureError
	var validation *opnsense.ValidationError
	if errors.As(err, &saved) || errors.As(err, &validation) {
		return err
	}
	remote, readErr := getFRRSingleton[map[string]json.RawMessage](ctx, c, opts)
	encoded, encodeErr := json.Marshal(payload)
	var expected map[string]string
	if readErr != nil || encodeErr != nil || json.Unmarshal(encoded, &expected) != nil || !exactBGPPayload(*remote, expected) {
		return err
	}
	return &opnsense.MutationReconfigureError{Operation: "save reconciled", Endpoint: opts.UpdateEndpoint, Cause: err}
}
