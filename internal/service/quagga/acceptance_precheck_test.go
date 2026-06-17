// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"io"
	"strings"
	"testing"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func preCheck(t *testing.T) {
	t.Helper()
	acctest.PreCheck(t)

	client := acctest.TestClient(t)
	resp, err := client.HTTPClient().Get(client.BaseURL() + "/api/quagga/service/reconfigure")
	if err != nil {
		t.Skipf("FRR/Quagga reconfigure endpoint is not reachable on this appliance: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 500 || strings.Contains(string(body), `"status":"failed"`) {
		t.Skipf("FRR/Quagga reconfigure endpoint is unhealthy on this appliance: HTTP %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}
