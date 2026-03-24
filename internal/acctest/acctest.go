// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package acctest provides shared helpers for acceptance tests.
package acctest

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
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
