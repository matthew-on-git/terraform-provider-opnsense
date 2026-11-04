// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAssignmentApplyUsesExpectedPathsAndReadback(t *testing.T) {
	t.Parallel()
	current := assignmentFixture("192.0.2.10/24", "2001:db8::10/64")
	var mu sync.Mutex
	var paths []string
	var posted map[string]map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		paths = append(paths, req.Method+" "+req.URL.Path)
		mu.Unlock()
		switch req.URL.Path {
		case "/api/core/firmware/info":
			_, _ = io.WriteString(w, `{"product_version":"26.7.6"}`)
		case "/api/interfaces/assignment/get_item/opt3":
			mu.Lock()
			body, _ := json.Marshal(map[string]any{"interface": current})
			mu.Unlock()
			_, _ = w.Write(body)
		case "/api/interfaces/assignment/set_item/opt3":
			if req.Method != http.MethodPost {
				t.Errorf("set_item method = %s", req.Method)
			}
			if err := json.NewDecoder(req.Body).Decode(&posted); err != nil {
				t.Errorf("decode payload: %v", err)
			}
			mu.Lock()
			for key, value := range posted["interface"] {
				current[key] = value
			}
			mu.Unlock()
			_, _ = io.WriteString(w, `{"result":"saved"}`)
		case "/api/interfaces/assignment/reconfigure":
			_, _ = io.WriteString(w, `{"status":"ok"}`)
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()

	r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
	plan := AssignmentResourceModel{
		Name:             types.StringValue("opt3"),
		Enabled:          types.BoolValue(true),
		IPv4Mode:         types.StringValue("staticv4"),
		IPv4Address:      types.StringValue("192.0.2.20"),
		IPv4PrefixLength: types.Int64Value(25),
		IPv4Gateway:      types.StringValue("GW_TEST"),
		IPv6Mode:         types.StringValue("none"),
	}
	var got AssignmentResourceModel
	var diagnostics diag.Diagnostics
	r.apply(context.Background(), &plan, "opt3", "updating", &diagnostics, func(model *AssignmentResourceModel) {
		got = *model
	})
	if diagnostics.HasError() {
		t.Fatalf("apply diagnostics: %v", diagnostics)
	}
	if got.IPv4Address.ValueString() != "192.0.2.20" || got.IPv4PrefixLength.ValueInt64() != 25 {
		t.Fatalf("readback = %s/%d", got.IPv4Address.ValueString(), got.IPv4PrefixLength.ValueInt64())
	}
	payload := posted["interface"]
	if payload["ipaddr"] != "192.0.2.20/25" || payload["ipaddrv6"] != "" || payload["type6"] != "" {
		t.Fatalf("unexpected address payload: %#v", payload)
	}
	if payload["if"] != "vtnet2" || payload["descr"] != "disposable" || payload["blockpriv"] != "1" {
		t.Fatalf("unmanaged fields were not preserved: %#v", payload)
	}
	if payload["disablevlanhwfilter"] != "0" {
		t.Fatalf("indexed choice field was not normalized: %#v", payload["disablevlanhwfilter"])
	}
	wantSequence := []string{
		"GET /api/core/firmware/info", "GET /api/interfaces/assignment/get_item/opt3",
		"POST /api/interfaces/assignment/set_item/opt3", "POST /api/interfaces/assignment/reconfigure",
		"GET /api/core/firmware/info", "GET /api/interfaces/assignment/get_item/opt3",
	}
	mu.Lock()
	defer mu.Unlock()
	if fmt.Sprint(paths) != fmt.Sprint(wantSequence) {
		t.Fatalf("request sequence = %v, want %v", paths, wantSequence)
	}
}

func TestAssignmentApplySurfacesHTTP200Validation(t *testing.T) {
	t.Parallel()
	server := assignmentTestServer(t, "26.7.6", assignmentFixture("", ""), http.StatusOK, `{"result":"failed","validations":{"interface.ipaddr":"invalid address"}}`)
	defer server.Close()
	r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
	plan := AssignmentResourceModel{Name: types.StringValue("opt3"), IPv4Mode: types.StringValue("none"), IPv6Mode: types.StringValue("none")}
	var diagnostics diag.Diagnostics
	r.apply(context.Background(), &plan, "opt3", "updating", &diagnostics, func(*AssignmentResourceModel) {})
	if !diagnostics.HasError() || !strings.Contains(diagnostics[0].Detail(), "interface.ipaddr: invalid address") {
		t.Fatalf("validation diagnostics = %v", diagnostics)
	}
}

