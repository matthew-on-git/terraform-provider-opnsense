// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	schemavalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDHCPv4AdvancedOptionsUseTaggedWireShape(t *testing.T) {
	routes := types.SetValueMust(types.ObjectType{AttrTypes: staticRouteAttrTypes}, []attr.Value{
		types.ObjectValueMust(staticRouteAttrTypes, map[string]attr.Value{"destination": types.StringValue("198.51.100.1"), "router": types.StringValue("192.0.2.2")}),
		types.ObjectValueMust(staticRouteAttrTypes, map[string]attr.Value{"destination": types.StringValue("203.0.113.1"), "router": types.StringValue("192.0.2.1")}),
	})
	model := DHCPv4SubnetResourceModel{
		Subnet: types.StringValue("192.0.2.0/24"), AutoCollect: types.BoolValue(false),
		DNSServers: stringSet("9.9.9.9", "1.1.1.1"), DomainName: types.StringValue("example.test"),
		DomainSearch: stringSet("b.example.test", "a.example.test"), Routers: stringSet("192.0.2.1"),
		StaticRoutes: routes, NTPServers: stringSet("192.0.2.20"), TimeServers: stringSet("192.0.2.21"),
		NextServer: types.StringValue("192.0.2.22"), TFTPServer: types.StringValue("tftp.example.test"),
		TFTPBootfile: types.StringValue("bootx64.efi"),
	}
	body, err := json.Marshal(map[string]any{"subnet4": model.toAPI(context.Background())})
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, fragment := range []string{
		`"option_data_autocollect":"0"`, `"next_server":"192.0.2.22"`,
		`"domain_name_servers":"1.1.1.1,9.9.9.9"`, `"domain_search":"a.example.test,b.example.test"`,
		`"static_routes":"198.51.100.1,192.0.2.2,203.0.113.1,192.0.2.1"`,
		`"tftp_server_name":"tftp.example.test"`, `"boot_file_name":"bootx64.efi"`,
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("request %s lacks %s", text, fragment)
		}
	}
	if strings.Contains(text, ";") {
		t.Fatalf("static routes used deprecated semicolon encoding: %s", text)
	}
}

func TestDHCPv4AdvancedOptionsDecodeSelectedMapsAndRejectMalformedRoutes(t *testing.T) {
	raw := []byte(`{"subnet":"192.0.2.0/24","match-client-id":"1","ping_check":"0","option_data_autocollect":"0","option_data":{"domain_name_servers":{"9.9.9.9":{"selected":1},"1.1.1.1":{"selected":"1"}},"domain_search":{"example.test":{"selected":1}},"routers":{"192.0.2.1":{"selected":1}},"static_routes":"198.51.100.1,192.0.2.2,203.0.113.1,192.0.2.1","ntp_servers":{"192.0.2.20":{"selected":1}},"time_servers":{"192.0.2.21":{"selected":1}}}}`)
	var response dhcpv4SubnetAPIResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	var model DHCPv4SubnetResourceModel
	if err := model.fromAPI(context.Background(), &response, "id"); err != nil {
		t.Fatal(err)
	}
	if len(model.DNSServers.Elements()) != 2 || len(model.StaticRoutes.Elements()) != 2 || model.AutoCollect.ValueBool() {
		t.Fatalf("unexpected decoded model: %#v", model)
	}
	response.OptionData.StaticRoutes = "198.51.100.1,192.0.2.2,orphan"
	if err := model.fromAPI(context.Background(), &response, "id"); err == nil || !strings.Contains(err.Error(), "destination/router pairs") {
		t.Fatalf("expected malformed API diagnostic, got %v", err)
	}
	response.OptionData.StaticRoutes = ""
	response.AutoCollect = "2"
	if err := model.fromAPI(context.Background(), &response, "id"); err == nil || !strings.Contains(err.Error(), "expected 0 or 1") {
		t.Fatalf("expected malformed auto-collect diagnostic, got %v", err)
	}
}

