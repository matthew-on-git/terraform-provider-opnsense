// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

// This test requires real Terraform Core, not TF_ACC or a mocked state writer.
func TestLiveInterfaceCoreStateAndNoChange(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("actual Terraform Core is required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	var diagnosticReads, overviewReads, mutations atomic.Int64
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.ContentLength > 0 {
			mutations.Add(1)
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/api/core/menu/search":
			_, _ = fmt.Fprint(w, `{"items":[]}`)
		case "/api/diagnostics/interface/getInterfaceConfig":
			diagnosticReads.Add(1)
			_, _ = fmt.Fprint(w, `{"test1":{"device":"test1"},"test0":{"device":"test0","mtu":"1500","media":"","is_physical":false,"ipv4":[{"ipaddr":"192.0.2.2","subnetbits":24,"tunnel":false},{"ipaddr":"192.0.2.1","subnetbits":24,"tunnel":true}],"config":{"secret":"never-state"}}}`)
		case "/api/interfaces/overview/interfaces_info":
			overviewReads.Add(1)
			_, _ = fmt.Fprint(w, `{"total":2,"rowCount":2,"current":1,"rows":[{"device":"test1","identifier":""},{"device":"test0","identifier":"lan","enabled":false,"mtu":"1500","addr4":"192.0.2.2/24","vlan_tag":null,"ifctl.nameserver":["","192.0.2.53 192.0.2.54\n"],"ipv4":[{"ipaddr":"192.0.2.2/24","vhid":"0"}],"config":{"secret":"never-state"}}]}`)
		default:
			mutations.Add(1)
			http.Error(w, "unexpected request", 400)
		}
	}))
	defer api.Close()
	config := fmt.Sprintf(`provider "opnsense" {
 uri=%q
 api_key="fixture-key"
 api_secret="fixture-secret"
}
data "opnsense_interface" "one" { device="test0" }
data "opnsense_interface_all" "all" {}
data "opnsense_interfaces_overview" "one" { device="test0" }
data "opnsense_interfaces_overview_all" "all" {}
`, api.URL)
	factories := map[string]func() (tfprotov6.ProviderServer, error){"opnsense": providerserver.NewProtocol6WithError(provider.New("fixture")())}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: factories, Steps: []resource.TestStep{
		{Config: config, Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.opnsense_interface.one", "mtu", "1500"),
			resource.TestCheckResourceAttr("data.opnsense_interface.one", "media", ""),
			resource.TestCheckResourceAttr("data.opnsense_interface.one", "is_physical", "false"),
			resource.TestCheckResourceAttr("data.opnsense_interface.one", "ipv4.0.ipaddr", "192.0.2.1"),
			resource.TestCheckNoResourceAttr("data.opnsense_interface.one", "status"),
			resource.TestCheckNoResourceAttr("data.opnsense_interface.one", "config"),
			resource.TestCheckResourceAttr("data.opnsense_interface_all.all", "interfaces.#", "2"),
			resource.TestCheckResourceAttr("data.opnsense_interface_all.all", "interfaces.0.device", "test0"),
			resource.TestCheckResourceAttr("data.opnsense_interface_all.all", "interfaces.1.ipv4.#", "0"),
			resource.TestCheckResourceAttr("data.opnsense_interface_all.all", "interfaces.1.flags.#", "0"),
			resource.TestCheckResourceAttr("data.opnsense_interfaces_overview_all.all", "interfaces.1.routes.#", "0"),
			resource.TestCheckResourceAttr("data.opnsense_interfaces_overview.one", "mtu", "1500"),
			resource.TestCheckResourceAttr("data.opnsense_interfaces_overview.one", "enabled", "false"),
			resource.TestCheckResourceAttr("data.opnsense_interfaces_overview.one", "ifctl_nameserver.#", "2"),
			resource.TestCheckNoResourceAttr("data.opnsense_interfaces_overview.one", "vlan"),
			resource.TestCheckNoResourceAttr("data.opnsense_interfaces_overview.one", "config"),
			resource.TestCheckResourceAttr("data.opnsense_interfaces_overview_all.all", "interfaces.1.identifier", ""),
		)},
		{Config: config, PlanOnly: true},
	}})
	if diagnosticReads.Load() < 4 || overviewReads.Load() < 4 || mutations.Load() != 0 {
		t.Fatalf("read-only refresh not proved: diagnostics=%d overview=%d unexpected=%d", diagnosticReads.Load(), overviewReads.Load(), mutations.Load())
	}
}

func TestLiveInterfaceCoreExactSelectionAndEmptyInventory(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("actual Terraform Core required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	for _, test := range []struct {
		name, source, diagnostic, overview, selector string
		failure                                      bool
	}{
		{name: "diagnostic empty single", source: "interface", diagnostic: `[ ]`, overview: `{}`, selector: "test0", failure: true},
		{name: "overview empty single", source: "interfaces_overview", diagnostic: `{}`, overview: `{"total":0,"rowCount":0,"current":1,"rows":[]}`, selector: "test0", failure: true},
		{name: "identifier not alias", source: "interfaces_overview", diagnostic: `{}`, overview: `{"total":1,"rowCount":1,"current":1,"rows":[{"identifier":"lan","device":"test0"}]}`, selector: "lan", failure: true},
		{name: "diagnostic empty aggregate", source: "interface_all", diagnostic: `[ ]`, overview: `{}`},
		{name: "overview empty aggregate", source: "interfaces_overview_all", diagnostic: `{}`, overview: `{"total":0,"rowCount":0,"current":1,"rows":[]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Error("unexpected mutation")
					http.Error(w, "GET only", http.StatusMethodNotAllowed)
					return
				}
				switch r.URL.Path {
				case "/api/core/menu/search":
					_, _ = fmt.Fprint(w, `{"items":[]}`)
				case "/api/diagnostics/interface/getInterfaceConfig":
					_, _ = fmt.Fprint(w, test.diagnostic)
				case "/api/interfaces/overview/interfaces_info":
					_, _ = fmt.Fprint(w, test.overview)
				default:
					t.Error("unexpected route")
					http.Error(w, "route rejected", http.StatusBadRequest)
				}
			}))
			defer api.Close()
			selector := ""
			if test.selector != "" {
				selector = fmt.Sprintf("device=%q", test.selector)
			}
			config := fmt.Sprintf(`provider "opnsense" {
 uri=%q
 api_key="fixture-key"
 api_secret="fixture-secret"
}
data "opnsense_%s" "test" { %s }
`, api.URL, test.source, selector)
			step := resource.TestStep{Config: config}
			if test.failure {
				step.ExpectError = regexp.MustCompile("Interface not found")
			} else {
				step.Check = resource.TestCheckResourceAttr("data.opnsense_"+test.source+".test", "interfaces.#", "0")
			}
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"opnsense": providerserver.NewProtocol6WithError(provider.New("fixture")())}, Steps: []resource.TestStep{step}})
		})
	}
}