func TestAssignmentApplyRecoversPersistedStateAfterReconfigureFailure(t *testing.T) {
	t.Parallel()
	current := assignmentFixture("192.0.2.10/24", "")
	getCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/core/firmware/info":
			_, _ = io.WriteString(w, `{"product_version":"26.7.6"}`)
		case "/api/interfaces/assignment/get_item/opt3":
			getCount++
			_ = json.NewEncoder(w).Encode(map[string]any{"interface": current})
		case "/api/interfaces/assignment/set_item/opt3":
			var payload map[string]map[string]any
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			for key, value := range payload["interface"] {
				current[key] = value
			}
			_, _ = io.WriteString(w, `{"result":"saved"}`)
		case "/api/interfaces/assignment/reconfigure":
			_, _ = io.WriteString(w, `{"status":"failed"}`)
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()

	r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
	plan := AssignmentResourceModel{
		Name: types.StringValue("opt3"), IPv4Mode: types.StringValue("staticv4"),
		IPv4Address: types.StringValue("192.0.2.44"), IPv4PrefixLength: types.Int64Value(25),
	}
	var got AssignmentResourceModel
	var diagnostics diag.Diagnostics
	r.apply(context.Background(), &plan, "opt3", "updating", &diagnostics, func(model *AssignmentResourceModel) { got = *model })
	if getCount != 2 {
		t.Fatalf("assignment GET count = %d, want preflight and recovery read-back", getCount)
	}
	if got.ID.ValueString() != "opt3" || got.IPv4Address.ValueString() != "192.0.2.44" || got.IPv4PrefixLength.ValueInt64() != 25 {
		t.Fatalf("persisted recovery state = %#v", got)
	}
	if !diagnostics.HasError() || diagnostics[0].Summary() != "Interface assignment saved but not activated" || !strings.Contains(diagnostics[0].Detail(), "may not be active") {
		t.Fatalf("reconfigure diagnostics = %v", diagnostics)
	}
}

func TestAssignmentCreatePersistsIdentityAfterMandatoryReadbackFailure(t *testing.T) {
	t.Parallel()
	assignmentGets := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/core/firmware/info":
			_, _ = io.WriteString(w, `{"product_version":"26.7.6"}`)
		case "/api/interfaces/assignment/get_item/opt3":
			assignmentGets++
			if assignmentGets == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{"interface": assignmentFixture("192.0.2.10/24", "")})
				return
			}
			http.Error(w, "temporary failure", http.StatusServiceUnavailable)
		case "/api/interfaces/assignment/set_item/opt3":
			_, _ = io.WriteString(w, `{"result":"saved"}`)
		case "/api/interfaces/assignment/reconfigure":
			_, _ = io.WriteString(w, `{"status":"ok"}`)
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()

	r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
	planModel := AssignmentResourceModel{
		Name: types.StringValue("opt3"), IPv4Mode: types.StringValue("staticv4"),
		IPv4Address: types.StringValue("192.0.2.44"), IPv4PrefixLength: types.Int64Value(24),
		IPv6Mode: types.StringValue("none"),
	}
	planState := assignmentTestState(t, r, planModel)
	response := resource.CreateResponse{State: tfsdk.State{Schema: planState.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(planState)}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("expected mandatory read-back diagnostic")
	}
	var recovered AssignmentResourceModel
	if diagnostics := response.State.Get(context.Background(), &recovered); diagnostics.HasError() {
		t.Fatalf("read recovery state: %v", diagnostics)
	}
	if recovered.ID.ValueString() != "opt3" || recovered.Name.ValueString() != "opt3" {
		t.Fatalf("Create recovery identity = id %q, name %q", recovered.ID.ValueString(), recovered.Name.ValueString())
	}
}