func TestDHCPv6OptionSetsPreserveUnknownAndDecodeSelectedMaps(t *testing.T) {
	model := DHCPv6SubnetResourceModel{DNSServers: types.SetUnknown(types.StringType), DomainSearch: types.SetNull(types.StringType)}
	body, err := json.Marshal(model.toAPI(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "dns_servers") || strings.Contains(string(body), "domain_search") {
		t.Fatalf("unknown/null values emitted as clears: %s", body)
	}
	var response dhcpv6SubnetAPIResponse
	if err := json.Unmarshal([]byte(`{"subnet":"2001:db8::/64","interface":{"opt1":{"selected":1}},"option_data":{"dns_servers":{"2001:db8::53":{"selected":1}},"domain_search":{"example.test":{"selected":1}}}}`), &response); err != nil {
		t.Fatal(err)
	}
	if err := model.fromAPI(context.Background(), &response, "id"); err != nil {
		t.Fatal(err)
	}
	if len(model.DNSServers.Elements()) != 1 || len(model.DomainSearch.Elements()) != 1 {
		t.Fatalf("selected-map lists were not preserved: %#v", model)
	}
	if err := json.Unmarshal([]byte(`{"option_data":{"dns_servers":42}}`), &response); err == nil || !strings.Contains(err.Error(), "selected-map list") {
		t.Fatalf("malformed API list shape did not produce a diagnostic: %v", err)
	}
}

func TestSubnetOptionRequestsDistinguishExplicitEmptyFromUnknown(t *testing.T) {
	emptyStrings := stringSet()
	emptyRoutes := types.SetValueMust(types.ObjectType{AttrTypes: staticRouteAttrTypes}, []attr.Value{})
	v4 := DHCPv4SubnetResourceModel{
		Subnet: types.StringValue("192.0.2.0/24"), MatchClientID: types.BoolValue(true), PingCheck: types.BoolValue(false), AutoCollect: types.BoolValue(false),
		DNSServers: emptyStrings, DomainName: types.StringValue(""), DomainSearch: types.SetUnknown(types.StringType), Routers: emptyStrings,
		StaticRoutes: emptyRoutes, NTPServers: types.SetNull(types.StringType), TimeServers: emptyStrings, NextServer: types.StringValue(""),
		TFTPServer: types.StringUnknown(), TFTPBootfile: types.StringValue(""),
	}
	body, err := json.Marshal(v4.toAPI(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, fragment := range []string{`"domain_name_servers":""`, `"domain_name":""`, `"routers":""`, `"static_routes":""`, `"time_servers":""`, `"next_server":""`, `"boot_file_name":""`} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("explicit empty value %s was omitted from %s", fragment, text)
		}
	}
	for _, key := range []string{"domain_search", "ntp_servers", "tftp_server_name"} {
		if strings.Contains(text, `"`+key+`"`) {
			t.Fatalf("unknown/null field %s was emitted in %s", key, text)
		}
	}

	v6 := DHCPv6SubnetResourceModel{DNSServers: emptyStrings, DomainSearch: types.SetUnknown(types.StringType)}
	body, err = json.Marshal(v6.toAPI(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"dns_servers":""`) || strings.Contains(string(body), `"domain_search"`) {
		t.Fatalf("DHCPv6 empty/unknown distinction was lost: %s", body)
	}
}

func TestSubnetRequestsOmitAllUnknownAndNullScalars(t *testing.T) {
	v4 := DHCPv4SubnetResourceModel{
		Subnet: types.StringUnknown(), Allocator: types.StringNull(), Pools: types.StringUnknown(),
		MatchClientID: types.BoolUnknown(), PingCheck: types.BoolNull(), AutoCollect: types.BoolUnknown(),
		DNSServers: types.SetUnknown(types.StringType), DomainSearch: types.SetNull(types.StringType),
		StaticRoutes: types.SetUnknown(types.ObjectType{AttrTypes: staticRouteAttrTypes}), Description: types.StringNull(),
	}
	body, err := json.Marshal(v4.toAPI(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"subnet", "allocator", "pools", "match-client-id", "ping_check", "option_data_autocollect", "domain_name_servers", "domain_search", "static_routes", "description"} {
		if strings.Contains(string(body), `"`+key+`"`) {
			t.Fatalf("unknown/null field %s emitted in %s", key, body)
		}
	}

	v6 := DHCPv6SubnetResourceModel{
		Subnet: types.StringUnknown(), Interface: types.StringNull(), Allocator: types.StringUnknown(),
		PDAllocator: types.StringNull(), Pools: types.StringUnknown(), DNSServers: types.SetUnknown(types.StringType),
		DomainSearch: types.SetNull(types.StringType), Description: types.StringUnknown(),
	}
	body, err = json.Marshal(v6.toAPI(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"subnet", "interface", "allocator", "pd-allocator", "pools", "dns_servers", "domain_search", "description"} {
		if strings.Contains(string(body), `"`+key+`"`) {
			t.Fatalf("unknown/null field %s emitted in %s", key, body)
		}
	}
}

func TestPartiallyUnknownCollectionsAreOmitted(t *testing.T) {
	partialStrings := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("192.0.2.1"), types.StringUnknown()})
	partialRoutes := types.SetValueMust(types.ObjectType{AttrTypes: staticRouteAttrTypes}, []attr.Value{
		types.ObjectValueMust(staticRouteAttrTypes, map[string]attr.Value{"destination": types.StringValue("198.51.100.1"), "router": types.StringUnknown()}),
	})
	model := DHCPv4SubnetResourceModel{DNSServers: partialStrings, StaticRoutes: partialRoutes}
	body, err := json.Marshal(model.toAPI(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "domain_name_servers") || strings.Contains(string(body), "static_routes") {
		t.Fatalf("partial unknown became a clear/value: %s", body)
	}
}

func TestPartiallyUnknownCollectionsDeferWithoutValidationDiagnostics(t *testing.T) {
	partialAddresses := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("192.0.2.1"), types.StringUnknown()})
	partialDomains := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("example.test"), types.StringUnknown()})
	partialRoutes := types.SetValueMust(types.ObjectType{AttrTypes: staticRouteAttrTypes}, []attr.Value{
		types.ObjectValueMust(staticRouteAttrTypes, map[string]attr.Value{"destination": types.StringValue("198.51.100.1"), "router": types.StringUnknown()}),
	})
	var response resource.ValidateConfigResponse
	validateDHCPv4Subnet(context.Background(), DHCPv4SubnetResourceModel{AutoCollect: types.BoolValue(true), DNSServers: partialAddresses, DomainSearch: partialDomains, StaticRoutes: partialRoutes}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("partially unknown v4 collections produced diagnostics: %v", response.Diagnostics)
	}

	response = resource.ValidateConfigResponse{}
	partialV6 := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("2001:db8::53"), types.StringUnknown()})
	validateDHCPv6Subnet(context.Background(), DHCPv6SubnetResourceModel{DNSServers: partialV6, DomainSearch: partialDomains, Interface: types.StringUnknown()}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("partially unknown v6 collections produced diagnostics: %v", response.Diagnostics)
	}

	response = resource.ValidateConfigResponse{}
	invalidKnown := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("not-an-address"), types.StringUnknown()})
	validateDHCPv4Subnet(context.Background(), DHCPv4SubnetResourceModel{DNSServers: invalidKnown}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("known invalid member beside unknown member was not validated")
	}
}

func TestPDDescriptionKnownEmptyClearsAndUnknownOmits(t *testing.T) {
	known := (&DHCPv6PDPoolResourceModel{Description: types.StringValue("")}).toAPI(context.Background())
	body, err := json.Marshal(known)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"description":""`) {
		t.Fatalf("known empty description omitted: %s", body)
	}
	unknown := (&DHCPv6PDPoolResourceModel{Description: types.StringUnknown()}).toAPI(context.Background())
	body, err = json.Marshal(unknown)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "description") {
		t.Fatalf("unknown description emitted: %s", body)
	}
}

func TestStrictKeaSelectedMapsRejectMissingMultipleAndMalformedFlags(t *testing.T) {
	for _, raw := range []string{
		`{}`,
		`{"primary":{"selected":1},"standby":{"selected":"1"}}`,
		`{"primary":{}}`,
		`{"primary":{"selected":2}}`,
		`"primary"`,
	} {
		var selected strictSelectedMap
		if err := json.Unmarshal([]byte(raw), &selected); err == nil {
			t.Fatalf("strict selected map accepted %s", raw)
		}
	}
	var selected strictSelectedMap
	if err := json.Unmarshal([]byte(`{"primary":{"selected":1},"standby":{"selected":0}}`), &selected); err != nil || selected != "primary" {
		t.Fatalf("valid selected map decoded as %q: %v", selected, err)
	}
}

func TestStrictSelectedMapRejectsMixedEmptySentinelDeterministically(t *testing.T) {
	for i := 0; i < 100; i++ {
		for _, raw := range []string{
			`{"":{"selected":1},"lan":{"selected":1}}`,
			`{"lan":{"selected":1},"":{"selected":1}}`,
		} {
			var selected strictSelectedMap
			if err := json.Unmarshal([]byte(raw), &selected); err == nil {
				t.Fatalf("iteration %d: accepted multiple required selections in %s", i, raw)
			}
		}
	}
}

func TestDHCPv6PeerResponseRejectsUnsupportedSelectedRole(t *testing.T) {
	var response dhcpv6PeerAPIResponse
	if err := json.Unmarshal([]byte(`{"name":"peer","role":{"secondary":{"selected":1}},"url":"http://[2001:db8::1]:8000/"}`), &response); err != nil {
		t.Fatal(err)
	}
	var model DHCPv6PeerResourceModel
	if err := model.fromAPI(context.Background(), &response, "peer-id"); err == nil || !strings.Contains(err.Error(), "primary or standby") {
		t.Fatalf("expected unsupported response-role error, got %v", err)
	}
}

func TestPeerResponseRejectsEmptyNameAndInvalidURL(t *testing.T) {
	for _, model := range []DHCPv6PeerResourceModel{
		{Name: types.StringValue(""), Role: types.StringValue("primary"), URL: types.StringValue("http://[2001:db8::1]:8000/")},
		{Name: types.StringValue("peer"), Role: types.StringValue("primary"), URL: types.StringValue("ftp://example.test/")},
	} {
		if err := validateDHCPv6PeerResponse(&model); err == nil {
			t.Fatalf("invalid peer response accepted: %#v", model)
		}
	}
}

func TestDHCPv4AllocatorResponseIsStrictOptionalAndSupported(t *testing.T) {
	for _, raw := range []string{`{"a":{"selected":1},"b":{"selected":1}}`, `{"a":{"selected":2}}`} {
		var response dhcpv4SubnetAPIResponse
		if err := json.Unmarshal([]byte(fmt.Sprintf(`{"allocator":%s}`, raw)), &response); err == nil {
			t.Fatalf("malformed allocator accepted: %s", raw)
		}
	}
	model := DHCPv4SubnetResourceModel{Allocator: types.StringValue("unsupported"), Subnet: types.StringValue("192.0.2.0/24")}
	if err := validateDHCPv4SubnetResponse(&model); err == nil {
		t.Fatal("unsupported allocator accepted")
	}
}

func TestSubnetAllocatorsValidateBeforeMutationAndOnResponse(t *testing.T) {
	for _, allocator := range []string{"", "iterative", "random", "flq"} {
		var response resource.ValidateConfigResponse
		validateDHCPv4Subnet(context.Background(), DHCPv4SubnetResourceModel{Allocator: types.StringValue(allocator)}, &response)
		if response.Diagnostics.HasError() {
			t.Fatalf("valid v4 allocator %q rejected: %v", allocator, response.Diagnostics)
		}
	}
	var response resource.ValidateConfigResponse
	validateDHCPv4Subnet(context.Background(), DHCPv4SubnetResourceModel{Allocator: types.StringValue("unsupported")}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("unsupported v4 allocator accepted before mutation")
	}

	for field, values := range map[string][]string{
		"allocator":    {"", "iterative", "random"},
		"pd_allocator": {"", "iterative", "random", "flq"},
	} {
		for _, value := range values {
			model := DHCPv6SubnetResourceModel{Interface: types.StringValue("opt1"), Allocator: types.StringValue(""), PDAllocator: types.StringValue("")}
			if field == "allocator" {
				model.Allocator = types.StringValue(value)
			} else {
				model.PDAllocator = types.StringValue(value)
			}
			response = resource.ValidateConfigResponse{}
			validateDHCPv6Subnet(context.Background(), model, &response)
			if response.Diagnostics.HasError() {
				t.Fatalf("valid v6 %s %q rejected: %v", field, value, response.Diagnostics)
			}
		}
	}
	for field, mutate := range map[string]func(*DHCPv6SubnetResourceModel){
		"allocator":    func(model *DHCPv6SubnetResourceModel) { model.Allocator = types.StringValue("flq") },
		"pd_allocator": func(model *DHCPv6SubnetResourceModel) { model.PDAllocator = types.StringValue("unsupported") },
	} {
		model := DHCPv6SubnetResourceModel{Interface: types.StringValue("opt1"), Allocator: types.StringValue(""), PDAllocator: types.StringValue("")}
		mutate(&model)
		response = resource.ValidateConfigResponse{}
		validateDHCPv6Subnet(context.Background(), model, &response)
		if !response.Diagnostics.HasError() {
			t.Fatalf("unsupported v6 %s accepted before mutation", field)
		}
		if err := validateDHCPv6SubnetResponse(&model); err == nil {
			t.Fatalf("unsupported v6 %s accepted on response", field)
		}
	}
}

func TestRequiredResponseRelationshipsRejectEmptyValues(t *testing.T) {
	v6 := DHCPv6SubnetResourceModel{Subnet: types.StringValue("2001:db8::/64"), Interface: types.StringValue(""), Allocator: types.StringValue(""), PDAllocator: types.StringValue(""), DNSServers: stringSet(), DomainSearch: stringSet()}
	if err := validateDHCPv6SubnetResponse(&v6); err == nil || !strings.Contains(err.Error(), "interface") {
		t.Fatalf("empty interface response accepted: %v", err)
	}
	pd := DHCPv6PDPoolResourceModel{SubnetID: types.StringValue(""), Prefix: types.StringValue("2001:db8:31::"), PrefixLength: types.Int64Value(48), DelegatedLength: types.Int64Value(64)}
	if err := validateDHCPv6PDPoolResponse(&pd); err == nil || !strings.Contains(err.Error(), "subnet_id") {
		t.Fatalf("empty parent response accepted: %v", err)
	}
}

func TestStaticRouteResponseRequiresNonEmptyIPv4Members(t *testing.T) {
	for field, values := range map[string]map[string]attr.Value{
		"destination": {"destination": types.StringValue(""), "router": types.StringValue("192.0.2.1")},
		"router":      {"destination": types.StringValue("198.51.100.1"), "router": types.StringValue("")},
	} {
		model := DHCPv4SubnetResourceModel{
			Subnet: types.StringValue("192.0.2.0/24"), Allocator: types.StringValue(""), DNSServers: stringSet(), DomainSearch: stringSet(), Routers: stringSet(),
			StaticRoutes: types.SetValueMust(types.ObjectType{AttrTypes: staticRouteAttrTypes}, []attr.Value{types.ObjectValueMust(staticRouteAttrTypes, values)}),
			NTPServers:   stringSet(), TimeServers: stringSet(),
		}
		if err := validateDHCPv4SubnetResponse(&model); err == nil || !strings.Contains(err.Error(), field) {
			t.Fatalf("empty %s response accepted: %v", field, err)
		}
	}
}

func TestDHCPv6PeerRoleIsOptionalWithPrimaryDefault(t *testing.T) {
	var response resource.SchemaResponse
	(&dhcpv6PeerResource{}).Schema(context.Background(), resource.SchemaRequest{}, &response)
	role, ok := response.Schema.Attributes["role"].(resourceschema.StringAttribute)
	if !ok || !role.Optional || !role.Computed || role.Default == nil {
		t.Fatalf("role schema does not expose optional computed default: %#v", response.Schema.Attributes["role"])
	}
}

func TestKeaOptionListRejectsMissingAndNonBinarySelectedFlags(t *testing.T) {
	for _, raw := range []string{
		`{"one":{}}`, `{"one":{"selected":2}}`, `{"one":{"selected":"yes"}}`,
		`null`, `[]`, `["one"]`, `"one,two"`, `""`,
	} {
		var values keaOptionList
		if err := json.Unmarshal([]byte(raw), &values); err == nil {
			t.Fatalf("option list accepted %s", raw)
		}
	}
	for _, raw := range []string{`{}`, `{"":{"selected":1}}`} {
		var values keaOptionList
		if err := json.Unmarshal([]byte(raw), &values); err != nil || len(values) != 0 {
			t.Fatalf("explicit empty selected map %s: values=%#v err=%v", raw, values, err)
		}
	}
}

func TestKeaOptionListAcceptsOPNsenseEmptySentinel(t *testing.T) {
	var values keaOptionList
	if err := json.Unmarshal([]byte(`{"":{"selected":1}}`), &values); err != nil {
		t.Fatalf("decode empty sentinel: %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("decoded values = %#v, want empty", values)
	}

	for i := 0; i < 100; i++ {
		for _, raw := range []string{`{"":{"selected":1},"192.0.2.53":{"selected":1}}`, `{"192.0.2.53":{"selected":1},"":{"selected":1}}`} {
			if err := json.Unmarshal([]byte(raw), &values); err == nil {
				t.Fatalf("iteration %d: expected mixed empty sentinel and selected value to fail", i)
			}
		}
	}
}

func TestStrictOptionalSelectedMapRejectsMixedEmptySentinelDeterministically(t *testing.T) {
	for i := 0; i < 100; i++ {
		for _, raw := range []string{
			`{"":{"selected":1},"random":{"selected":1}}`,
			`{"random":{"selected":1},"":{"selected":1}}`,
		} {
			var value strictOptionalSelectedMap
			if err := json.Unmarshal([]byte(raw), &value); err == nil {
				t.Fatalf("iteration %d: accepted multiple selections in %s", i, raw)
			}
		}
	}
}

func TestSubnetResponsesRequireOptionDataAndSemanticValues(t *testing.T) {
	for _, raw := range []string{`{"subnet":"192.0.2.0/24","match-client-id":"1","ping_check":"0","option_data_autocollect":"0"}`, `{"subnet":"192.0.2.0/24","match-client-id":"1","ping_check":"0","option_data_autocollect":"0","option_data":null}`} {
		var response dhcpv4SubnetAPIResponse
		if err := json.Unmarshal([]byte(raw), &response); err != nil {
			t.Fatal(err)
		}
		if err := (&DHCPv4SubnetResourceModel{}).fromAPI(context.Background(), &response, "id"); err == nil {
			t.Fatalf("accepted response without option_data: %s", raw)
		}
	}
	var v4 dhcpv4SubnetAPIResponse
	if err := json.Unmarshal([]byte(`{"subnet":"192.0.2.0/24","match-client-id":"1","ping_check":"0","option_data_autocollect":"0","option_data":{"domain_name_servers":{"not-an-ip":{"selected":1}},"domain_search":{},"routers":{},"ntp_servers":{},"time_servers":{}}}`), &v4); err != nil {
		t.Fatal(err)
	}
	if err := (&DHCPv4SubnetResourceModel{}).fromAPI(context.Background(), &v4, "id"); err == nil || !strings.Contains(err.Error(), "dns_servers") {
		t.Fatalf("invalid v4 response was accepted: %v", err)
	}

	for _, raw := range []string{
		`{"subnet":"2001:db8::/64","interface":{},"option_data":{}}`,
		`{"subnet":"2001:db8::/64","interface":{"one":{"selected":1},"two":{"selected":1}},"option_data":{}}`,
		`{"subnet":"2001:db8::/64","interface":{"one":{"selected":1}},"allocator":{"a":{"selected":1},"b":{"selected":1}},"option_data":{}}`,
	} {
		var response dhcpv6SubnetAPIResponse
		if err := json.Unmarshal([]byte(raw), &response); err == nil {
			t.Fatalf("accepted malformed v6 selected map: %s", raw)
		}
	}
	var v6 dhcpv6SubnetAPIResponse
	if err := json.Unmarshal([]byte(`{"subnet":"2001:db8::/64","interface":{"opt1":{"selected":1}},"allocator":{},"pd-allocator":null,"option_data":{"dns_servers":{"192.0.2.1":{"selected":1}},"domain_search":{}}}`), &v6); err != nil {
		t.Fatal(err)
	}
	if err := (&DHCPv6SubnetResourceModel{}).fromAPI(context.Background(), &v6, "id"); err == nil || !strings.Contains(err.Error(), "dns_servers") {
		t.Fatalf("invalid v6 response was accepted: %v", err)
	}
}

func TestDHCPv6PDPoolResponseValidation(t *testing.T) {
	var response dhcpv6PdPoolAPIResponse
	if err := json.Unmarshal([]byte(`{"subnet":{"parent":{"selected":1}},"prefix":"2001:db8:31::/48","prefix_len":"48","delegated_len":"64","description":"ok"}`), &response); err != nil {
		t.Fatal(err)
	}
	var model DHCPv6PDPoolResourceModel
	if err := model.fromAPI(context.Background(), &response, "pd-id"); err != nil {
		t.Fatal(err)
	}
	if model.PrefixLength.ValueInt64() != 48 || model.DelegatedLength.ValueInt64() != 64 || model.SubnetID.ValueString() != "parent" {
		t.Fatalf("unexpected model: %#v", model)
	}
	response.Prefix = "2001:db8:31::1"
	if err := model.fromAPI(context.Background(), &response, "pd-id"); err == nil || !strings.Contains(err.Error(), "aligned") {
		t.Fatalf("expected response alignment error, got %v", err)
	}
	response.Prefix = "2001:db8:31::/48"
	response.DelegatedLength = "48"
	if err := model.fromAPI(context.Background(), &response, "pd-id"); err == nil || !strings.Contains(err.Error(), "greater") {
		t.Fatalf("expected response ordering error, got %v", err)
	}
}

func TestDHCPv6PDPoolCIDRUsesAddressWireValueAndStableState(t *testing.T) {
	model := DHCPv6PDPoolResourceModel{
		SubnetID: types.StringValue("parent"), Prefix: types.StringValue("2001:db8:31::/48"),
		PrefixLength: types.Int64Value(48), DelegatedLength: types.Int64Value(64), Description: types.StringValue("test"),
	}
	if got := model.toAPI(context.Background()).Prefix; got != "2001:db8:31::" {
		t.Fatalf("wire prefix = %q, want address-only form", got)
	}
	response := dhcpv6PdPoolAPIResponse{
		SubnetID: strictSelectedMap("parent"), Prefix: "2001:db8:31::",
		PrefixLength: "48", DelegatedLength: "64", Description: "test",
	}
	if err := model.fromAPI(context.Background(), &response, "pd-id"); err != nil {
		t.Fatal(err)
	}
	if got := model.Prefix.ValueString(); got != "2001:db8:31::/48" {
		t.Fatalf("state prefix = %q, want configured CIDR spelling", got)
	}
	var imported DHCPv6PDPoolResourceModel
	if err := imported.fromAPIResource(context.Background(), &response, "pd-id"); err != nil {
		t.Fatal(err)
	}
	if got := imported.Prefix.ValueString(); got != "2001:db8:31::/48" {
		t.Fatalf("imported resource prefix = %q, want reconstructed CIDR", got)
	}
	var dataSource DHCPv6PDPoolResourceModel
	if err := dataSource.fromAPI(context.Background(), &response, "pd-id"); err != nil {
		t.Fatal(err)
	}
	if got := dataSource.Prefix.ValueString(); got != "2001:db8:31::" {
		t.Fatalf("data source prefix = %q, want API address", got)
	}
}

func TestKeaParityValidators(t *testing.T) {
	valid := DHCPv6PDPoolResourceModel{SubnetID: types.StringValue("parent"), Prefix: types.StringValue("2001:db8:31::"), PrefixLength: types.Int64Value(48), DelegatedLength: types.Int64Value(64), Description: types.StringValue("")}
	for _, test := range []struct {
		name  string
		model DHCPv6PDPoolResourceModel
		bad   bool
	}{
		{name: "valid", model: valid},
		{name: "valid CIDR", model: func() DHCPv6PDPoolResourceModel {
			m := valid
			m.Prefix = types.StringValue("2001:db8:31::/48")
			return m
		}()},
		{name: "unaligned", model: func() DHCPv6PDPoolResourceModel { m := valid; m.Prefix = types.StringValue("2001:db8:31::1"); return m }(), bad: true},
		{name: "IPv4", model: func() DHCPv6PDPoolResourceModel { m := valid; m.Prefix = types.StringValue("192.0.2.0"); return m }(), bad: true},
		{name: "CIDR mismatch", model: func() DHCPv6PDPoolResourceModel {
			m := valid
			m.Prefix = types.StringValue("2001:db8:31::/64")
			return m
		}(), bad: true},
		{name: "empty parent", model: func() DHCPv6PDPoolResourceModel { m := valid; m.SubnetID = types.StringValue(" "); return m }(), bad: true},
		{name: "prefix length low", model: func() DHCPv6PDPoolResourceModel { m := valid; m.PrefixLength = types.Int64Value(0); return m }(), bad: true},
		{name: "prefix length high", model: func() DHCPv6PDPoolResourceModel { m := valid; m.PrefixLength = types.Int64Value(129); return m }(), bad: true},
		{name: "delegation not longer", model: func() DHCPv6PDPoolResourceModel { m := valid; m.DelegatedLength = types.Int64Value(48); return m }(), bad: true},
		{name: "delegation high", model: func() DHCPv6PDPoolResourceModel { m := valid; m.DelegatedLength = types.Int64Value(129); return m }(), bad: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			owner := &dhcpv6PdPoolResource{}
			state := keaResourceState(t, owner, test.model)
			var response resource.ValidateConfigResponse
			owner.ConfigValidators(context.Background())[0].ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config(state)}, &response)
			if response.Diagnostics.HasError() != test.bad {
				t.Fatalf("diagnostics=%v want error=%t", response.Diagnostics, test.bad)
			}
		})
	}
}

