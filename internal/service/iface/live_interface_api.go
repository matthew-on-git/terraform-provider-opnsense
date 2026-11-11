// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const (
	diagnosticEndpoint = "/api/diagnostics/interface/getInterfaceConfig"
	overviewEndpoint   = "/api/interfaces/overview/interfaces_info"
)

func readLiveInterfaces(ctx context.Context, client *opnsense.Client, overview bool) ([]types.Object, error) {
	if client == nil {
		return nil, fmt.Errorf("configure the OPNsense provider before reading interfaces")
	}
	endpoint := diagnosticEndpoint
	if overview {
		endpoint = overviewEndpoint
	}
	if err := client.AcquireRead(ctx); err != nil {
		return nil, fmt.Errorf("interface read cancelled while waiting for read concurrency: %w", err)
	}
	defer client.ReleaseRead()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.BaseURL()+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot construct interface GET request")
	}
	response, err := client.ReadHTTPClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("interface GET failed; check connectivity, TLS trust and request timeout")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode >= 300 && response.StatusCode < 400 {
		return nil, fmt.Errorf("interface endpoint redirected; configure the direct supported API URL and verify access without login redirects")
	}
	// This is a core endpoint. A 404 is not evidence of a missing plugin.
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("interface GET authentication failed (HTTP 401); check API key and secret")
	case http.StatusForbidden:
		return nil, fmt.Errorf("interface GET authorization failed (HTTP 403); grant read ACL access to %s", endpoint)
	case http.StatusNotFound:
		return nil, fmt.Errorf("interface core endpoint unavailable (HTTP 404); check supported firmware and endpoint %s", endpoint)
	}
	if err := opnsense.CheckHTTPError(response.StatusCode, endpoint); err != nil {
		return nil, fmt.Errorf("interface GET failed: %w", err)
	}
	// Bound malformed backend payloads without ever including response bodies.
	body, err := io.ReadAll(io.LimitReader(response.Body, 16*1024*1024+1))
	if err != nil || len(body) > 16*1024*1024 {
		return nil, fmt.Errorf("cannot read bounded interface response")
	}
	return decodeLiveInterfaces(ctx, body, overview)
}

func decodeLiveInterfaces(ctx context.Context, body []byte, overview bool) ([]types.Object, error) {
	if err := validateLiveJSON(body); err != nil {
		return nil, err
	}
	rows := map[string]json.RawMessage{}
	shape := diagnosticShape
	if overview {
		shape = overviewShape
		var envelope map[string]json.RawMessage
		if json.Unmarshal(body, &envelope) != nil || envelope == nil {
			return nil, fmt.Errorf("malformed overview envelope")
		}
		var list []json.RawMessage
		raw, ok := envelope["rows"]
		if !ok || string(raw) == "null" || json.Unmarshal(raw, &list) != nil {
			return nil, fmt.Errorf("malformed overview rows")
		}
		for _, key := range []string{"total", "rowCount", "current"} {
			var count int64
			raw, ok := envelope[key]
			if !ok || string(raw) == "null" || json.Unmarshal(raw, &count) != nil {
				return nil, fmt.Errorf("invalid overview count %s", key)
			}
			expected := int64(len(list))
			if key == "current" {
				expected = 1
			}
			if count != expected {
				return nil, fmt.Errorf("incomplete overview: %s does not match complete first-page inventory", key)
			}
		}
		for _, raw := range list {
			device, err := liveDevice(raw)
			if err != nil {
				return nil, err
			}
			if _, exists := rows[device]; exists {
				return nil, fmt.Errorf("duplicate overview device identity")
			}
			rows[device] = raw
		}
	} else {
		var empty []json.RawMessage
		if json.Unmarshal(body, &empty) == nil && empty != nil && len(empty) == 0 {
			return []types.Object{}, nil
		}
		if json.Unmarshal(body, &rows) != nil || rows == nil {
			return nil, fmt.Errorf("malformed diagnostics interface inventory")
		}
	}
	devices := make([]string, 0, len(rows))
	for key, raw := range rows {
		device, err := liveDevice(raw)
		if err != nil {
			return nil, err
		}
		if key != device {
			return nil, fmt.Errorf("diagnostic key and device identity mismatch")
		}
		devices = append(devices, device)
	}
	sort.Strings(devices)
	result := make([]types.Object, 0, len(devices))
	for _, device := range devices {
		object, err := shape.decode(ctx, rows[device], !overview)
		if err != nil {
			return nil, fmt.Errorf("invalid interface observation: %w", err)
		}
		result = append(result, object)
	}
	return result, nil
}

func liveDevice(raw json.RawMessage) (string, error) {
	var row map[string]json.RawMessage
	if json.Unmarshal(raw, &row) != nil || row == nil {
		return "", fmt.Errorf("expected interface row object")
	}
	var device string
	if json.Unmarshal(row["device"], &device) != nil || strings.TrimSpace(device) == "" {
		return "", fmt.Errorf("interface row requires nonempty device identity")
	}
	return device, nil
}

// Reject duplicate object keys before map decoding can silently discard an identity.
func validateLiveJSON(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := uniqueLiveJSONValue(decoder); err != nil {
		return fmt.Errorf("malformed interface JSON or duplicate object key")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("malformed trailing interface JSON")
	}
	return nil
}

func uniqueLiveJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			str, ok := key.(string)
			if !ok || seen[str] {
				return fmt.Errorf("duplicate key")
			}
			seen[str] = true
			if err := uniqueLiveJSONValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := uniqueLiveJSONValue(decoder); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("invalid delimiter")
	}
	_, err = decoder.Token()
	return err
}