func TestAssignmentCompatibilityDiagnostics(t *testing.T) {
	t.Parallel()
	missingEnable := assignmentFixture("", "")
	delete(missingEnable, "enable")
	missingInterface := assignmentFixture("", "")
	delete(missingInterface, "if")
	tests := []struct {
		name       string
		version    string
		status     int
		assignment map[string]any
		body       string
		want       string
	}{
		{name: "old firmware", version: "26.7.5", status: http.StatusOK, assignment: assignmentFixture("", ""), want: "older than 26.7.6"},
		{name: "partial model", version: "26.7.6", status: http.StatusOK, assignment: map[string]any{"type4": "none"}, want: "missing required fields"},
		{name: "missing enable", version: "26.7.6", status: http.StatusOK, assignment: missingEnable, want: "enable"},
		{name: "missing interface", version: "26.7.6", status: http.StatusOK, assignment: missingInterface, want: "if"},
		{name: "missing endpoint body", version: "26.7.6", status: http.StatusOK, body: `{"message":"Endpoint not found"}`, want: "complete expanded interfaces/assignment API"},
		{name: "unauthorized", version: "26.7.6", status: http.StatusForbidden, body: `{}`, want: "HTTP 403"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := assignmentTestServer(t, tt.version, tt.assignment, tt.status, tt.body)
			defer server.Close()
			r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
			_, err := r.readCompatibleAssignment(context.Background(), "opt3")
			if err == nil || !strings.Contains(err.Error(), tt.want) || !strings.Contains(err.Error(), "26.7.6") {
				t.Fatalf("compatibility error = %v, want %q and minimum version", err, tt.want)
			}
		})
	}
}

func TestAssignmentRejectsUnsupportedExistingModesBeforeMutation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		field string
		mode  string
	}{
		{name: "PPPoE", field: "type4", mode: "pppoe"},
		{name: "6rd", field: "type6", mode: "6rd"},
		{name: "6to4", field: "type6", mode: "6to4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := assignmentFixture("", "")
			fixture[tt.field] = selected(tt.mode)
			mutations := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				switch req.URL.Path {
				case "/api/core/firmware/info":
					_, _ = io.WriteString(w, `{"product_version":"26.7.6"}`)
				case "/api/interfaces/assignment/get_item/opt3":
					_ = json.NewEncoder(w).Encode(map[string]any{"interface": fixture})
				case "/api/interfaces/assignment/set_item/opt3", "/api/interfaces/assignment/reconfigure":
					mutations++
				default:
					http.NotFound(w, req)
				}
			}))
			defer server.Close()
			r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
			var diagnostics diag.Diagnostics
			r.apply(context.Background(), &AssignmentResourceModel{Name: types.StringValue("opt3")}, "opt3", "updating", &diagnostics, func(*AssignmentResourceModel) {})
			if !diagnostics.HasError() || !strings.Contains(diagnostics[0].Detail(), "unsupported") || !strings.Contains(diagnostics[0].Detail(), tt.mode) {
				t.Fatalf("unsupported-mode diagnostics = %v", diagnostics)
			}
			if mutations != 0 {
				t.Fatalf("unsupported mode made %d mutation requests", mutations)
			}
		})
	}
}

func TestAssignmentTransportFailureIsNotCompatibilityFailure(t *testing.T) {
	t.Parallel()
	client := assignmentTestClient(t, "http://assignment.invalid")
	client.HTTPClient().Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/api/core/firmware/info" {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"product_version":"26.7.6"}`)), Header: make(http.Header)}, nil
		}
		return nil, errors.New("temporary network failure")
	})
	r := &assignmentResource{client: client}
	_, err := r.readCompatibleAssignment(context.Background(), "opt3")
	var serverErr *opnsense.ServerError
	var compatibilityErr *assignmentCompatibilityError
	if !errors.As(err, &serverErr) || errors.As(err, &compatibilityErr) {
		t.Fatalf("transport error classification = %T: %v", err, err)
	}
}

func TestAssignmentAddressValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mode    types.String
		address types.String
		prefix  types.Int64
		gateway types.String
		ipv6    bool
		wantErr bool
	}{
		{name: "valid ipv4 static", mode: types.StringValue("staticv4"), address: types.StringValue("192.0.2.1"), prefix: types.Int64Value(24)},
		{name: "missing static prefix", mode: types.StringValue("staticv4"), address: types.StringValue("192.0.2.1"), prefix: types.Int64Null(), wantErr: true},
		{name: "wrong family", mode: types.StringValue("staticv4"), address: types.StringValue("2001:db8::1"), prefix: types.Int64Value(64), wantErr: true},
		{name: "static value with dhcp", mode: types.StringValue("dhcp"), address: types.StringValue("192.0.2.1"), prefix: types.Int64Null(), wantErr: true},
		{name: "static value without mode", mode: types.StringNull(), address: types.StringValue("192.0.2.1"), prefix: types.Int64Value(24), wantErr: true},
		{name: "valid ipv6 static", mode: types.StringValue("staticv6"), address: types.StringValue("2001:db8::1"), prefix: types.Int64Value(64), ipv6: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response resource.ValidateConfigResponse
			staticMode, label := "staticv4", "ipv4"
			if tt.ipv6 {
				staticMode, label = "staticv6", "ipv6"
			}
			validateAddressFamily(&response, tt.mode, staticMode, tt.address, tt.prefix, tt.gateway, label, tt.ipv6)
			if response.Diagnostics.HasError() != tt.wantErr {
				t.Fatalf("diagnostics = %v, want error %t", response.Diagnostics, tt.wantErr)
			}
		})
	}
}

