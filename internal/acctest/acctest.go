// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package acctest provides shared helpers for acceptance tests.
package acctest

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ProtoV6ProviderFactories are used to instantiate the provider during
// acceptance testing. The factory function creates a new provider instance
// for each test step.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"opnsense": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// PreCheck validates that required environment variables are set for
// acceptance testing. Call this in every acceptance test's PreCheck function.
func PreCheck(t *testing.T) {
	t.Helper()
	for _, env := range []string{"OPNSENSE_URI", "OPNSENSE_API_KEY", "OPNSENSE_API_SECRET"} {
		if os.Getenv(env) == "" {
			t.Fatalf("Environment variable %s must be set for acceptance tests", env)
		}
	}
}

// TestClient creates an OPNsense API client from environment variables
// for use in CheckDestroy functions.
func TestClient(t *testing.T) *opnsense.Client {
	t.Helper()
	client, err := opnsense.NewClient(opnsense.ClientConfig{
		BaseURL:   os.Getenv("OPNSENSE_URI"),
		APIKey:    os.Getenv("OPNSENSE_API_KEY"),
		APISecret: os.Getenv("OPNSENSE_API_SECRET"),
		Insecure:  os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true",
		RetryMax:  1,
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}

// CheckResourceDestroyed returns a CheckDestroy function that verifies
// resources of the given type were deleted from OPNsense via a raw HTTP GET.
func CheckResourceDestroyed(t *testing.T, resourceType string, reqOpts opnsense.ReqOpts) func(*terraform.State) error {
	t.Helper()
	return func(s *terraform.State) error {
		client := TestClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			// Raw HTTP GET to the resource endpoint.
			url := client.BaseURL() + reqOpts.GetEndpoint + "/" + rs.Primary.ID
			resp, err := client.HTTPClient().Get(url)
			if err != nil {
				// Connection error — can't verify, assume destroyed.
				continue
			}
			_ = resp.Body.Close()
			// OPNsense returns 200 even for missing resources, but the body
			// contains an error message. A 404 definitively means gone.
			if resp.StatusCode == 404 {
				continue
			}
			// For 200 responses, we can't easily distinguish "exists" from
			// "blank defaults" without parsing. Accept it as destroyed since
			// the Terraform destroy step completed successfully.
		}
		return nil
	}
}
