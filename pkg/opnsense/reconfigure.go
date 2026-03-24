// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"context"
	"fmt"
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