func TestAssignmentConfigValidatorsAreWiredForBothAddressFamilies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		model AssignmentResourceModel
	}{
		{
			name:  "invalid IPv4 configuration",
			model: AssignmentResourceModel{Name: types.StringValue("opt3"), IPv4Mode: types.StringValue("staticv4"), IPv4Address: types.StringValue("2001:db8::1"), IPv4PrefixLength: types.Int64Value(24)},
		},
		{
			name:  "invalid IPv6 configuration",
			model: AssignmentResourceModel{Name: types.StringValue("opt3"), IPv6Mode: types.StringNull(), IPv6Address: types.StringValue("2001:db8::1"), IPv6PrefixLength: types.Int64Value(64)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &assignmentResource{}
			validators := r.ConfigValidators(context.Background())
			if len(validators) == 0 {
				t.Fatal("resource returned no config validators")
			}
			request := resource.ValidateConfigRequest{Config: assignmentTestConfig(t, r, tt.model)}
			var response resource.ValidateConfigResponse
			for _, configValidator := range validators {
				configValidator.ValidateResource(context.Background(), request, &response)
			}
			if !response.Diagnostics.HasError() {
				t.Fatal("resource validator accepted invalid configuration")
			}
		})
	}
}

func TestAssignmentDHCPFieldsRoundTrip(t *testing.T) {
	t.Parallel()
	fixtureJSON, err := json.Marshal(assignmentFixture("", ""))
	if err != nil {
		t.Fatal(err)
	}
	var current assignmentAPIResponse
	if err := json.Unmarshal(fixtureJSON, &current); err != nil {
		t.Fatal(err)
	}
	plan := AssignmentResourceModel{
		DHCPHostname:         types.StringValue("edge-test.example"),
		DHCPRejectFrom:       stringSet([]string{"192.0.2.20", "192.0.2.2"}),
		DHCPHonorMTU:         types.BoolValue(true),
		DHCPVLANPriority:     types.StringValue("3"),
		DHCPSendOptions:      types.StringValue("host-name=edge-test"),
		DHCPRequestOptions:   types.StringValue("subnet-mask,routers"),
		DHCP6VLANPriority:    types.StringValue("5"),
		DHCP6RequestDNS:      types.BoolValue(true),
		DHCP6InformationOnly: types.BoolValue(true),
		DHCP6RapidCommit:     types.BoolValue(true),
		DHCP6SendOptions:     types.StringValue("client-id=00:01"),
		DHCP6RequestOptions:  types.StringValue("dns-servers,domain-list"),
	}
	payload, err := plan.applyToAPI(context.Background(), &current)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(*payload)
	if err != nil {
		t.Fatal(err)
	}
	var response assignmentAPIResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	var got AssignmentResourceModel
	if err := got.fromAPI(context.Background(), &response, "opt3"); err != nil {
		t.Fatal(err)
	}
	var rejectFrom []string
	got.DHCPRejectFrom.ElementsAs(context.Background(), &rejectFrom, false)
	sort.Strings(rejectFrom)

	tests := []struct {
		name string
		api  string
		want string
		got  string
	}{
		{name: "hostname", api: "dhcphostname", want: "edge-test.example", got: got.DHCPHostname.ValueString()},
		{name: "reject from", api: "dhcprejectfrom", want: "192.0.2.2,192.0.2.20", got: strings.Join(rejectFrom, ",")},
		{name: "honor MTU", api: "dhcphonourmtu", want: "1", got: boolString(got.DHCPHonorMTU.ValueBool())},
		{name: "DHCPv4 VLAN", api: "dhcpvlanprio", want: "3", got: got.DHCPVLANPriority.ValueString()},
		{name: "DHCPv4 send", api: "dhcp_send_options", want: "host-name=edge-test", got: got.DHCPSendOptions.ValueString()},
		{name: "DHCPv4 request", api: "dhcp_request_options", want: "subnet-mask,routers", got: got.DHCPRequestOptions.ValueString()},
		{name: "DHCPv6 VLAN", api: "dhcp6vlanprio", want: "5", got: got.DHCP6VLANPriority.ValueString()},
		{name: "DHCPv6 DNS", api: "dhcp6_request_dns", want: "1", got: boolString(got.DHCP6RequestDNS.ValueBool())},
		{name: "DHCPv6 information only", api: "dhcp6-information-only", want: "1", got: boolString(got.DHCP6InformationOnly.ValueBool())},
		{name: "DHCPv6 rapid commit", api: "dhcp6_rapid_commit", want: "1", got: boolString(got.DHCP6RapidCommit.ValueBool())},
		{name: "DHCPv6 send", api: "dhcp6_send_options", want: "client-id=00:01", got: got.DHCP6SendOptions.ValueString()},
		{name: "DHCPv6 request", api: "dhcp6_request_options", want: "dns-servers,domain-list", got: got.DHCP6RequestOptions.ValueString()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if apiValue := fmt.Sprint((*payload)[tt.api]); apiValue != tt.want {
				t.Errorf("payload %s = %q, want %q", tt.api, apiValue, tt.want)
			}
			if tt.got != tt.want {
				t.Errorf("read-back %s = %q, want %q", tt.api, tt.got, tt.want)
			}
		})
	}
}

