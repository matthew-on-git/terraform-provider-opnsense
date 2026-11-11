// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

// TestAccLiveInterfacesReadOnly is deliberately opt-in. Credentials, raw API
// inventory and Core artifacts belong in the operator's PRIVATE evidence area.
// A skip is not appliance evidence. This checks actual Core, routing and identity;
// an independent all-field API/state comparison is still required for acceptance.
func TestAccLiveInterfacesReadOnly(t *testing.T) {
	if os.Getenv("OPNSENSE_LIVE_INTERFACE_ACCEPTANCE") != "1" {
		t.Skip("explicit live read-only acceptance opt-in required")
	}
	endpoint, key, secret, device := os.Getenv("OPNSENSE_LIVE_INTERFACE_URL"), os.Getenv("OPNSENSE_LIVE_INTERFACE_KEY"), os.Getenv("OPNSENSE_LIVE_INTERFACE_SECRET"), os.Getenv("OPNSENSE_LIVE_INTERFACE_DEVICE")
	if endpoint == "" || key == "" || secret == "" || device == "" {
		t.Fatal("supply private live URL, key, secret and exact device environment variables")
	}
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("actual Terraform Core required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	target, err := url.Parse(endpoint)
	if err != nil || target.Scheme != "https" || target.Hostname() == "" || target.User != nil || target.RawQuery != "" || target.Fragment != "" || (target.Path != "" && target.Path != "/") {
		t.Fatal("private live URL must be a direct HTTPS origin without credentials, path, query or fragment")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	transport.ResponseHeaderTimeout = 30 * time.Second
	// ReverseProxy executes one RoundTrip, never follows redirects, and fixes its
	// upstream origin from explicit operator configuration, not incoming requests.
	upstream := &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(request *httputil.ProxyRequest) {
			request.SetURL(target)
			request.Out.SetBasicAuth(key, secret)
		},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "private upstream GET failed; check TLS/connectivity", http.StatusBadGateway)
		},
	}
	var mutations, diagnostics, overviews atomic.Int64
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.ContentLength > 0 || r.URL.RawQuery != "" {
			mutations.Add(1)
			http.Error(w, "read-only route rejected", http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/api/core/menu/search":
		case "/api/diagnostics/interface/getInterfaceConfig":
			diagnostics.Add(1)
		case "/api/interfaces/overview/interfaces_info":
			overviews.Add(1)
		default:
			mutations.Add(1)
			http.Error(w, "endpoint rejected", http.StatusForbidden)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		upstream.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer proxy.Close()
	// Ensure the selector is safely quoted in generated private HCL.
	quoted, _ := json.Marshal(device)
	config := fmt.Sprintf(`provider "opnsense" {
 uri=%q
 api_key="loopback-fixture-key"
 api_secret="loopback-fixture-secret"
}
data "opnsense_interface" "one" { device=%s }
data "opnsense_interface_all" "all" {}
data "opnsense_interfaces_overview" "one" { device=%s }
data "opnsense_interfaces_overview_all" "all" {}
`, proxy.URL, quoted, quoted)
	factories := map[string]func() (tfprotov6.ProviderServer, error){"opnsense": providerserver.NewProtocol6WithError(provider.New("live-read-only")())}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: factories, Steps: []resource.TestStep{
		{Config: config, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("data.opnsense_interface.one", "device", device), resource.TestCheckResourceAttr("data.opnsense_interfaces_overview.one", "device", device))},
		{Config: config, PlanOnly: true},
	}})
	if mutations.Load() != 0 || diagnostics.Load() < 4 || overviews.Load() < 4 {
		t.Fatalf("read-only routing proof incomplete: mutations=%d diagnosticGETs=%d overviewGETs=%d", mutations.Load(), diagnostics.Load(), overviews.Load())
	}
}
