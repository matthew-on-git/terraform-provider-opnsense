// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Reconfigure triggers the OPNsense service reconfigure after a successful
// mutation. If opts.ReconfigureFunc is set, it is called instead of the
// standard endpoint. If neither is set, reconfigure is a no-op.
func Reconfigure(ctx context.Context, client *Client, opts ReqOpts) error {
	if opts.ReconfigureFunc != nil {
		return opts.ReconfigureFunc(ctx)
	}
	if opts.ReconfigureEndpoint == "" {
		return nil
	}

	url := client.BaseURL() + opts.ReconfigureEndpoint

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("reconfigure %s: %w", opts.ReconfigureEndpoint, err)
	}

	resp, err := client.HTTPClient().Do(req) //nolint:gosec // URL is from provider-configured ReqOpts, not user input
	if err != nil {
		return fmt.Errorf("reconfigure %s: %w", opts.ReconfigureEndpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reconfigure %s: unexpected status %d", opts.ReconfigureEndpoint, resp.StatusCode)
	}

	return nil
}

// FirewallFilterReconfigure returns a ReconfigureFunc that implements OPNsense's
// 3-step savepoint/apply/cancelRollback flow for firewall filter rules.
//
// This is SAFETY-CRITICAL: if cancelRollback is not called within 60 seconds,
// OPNsense automatically reverts the change. This protects against bad firewall
// rules that lock the operator out of the appliance.
//
// Flow:
//  1. POST /api/firewall/filter/savepoint → get revision
//  2. POST /api/firewall/filter/apply/{revision} → apply with rollback timer
//  3. POST /api/firewall/filter/cancelRollback/{revision} → confirm, cancel auto-revert
func FirewallFilterReconfigure(client *Client) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// Step 1: Create savepoint — get revision ID for rollback tracking.
		revision, err := firewallSavepoint(ctx, client)
		if err != nil {
			return fmt.Errorf("firewall filter savepoint failed: %w", err)
		}

		// Step 2: Apply with rollback protection — starts 60-second auto-revert timer.
		if err := firewallApply(ctx, client, revision); err != nil {
			return fmt.Errorf("firewall filter apply (revision %s) failed: %w", revision, err)
		}

		// Step 3: Confirm the change — cancel the auto-revert timer.
		if err := firewallCancelRollback(ctx, client, revision); err != nil {
			return fmt.Errorf("firewall filter cancelRollback (revision %s) failed — changes will auto-revert in 60 seconds: %w", revision, err)
		}

		return nil
	}
}

// savepointResponse is the JSON structure returned by the savepoint endpoint.
type savepointResponse struct {
	Revision string `json:"revision"`
}

// firewallSavepoint calls POST /api/firewall/filter/savepoint and returns
// the revision string used to track the rollback.
func firewallSavepoint(ctx context.Context, client *Client) (string, error) {
	url := client.BaseURL() + "/api/firewall/filter/savepoint"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.HTTPClient().Do(req) //nolint:gosec // URL from hardcoded OPNsense API path
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result savepointResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse savepoint response: %w", err)
	}

	if result.Revision == "" {
		return "", fmt.Errorf("savepoint returned empty revision")
	}

	return result.Revision, nil
}

// firewallApply calls POST /api/firewall/filter/apply/{revision} to apply
// pending changes with the 60-second automatic rollback safety net.
func firewallApply(ctx context.Context, client *Client, revision string) error {
	url := client.BaseURL() + "/api/firewall/filter/apply/" + revision

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.HTTPClient().Do(req) //nolint:gosec // URL from hardcoded OPNsense API path + revision from savepoint
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}

// firewallCancelRollback calls POST /api/firewall/filter/cancelRollback/{revision}
// to confirm the change and cancel the automatic 60-second revert.
func firewallCancelRollback(ctx context.Context, client *Client, revision string) error {
	url := client.BaseURL() + "/api/firewall/filter/cancelRollback/" + revision

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.HTTPClient().Do(req) //nolint:gosec // URL from hardcoded OPNsense API path + revision from savepoint
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
