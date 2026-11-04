// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const (
	settingsSaved       = "saved"
	settingsSaveUnknown = "save_unknown"
)

type settingsPrivate interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

func activateSettings(ctx context.Context, c *opnsense.Client) error {
	raw, err := settingsRequest(ctx, c, http.MethodPost, "/api/trust/settings/reconfigure", nil)
	if err != nil {
		return err
	}
	fields, err := settingsObject(raw, []string{"status", "result"})
	if err != nil {
		return err
	}
	var status *string
	if json.Unmarshal(fields["status"], &status) != nil || status == nil || *status != "ok" {
		return errors.New("settings activation requires status ok")
	}
	if value, ok := fields["result"]; ok {
		var result *string
		if json.Unmarshal(value, &result) != nil || result == nil || *result != "ok" {
			return errors.New("contradictory settings activation result")
		}
	}
	return nil
}

func settingsRequest(ctx context.Context, c *opnsense.Client, method, endpoint string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL()+endpoint, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	transport, ok := c.HTTPClient().Transport.(*retryablehttp.RoundTripper)
	if !ok || transport.Client == nil || transport.Client.HTTPClient == nil {
		return nil, errors.New("trust settings requires the authenticated single-attempt transport")
	}
	// Bypass shared automatic retries for non-idempotent settings writes. Copy the
	// underlying authenticated client without changing any shared client settings.
	singleAttempt := *transport.Client.HTTPClient
	singleAttempt.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	requester := &singleAttempt
	if method == http.MethodGet {
		// Copy retry policy and transports; never mutate the shared client's settings.
		configured := transport.Client
		retryPolicy := &retryablehttp.Client{HTTPClient: &singleAttempt, Logger: configured.Logger, RetryWaitMin: configured.RetryWaitMin, RetryWaitMax: configured.RetryWaitMax, RetryMax: configured.RetryMax, RequestLogHook: configured.RequestLogHook, ResponseLogHook: configured.ResponseLogHook, CheckRetry: configured.CheckRetry, Backoff: configured.Backoff, ErrorHandler: configured.ErrorHandler, PrepareRetry: configured.PrepareRetry}
		requester = retryPolicy.StandardClient()
		requester.CheckRedirect = singleAttempt.CheckRedirect
	}
	resp, err := requester.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := opnsense.CheckHTTPError(resp.StatusCode, endpoint); err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}
