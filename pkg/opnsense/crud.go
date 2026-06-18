// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Add creates a new resource via the OPNsense API. The resource struct is
// wrapped in the monad key and POSTed to the AddEndpoint. Returns the UUID
// of the created resource. Acquires the write mutex and calls reconfigure
// after success.
func Add[K any](ctx context.Context, c *Client, opts ReqOpts, resource *K) (string, error) {
	if err := c.LockMutex(ctx); err != nil {
		return "", fmt.Errorf("add %s: %w", opts.AddEndpoint, err)
	}
	defer c.UnlockMutex()

	body, err := marshalWrapped(opts.Monad, resource)
	if err != nil {
		return "", fmt.Errorf("add %s: %w", opts.AddEndpoint, err)
	}

	url := c.BaseURL() + opts.AddEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("add %s: %w", opts.AddEndpoint, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient().Do(req) //nolint:gosec // URL from provider-configured ReqOpts
	if err != nil {
		return "", NewServerError(opts.AddEndpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if httpErr := CheckHTTPError(resp.StatusCode, opts.AddEndpoint); httpErr != nil {
		return "", httpErr
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("add %s: failed to read response: %w", opts.AddEndpoint, err)
	}

	uuid, err := ParseMutationResponse(respBody)
	if err != nil {
		return "", err
	}

	if err := Reconfigure(ctx, c, opts); err != nil {
		return uuid, &MutationReconfigureError{Operation: "add", Endpoint: opts.AddEndpoint, ResourceID: uuid, Cause: err}
	}

	return uuid, nil
}

// Get reads a resource by UUID from the OPNsense API. The response is
// unwrapped from the monad key and returned as a clean struct. Returns
// NotFoundError if the resource doesn't exist. Acquires the read semaphore.
func Get[K any](ctx context.Context, c *Client, opts ReqOpts, id string) (*K, error) {
	if err := c.AcquireRead(ctx); err != nil {
		return nil, fmt.Errorf("get %s: %w", opts.GetEndpoint, err)
	}
	defer c.ReleaseRead()

	url := c.BaseURL() + opts.GetEndpoint + "/" + id
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", opts.GetEndpoint, err)
	}

	resp, err := c.HTTPClient().Do(req) //nolint:gosec // URL from provider-configured ReqOpts
	if err != nil {
		return nil, NewServerError(opts.GetEndpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if httpErr := CheckHTTPError(resp.StatusCode, opts.GetEndpoint); httpErr != nil {
		return nil, httpErr
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("get %s: failed to read response: %w", opts.GetEndpoint, err)
	}

	return unmarshalWrapped[K](opts.Monad, respBody)
}

// Update modifies an existing resource via the OPNsense API. The resource
// struct is wrapped in the monad key and POSTed to UpdateEndpoint/{id}.
// Acquires the write mutex and calls reconfigure after success.
func Update[K any](ctx context.Context, c *Client, opts ReqOpts, resource *K, id string) error {
	if err := c.LockMutex(ctx); err != nil {
		return fmt.Errorf("update %s: %w", opts.UpdateEndpoint, err)
	}
	defer c.UnlockMutex()

	body, err := marshalWrapped(opts.Monad, resource)
	if err != nil {
		return fmt.Errorf("update %s: %w", opts.UpdateEndpoint, err)
	}

	url := c.BaseURL() + opts.UpdateEndpoint + "/" + id
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("update %s: %w", opts.UpdateEndpoint, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient().Do(req) //nolint:gosec // URL from provider-configured ReqOpts
	if err != nil {
		return NewServerError(opts.UpdateEndpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if httpErr := CheckHTTPError(resp.StatusCode, opts.UpdateEndpoint); httpErr != nil {
		return httpErr
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("update %s: failed to read response: %w", opts.UpdateEndpoint, err)
	}

	if _, err := ParseMutationResponse(respBody); err != nil {
		return err
	}

	if err := Reconfigure(ctx, c, opts); err != nil {
		return &MutationReconfigureError{Operation: "update", Endpoint: opts.UpdateEndpoint, ResourceID: id, Cause: err}
	}

	return nil
}

// Delete removes a resource by UUID from the OPNsense API. POSTs to
// DeleteEndpoint/{id}. Acquires the write mutex and calls reconfigure
// after success. Parses delete failures while accepting OPNsense's normal
// non-"saved" delete success result values such as "deleted".
func Delete(ctx context.Context, c *Client, opts ReqOpts, id string) error {
	if err := c.LockMutex(ctx); err != nil {
		return fmt.Errorf("delete %s: %w", opts.DeleteEndpoint, err)
	}
	defer c.UnlockMutex()

	url := c.BaseURL() + opts.DeleteEndpoint + "/" + id
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("delete %s: %w", opts.DeleteEndpoint, err)
	}

	resp, err := c.HTTPClient().Do(req) //nolint:gosec // URL from provider-configured ReqOpts
	if err != nil {
		return NewServerError(opts.DeleteEndpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if httpErr := CheckHTTPError(resp.StatusCode, opts.DeleteEndpoint); httpErr != nil {
		return httpErr
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("delete %s: failed to read response: %w", opts.DeleteEndpoint, err)
	}
	if err := ParseDeleteResponse(respBody); err != nil {
		return err
	}

	if err := Reconfigure(ctx, c, opts); err != nil {
		return &MutationReconfigureError{Operation: "delete", Endpoint: opts.DeleteEndpoint, ResourceID: id, Cause: err}
	}

	return nil
}

// GetSingleton reads a singleton settings object from a get endpoint that has
// no UUID path segment (e.g. "/api/quagga/general/get"). The response is
// unwrapped from the monad key and returned as a clean struct. Returns
// NotFoundError if the monad value is missing or empty. Acquires the read
// semaphore and never calls reconfigure.
func GetSingleton[K any](ctx context.Context, c *Client, opts ReqOpts) (*K, error) {
	if err := c.AcquireRead(ctx); err != nil {
		return nil, fmt.Errorf("get %s: %w", opts.GetEndpoint, err)
	}
	defer c.ReleaseRead()

	url := c.BaseURL() + opts.GetEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", opts.GetEndpoint, err)
	}

	resp, err := c.HTTPClient().Do(req) //nolint:gosec // URL from provider-configured ReqOpts
	if err != nil {
		return nil, NewServerError(opts.GetEndpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if httpErr := CheckHTTPError(resp.StatusCode, opts.GetEndpoint); httpErr != nil {
		return nil, httpErr
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("get %s: failed to read response: %w", opts.GetEndpoint, err)
	}

	return unmarshalWrapped[K](opts.Monad, respBody)
}

// UpdateSingleton updates a singleton settings object via a set endpoint that
// has no UUID path segment (e.g. "/api/quagga/general/set"). The resource
// struct is wrapped in the monad key and POSTed. Acquires the write mutex and
// calls reconfigure after success.
func UpdateSingleton[K any](ctx context.Context, c *Client, opts ReqOpts, resource *K) error {
	if err := c.LockMutex(ctx); err != nil {
		return fmt.Errorf("update %s: %w", opts.UpdateEndpoint, err)
	}
	defer c.UnlockMutex()

	body, err := marshalWrapped(opts.Monad, resource)
	if err != nil {
		return fmt.Errorf("update %s: %w", opts.UpdateEndpoint, err)
	}

	url := c.BaseURL() + opts.UpdateEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("update %s: %w", opts.UpdateEndpoint, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient().Do(req) //nolint:gosec // URL from provider-configured ReqOpts
	if err != nil {
		return NewServerError(opts.UpdateEndpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if httpErr := CheckHTTPError(resp.StatusCode, opts.UpdateEndpoint); httpErr != nil {
		return httpErr
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("update %s: failed to read response: %w", opts.UpdateEndpoint, err)
	}

	if _, err := ParseMutationResponse(respBody); err != nil {
		return err
	}

	if err := Reconfigure(ctx, c, opts); err != nil {
		return &MutationReconfigureError{Operation: "update", Endpoint: opts.UpdateEndpoint, Cause: err}
	}

	return nil
}

// marshalWrapped marshals a resource struct wrapped in the monad key. The monad
// may be a dotted path to nest deeper than one level, which is required for
// settings singletons whose model nests fields under a sub-section.
// Examples: monad="server", resource={Name:"web1"} → {"server":{"name":"web1"}};
// monad="unbound.general" → {"unbound":{"general":{...}}}.
func marshalWrapped[K any](monad string, resource *K) ([]byte, error) {
	var wrapped interface{} = resource
	parts := strings.Split(monad, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		wrapped = map[string]interface{}{parts[i]: wrapped}
	}
	return json.Marshal(wrapped)
}

// unmarshalWrapped extracts a resource struct from a monad-wrapped JSON response.
// The monad may be a dotted path (e.g. "unbound.general") to descend through
// nested sections. Returns NotFoundError if any path segment is missing or the
// inner value is empty.
func unmarshalWrapped[K any](monad string, body []byte) (*K, error) {
	inner := json.RawMessage(body)
	for _, part := range strings.Split(monad, ".") {
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(inner, &envelope); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		next, ok := envelope[part]
		if !ok || len(next) == 0 || string(next) == "null" || string(next) == "{}" {
			return nil, &NotFoundError{Message: "resource not found"}
		}
		inner = next
	}

	var result K
	if err := json.Unmarshal(inner, &result); err != nil {
		return nil, &NotFoundError{Message: fmt.Sprintf("failed to unmarshal resource: %s", err)}
	}

	return &result, nil
}