func TestDHCPv4SubnetNegativeValidators(t *testing.T) {
	emptyStrings := stringSet()
	emptyRoutes := types.SetValueMust(types.ObjectType{AttrTypes: staticRouteAttrTypes}, []attr.Value{})
	valid := DHCPv4SubnetResourceModel{
		Subnet: types.StringValue("192.0.2.0/24"), MatchClientID: types.BoolValue(true), PingCheck: types.BoolValue(false), AutoCollect: types.BoolValue(false),
		DNSServers: emptyStrings, DomainName: types.StringValue("example.test"), DomainSearch: emptyStrings, Routers: emptyStrings,
		StaticRoutes: emptyRoutes, NTPServers: emptyStrings, TimeServers: emptyStrings, NextServer: types.StringValue(""),
		TFTPServer: types.StringValue("tftp.example.test"), TFTPBootfile: types.StringValue("bootx64.efi"),
	}
	invalidRoute := func(destination, router string) types.Set {
		return types.SetValueMust(types.ObjectType{AttrTypes: staticRouteAttrTypes}, []attr.Value{types.ObjectValueMust(staticRouteAttrTypes, map[string]attr.Value{"destination": types.StringValue(destination), "router": types.StringValue(router)})})
	}
	for _, test := range []struct {
		name   string
		mutate func(*DHCPv4SubnetResourceModel)
	}{
		{"subnet", func(m *DHCPv4SubnetResourceModel) { m.Subnet = types.StringValue("192.0.2.1/24") }},
		{"dns", func(m *DHCPv4SubnetResourceModel) { m.DNSServers = stringSet("2001:db8::53") }},
		{"router", func(m *DHCPv4SubnetResourceModel) { m.Routers = stringSet("bad") }},
		{"ntp", func(m *DHCPv4SubnetResourceModel) { m.NTPServers = stringSet("bad") }},
		{"time", func(m *DHCPv4SubnetResourceModel) { m.TimeServers = stringSet("bad") }},
		{"domain name", func(m *DHCPv4SubnetResourceModel) { m.DomainName = types.StringValue("bad_domain") }},
		{"domain search", func(m *DHCPv4SubnetResourceModel) { m.DomainSearch = stringSet("bad_domain") }},
		{"next server", func(m *DHCPv4SubnetResourceModel) { m.NextServer = types.StringValue("bad") }},
		{"static destination", func(m *DHCPv4SubnetResourceModel) { m.StaticRoutes = invalidRoute("bad", "192.0.2.1") }},
		{"static router", func(m *DHCPv4SubnetResourceModel) { m.StaticRoutes = invalidRoute("198.51.100.1", "bad") }},
		{"empty static destination", func(m *DHCPv4SubnetResourceModel) { m.StaticRoutes = invalidRoute("", "192.0.2.1") }},
		{"empty static router", func(m *DHCPv4SubnetResourceModel) { m.StaticRoutes = invalidRoute("198.51.100.1", "") }},
		{"auto collect conflict", func(m *DHCPv4SubnetResourceModel) {
			m.AutoCollect = types.BoolValue(true)
			m.DNSServers = stringSet("192.0.2.53")
		}},
		{"TFTP quote", func(m *DHCPv4SubnetResourceModel) { m.TFTPServer = types.StringValue(`bad"name`) }},
		{"bootfile newline", func(m *DHCPv4SubnetResourceModel) { m.TFTPBootfile = types.StringValue("bad\nname") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := valid
			test.mutate(&model)
			if !keaValidationHasError(t, &dhcpv4SubnetResource{}, model) {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDHCPv6SubnetAndPeerNegativeValidators(t *testing.T) {
	validSubnet := DHCPv6SubnetResourceModel{Subnet: types.StringValue("2001:db8::/64"), Interface: types.StringValue("opt1"), DNSServers: stringSet(), DomainSearch: stringSet()}
	for _, test := range []struct {
		name   string
		mutate func(*DHCPv6SubnetResourceModel)
	}{
		{"subnet family", func(m *DHCPv6SubnetResourceModel) { m.Subnet = types.StringValue("192.0.2.0/24") }},
		{"unaligned subnet", func(m *DHCPv6SubnetResourceModel) { m.Subnet = types.StringValue("2001:db8::1/64") }},
		{"DNS family", func(m *DHCPv6SubnetResourceModel) { m.DNSServers = stringSet("192.0.2.53") }},
		{"domain", func(m *DHCPv6SubnetResourceModel) { m.DomainSearch = stringSet("bad_domain") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := validSubnet
			test.mutate(&model)
			if !keaValidationHasError(t, &dhcpv6SubnetResource{}, model) {
				t.Fatal("expected validation error")
			}
		})
	}
	for _, rawURL := range []string{"relative/path", "ftp://example.test/", "://bad"} {
		model := DHCPv6PeerResourceModel{Name: types.StringValue("peer"), Role: types.StringValue("primary"), URL: types.StringValue(rawURL)}
		if !keaValidationHasError(t, &dhcpv6PeerResource{}, model) {
			t.Fatalf("expected URL validation error for %q", rawURL)
		}
	}
}

func TestDHCPv6PeerRoleValidatorRejectsUnsupportedRole(t *testing.T) {
	var schemaResponse resource.SchemaResponse
	(&dhcpv6PeerResource{}).Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	attribute, ok := schemaResponse.Schema.Attributes["role"].(resourceschema.StringAttribute)
	if !ok || len(attribute.Validators) == 0 {
		t.Fatal("role schema lacks a string validator")
	}
	var response schemavalidator.StringResponse
	attribute.Validators[0].ValidateString(context.Background(), schemavalidator.StringRequest{Path: path.Root("role"), ConfigValue: types.StringValue("secondary")}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("unsupported DHCPv6 peer role was accepted")
	}
}

func keaValidationHasError(t *testing.T, owner resource.ResourceWithConfigValidators, model any) bool {
	t.Helper()
	state := keaResourceState(t, owner, model)
	var response resource.ValidateConfigResponse
	owner.ConfigValidators(context.Background())[0].ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config(state)}, &response)
	return response.Diagnostics.HasError()
}

func TestKeaRegistrationCountsAndNoDeprecatedAliases(t *testing.T) {
	if got := len(Resources()); got != 12 {
		t.Fatalf("resources=%d want 12", got)
	}
	if got := len(DataSources()); got != 11 {
		t.Fatalf("data sources=%d want 11", got)
	}
	for _, name := range []string{"kea_peer", "kea_reservation", "kea_subnet"} {
		for _, constructor := range Resources() {
			var metadata resource.MetadataResponse
			constructor().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "opnsense"}, &metadata)
			if metadata.TypeName == "opnsense_"+name {
				t.Fatalf("deprecated generic alias %s is registered", metadata.TypeName)
			}
		}
		for _, constructor := range DataSources() {
			var metadata datasource.MetadataResponse
			constructor().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "opnsense"}, &metadata)
			if metadata.TypeName == "opnsense_"+name {
				t.Fatalf("deprecated generic data-source alias %s is registered", metadata.TypeName)
			}
		}
	}
}

func stringSet(values ...string) types.Set {
	items := make([]attr.Value, len(values))
	for i, value := range values {
		items[i] = types.StringValue(value)
	}
	return types.SetValueMust(types.StringType, items)
}

func keaResourceState(t *testing.T, owner resource.Resource, model any) tfsdk.State {
	t.Helper()
	var schemaResponse resource.SchemaResponse
	owner.Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	state := tfsdk.State{Schema: schemaResponse.Schema}
	if diagnostics := state.Set(context.Background(), model); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	return state
}
