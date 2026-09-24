// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reconfigure %s: failed to read response: %w", opts.ReconfigureEndpoint, err)
	}
	if strings.TrimSpace(string(body)) == "" {
		return nil
	}
	var result struct {
		Status string `json:"status"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("reconfigure %s: failed to parse response: %w", opts.ReconfigureEndpoint, err)
	}
	if result.Status == "failed" || result.Result == "failed" {
		return fmt.Errorf("reconfigure %s: failed", opts.ReconfigureEndpoint)
	}

	return nil
}

// FirewallFilterReconfigure returns a ReconfigureFunc that applies pending
// firewall filter changes through OPNsense's filter apply endpoint.
func FirewallFilterReconfigure(client *Client) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if err := firewallApply(ctx, client); err != nil {
			return fmt.Errorf("firewall filter apply failed: %w", err)
		}

		return nil
	}
}

// firewallApply calls POST /api/firewall/filter/apply to apply pending changes.
func firewallApply(ctx context.Context, client *Client) error {
	url := client.BaseURL() + "/api/firewall/filter/apply"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.HTTPClient().Do(req) //nolint:gosec // URL from hardcoded OPNsense API path
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