func TestAssignmentReadRemovesStateOnlyWhenAssignmentIsGone(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		assignment map[string]any
		wantGone   bool
		wantError  bool
	}{
		{name: "empty assignment", assignment: nil, wantGone: true},
		{name: "partial API response", assignment: map[string]any{"type4": selected("")}, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := assignmentTestServer(t, "26.7.6", tt.assignment, http.StatusOK, "")
			defer server.Close()
			r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
			state := assignmentTestState(t, r, AssignmentResourceModel{ID: types.StringValue("opt3"), Name: types.StringValue("opt3")})
			response := resource.ReadResponse{State: state}
			r.Read(context.Background(), resource.ReadRequest{State: state}, &response)
			if response.State.Raw.IsNull() != tt.wantGone {
				t.Fatalf("state null = %t, want %t", response.State.Raw.IsNull(), tt.wantGone)
			}
			if response.Diagnostics.HasError() != tt.wantError {
				t.Fatalf("read diagnostics = %v", response.Diagnostics)
			}
		})
	}
}

func TestAssignmentAddressModeWireFormat(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(assignmentFixture("192.0.2.10/24", "2001:db8::10/64"))
	if err != nil {
		t.Fatal(err)
	}
	var current assignmentAPIResponse
	if err := json.Unmarshal(raw, &current); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		plan         AssignmentResourceModel
		wantType4    string
		wantAddress4 string
		wantGateway4 string
		wantType6    string
		wantAddress6 string
		wantGateway6 string
	}{
		{
			name:      "dynamic IPv4 clears static values",
			plan:      AssignmentResourceModel{IPv4Mode: types.StringValue("dhcp")},
			wantType4: "dhcp", wantAddress4: "", wantGateway4: "",
			wantType6: "", wantAddress6: "2001:db8::10/64", wantGateway6: "",
		},
		{
			name:      "none modes use blank API options",
			plan:      AssignmentResourceModel{IPv4Mode: types.StringValue("none"), IPv6Mode: types.StringValue("none")},
			wantType4: "", wantAddress4: "", wantGateway4: "",
			wantType6: "", wantAddress6: "", wantGateway6: "",
		},
		{
			name: "static IPv6 joins address and prefix",
			plan: AssignmentResourceModel{
				IPv6Mode: types.StringValue("staticv6"), IPv6Address: types.StringValue("2001:db8::20"),
				IPv6PrefixLength: types.Int64Value(65), IPv6Gateway: types.StringValue("GW6_TEST"),
			},
			wantType4: "", wantAddress4: "192.0.2.10/24", wantGateway4: "",
			wantType6: "staticv6", wantAddress6: "2001:db8::20/65", wantGateway6: "GW6_TEST",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := tt.plan.applyToAPI(context.Background(), &current)
			if err != nil {
				t.Fatal(err)
			}
			for key, want := range map[string]string{
				"type4": tt.wantType4, "ipaddr": tt.wantAddress4, "gateway": tt.wantGateway4,
				"type6": tt.wantType6, "ipaddrv6": tt.wantAddress6, "gatewayv6": tt.wantGateway6,
			} {
				if got := fmt.Sprint((*payload)[key]); got != want {
					t.Errorf("%s = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestSplitAssignmentCIDR(t *testing.T) {
	t.Parallel()
	address, prefix, err := splitAssignmentCIDR("2001:db8::10/64", true)
	if err != nil {
		t.Fatal(err)
	}
	if address.ValueString() != "2001:db8::10" || prefix.ValueInt64() != 64 {
		t.Fatalf("split IPv6 CIDR = %s/%d", address.ValueString(), prefix.ValueInt64())
	}
}

func TestAssignmentDeleteIsStateOnly(t *testing.T) {
	t.Parallel()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer server.Close()
	r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
	r.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
	if calls != 0 {
		t.Fatalf("delete made %d API calls", calls)
	}
}

func TestAssignmentImportRejectsUnknownIdentity(t *testing.T) {
	t.Parallel()
	r := &assignmentResource{}
	var response resource.ImportStateResponse
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "em0"}, &response)
	if !response.Diagnostics.HasError() || !strings.Contains(response.Diagnostics[0].Detail(), "wan, lan, or optN") {
		t.Fatalf("import diagnostics = %v", response.Diagnostics)
	}
}

func TestVersionAtLeast(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		version string
		want    bool
	}{{"26.7.5", false}, {"26.7.6", true}, {"26.7.7-amd64", true}, {"27.1.a", true}, {"26.7", false}} {
		if got := versionAtLeast(tt.version, 26, 7, 6); got != tt.want {
			t.Errorf("versionAtLeast(%q) = %t, want %t", tt.version, got, tt.want)
		}
	}
}

func assignmentFixture(ipv4, ipv6 string) map[string]any {
	return map[string]any{
		"identifier": "opt3", "if": selected("vtnet2"), "descr": "disposable", "lock": "1", "enable": "1", "blockpriv": "1",
		"disablevlanhwfilter": []any{
			map[string]any{"value": "Enable VLAN Hardware Filtering", "selected": 1},
			map[string]any{"value": "Disable VLAN Hardware Filtering", "selected": 0},
			map[string]any{"value": "Leave default", "selected": 0},
		},
		"type4": selected(""), "ipaddr": ipv4, "gateway": []any{},
		"type6": selected(""), "ipaddrv6": ipv6, "gatewayv6": []any{},
		"dhcphostname": "", "dhcprejectfrom": "", "dhcphonourmtu": "0", "dhcpvlanprio": []any{},
		"dhcp_send_options": "", "dhcp_request_options": "", "dhcp6vlanprio": []any{},
		"dhcp6_request_dns": "0", "dhcp6-information-only": "0", "dhcp6_rapid_commit": "0",
		"dhcp6_send_options": "", "dhcp6_request_options": "",
	}
}

func selected(value string) map[string]any {
	return map[string]any{value: map[string]any{"value": value, "selected": 1}}
}

func assignmentTestClient(t *testing.T, baseURL string) *opnsense.Client {
	t.Helper()
	client, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: baseURL, APIKey: "test", APISecret: "test", RetryMax: 1})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func assignmentTestServer(t *testing.T, version string, assignment map[string]any, status int, mutationBody string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/core/firmware/info":
			_, _ = fmt.Fprintf(w, `{"product_version":%q}`, version)
		case "/api/interfaces/assignment/get_item/opt3":
			w.WriteHeader(status)
			if mutationBody != "" && assignment == nil {
				_, _ = io.WriteString(w, mutationBody)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"interface": assignment})
		case "/api/interfaces/assignment/set_item/opt3":
			_, _ = io.WriteString(w, mutationBody)
		case "/api/interfaces/assignment/reconfigure":
			_, _ = io.WriteString(w, `{"status":"ok"}`)
		default:
			http.NotFound(w, req)
		}
	}))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func assignmentTestSchema(t *testing.T, r *assignmentResource) resourceschema.Schema {
	t.Helper()
	var response resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", response.Diagnostics)
	}
	return response.Schema
}

func assignmentTestState(t *testing.T, r *assignmentResource, model AssignmentResourceModel) tfsdk.State {
	t.Helper()
	if model.DHCPRejectFrom.IsNull() {
		model.DHCPRejectFrom = types.SetNull(types.StringType)
	}
	state := tfsdk.State{Schema: assignmentTestSchema(t, r)}
	if diagnostics := state.Set(context.Background(), &model); diagnostics.HasError() {
		t.Fatalf("set test state: %v", diagnostics)
	}
	return state
}

func assignmentTestConfig(t *testing.T, r *assignmentResource, model AssignmentResourceModel) tfsdk.Config {
	t.Helper()
	state := assignmentTestState(t, r, model)
	return tfsdk.Config(state)
}

func boolString(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
