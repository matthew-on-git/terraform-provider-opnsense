// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package opnsense provides an HTTP client for the OPNsense REST API.
// It handles authentication, TLS, retries, and connection management.
package opnsense

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/sync/semaphore"
)

// Client is the OPNsense API client. It handles authentication, TLS
// configuration, and HTTP transport for all API interactions.
type Client struct {
	httpClient *http.Client
	baseURL    string
	writeMu    sync.Mutex
	readSem    *semaphore.Weighted
}

// ClientConfig holds configuration for creating a new Client.
type ClientConfig struct {
	// BaseURL is the OPNsense appliance URL (e.g., "https://opnsense.example.com").
	BaseURL string
	// APIKey is the OPNsense API key (used as HTTP Basic Auth username).
	APIKey string //nolint:gosec // Not a hardcoded credential — this is a config field name
	// APISecret is the OPNsense API secret (used as HTTP Basic Auth password).
	APISecret string //nolint:gosec // Not a hardcoded credential — this is a config field name
	// Insecure disables TLS certificate verification. Required for self-signed certificates.
	Insecure bool
	// RetryMax is the maximum number of retries for transient failures. Default: 3.
	RetryMax int
	// RetryWaitMin is the minimum wait time between retries. Default: 1s.
	RetryWaitMin time.Duration
	// RetryWaitMax is the maximum wait time between retries. Default: 30s.
	RetryWaitMax time.Duration
	// MaxReadConcurrency limits concurrent read operations. Default: 10.
	// Protects OPNsense PHP-FPM worker pool from exhaustion.
	MaxReadConcurrency int64
}

// NewClient creates a new OPNsense API client with the given configuration.
// The client uses go-retryablehttp for automatic retries on transient failures
// and a custom RoundTripper for HTTP Basic Auth injection on every request.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if cfg.APISecret == "" {
		return nil, fmt.Errorf("API secret is required")
	}

	// Apply defaults.
	if cfg.RetryMax == 0 {
		cfg.RetryMax = 3
	}
	if cfg.RetryWaitMin == 0 {
		cfg.RetryWaitMin = 1 * time.Second
	}
	if cfg.RetryWaitMax == 0 {
		cfg.RetryWaitMax = 30 * time.Second
	}

	// Configure TLS transport.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Insecure, //nolint:gosec // Intentional: user-configured for self-signed certs
		},
	}

	// Wrap transport with API key authentication.
	authTransport := &apiKeyTransport{
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
		base:      transport,
	}

	// Configure retryable HTTP client.
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = authTransport
	retryClient.RetryMax = cfg.RetryMax
	retryClient.RetryWaitMin = cfg.RetryWaitMin
	retryClient.RetryWaitMax = cfg.RetryWaitMax
	retryClient.Logger = nil // Suppress default logger; we'll use terraform-plugin-log later.

	// Normalize base URL (remove trailing slash).
	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	// Apply read concurrency default.
	maxReads := cfg.MaxReadConcurrency
	if maxReads == 0 {
		maxReads = 10
	}

	return &Client{
		httpClient: retryClient.StandardClient(),
		baseURL:    baseURL,
		readSem:    semaphore.NewWeighted(maxReads),
	}, nil
}

// BaseURL returns the configured OPNsense appliance base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// HTTPClient returns the underlying HTTP client for making requests.
// This is used by the CRUD layer to execute API calls.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// LockMutex acquires the global write mutex, serializing all mutation
// operations. It respects context cancellation to support Terraform shutdown.
// The CRUD layer calls this before every Create, Update, and Delete.
//
// If the context is cancelled while waiting, the background goroutine will
// still eventually acquire the mutex. A cleanup goroutine releases it
// immediately to prevent deadlock.
func (c *Client) LockMutex(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		c.writeMu.Lock()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// The goroutine will eventually acquire the mutex. Release it
		// immediately to prevent a permanent deadlock.
		go func() {
			<-done
			c.writeMu.Unlock()
		}()
		return ctx.Err()
	}
}

// UnlockMutex releases the global write mutex.
func (c *Client) UnlockMutex() {
	c.writeMu.Unlock()
}

// AcquireRead acquires one slot from the read semaphore, limiting concurrent
// read operations to protect OPNsense's PHP-FPM worker pool.
func (c *Client) AcquireRead(ctx context.Context) error {
	return c.readSem.Acquire(ctx, 1)
}

// ReleaseRead releases one slot back to the read semaphore.
func (c *Client) ReleaseRead() {
	c.readSem.Release(1)
}

// apiKeyTransport is an http.RoundTripper that injects HTTP Basic Auth
// credentials on every request. Credentials are never exposed outside
// the transport layer.
type apiKeyTransport struct {
	apiKey    string
	apiSecret string
	base      http.RoundTripper
}

// RoundTrip implements http.RoundTripper. It clones the request and adds
// Basic Auth before delegating to the base transport.
func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.SetBasicAuth(t.apiKey, t.apiSecret)
	return t.base.RoundTrip(clone)
}
