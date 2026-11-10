// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package coverageaudit validates and renders the pinned provider coverage audit.
package coverageaudit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	providerimpl "github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

const (
	// ManifestPath is the repository-relative canonical audit manifest.
	ManifestPath = "_bmad-output/planning-artifacts/provider-ecosystem-coverage-audit.json"
	// ReportPath is the repository-relative generated human audit.
	ReportPath = "_bmad-output/planning-artifacts/provider-ecosystem-coverage-audit.md"
)

var allowedClassifications = map[string]bool{
	"covered":                  true,
	"covered-different-name":   true,
	"covered-broader":          true,
	"missing-buildable":        true,
	"missing-upstream-blocked": true,
	"intentionally-different":  true,
	"needs-field-audit":        true,
}

var actionablePriority = regexp.MustCompile(`(?m)^## Priority\s*\n+\s*P[01]\b`)

const positioningStoryID = "31-1-complete-provider-positioning-docs"

var expectedSurfaces = map[string]string{
	"resource":               `cron_job dnsmasq_host firewall_alias firewall_category firewall_filter firewall_nat firewall_nat_one_to_one firewall_nat_port_forward interfaces_vip interfaces_vlan ipsec_auth_local ipsec_auth_remote ipsec_child ipsec_connection ipsec_psk ipsec_vti kea_dhcpv4_peer kea_dhcpv4_reservation kea_dhcpv4_subnet kea_dhcpv6_pd_pool kea_dhcpv6_peer kea_dhcpv6_reservation kea_dhcpv6_subnet kea_peer kea_reservation kea_subnet openvpn_client_overwrite openvpn_instance openvpn_static_key quagga_bgp_aspath quagga_bgp_communitylist quagga_bgp_neighbor quagga_bgp_prefixlist quagga_bgp_routemap route trust_ca trust_cert trust_settings unbound_acl unbound_domain_override unbound_forward unbound_host_alias unbound_host_override unbound_settings wireguard_client wireguard_server wireguard_settings`,
	"data-source":            `cron_job dnsmasq_host firewall_alias firewall_category firewall_filter firewall_nat firewall_nat_one_to_one firewall_nat_port_forward interface interface_all interfaces_overview interfaces_overview_all interfaces_vip interfaces_vlan kea_dhcpv4_peer kea_dhcpv4_reservation kea_dhcpv4_subnet kea_dhcpv6_pd_pool kea_dhcpv6_peer kea_dhcpv6_reservation kea_dhcpv6_subnet kea_peer kea_reservation kea_subnet openvpn_client_overwrite openvpn_instance openvpn_static_key quagga_bgp_aspath quagga_bgp_communitylist quagga_bgp_neighbor quagga_bgp_prefixlist quagga_bgp_routemap route trust_ca trust_cert trust_settings unbound_acl unbound_domain_override unbound_forward unbound_host_alias unbound_host_override unbound_settings wireguard_client wireguard_server wireguard_settings`,
	"ephemeral-resource":     "openvpn_generate_key",
	"provider-configuration": "retry-settings",
}

// Manifest is the machine-readable coverage source of truth.
type Manifest struct {
	SchemaVersion          int            `json:"schema_version"`
	Baseline               Baseline       `json:"baseline"`
	LocalInventory         LocalInventory `json:"local_inventory"`
	AllowedClassifications []string       `json:"allowed_classifications"`
	OwnershipDirection     string         `json:"ownership_direction"`
	Surfaces               []Surface      `json:"surfaces"`
}

// Baseline pins the external comparison input.
type Baseline struct {
	Source              string `json:"source"`
	Version             string `json:"version"`
	Commit              string `json:"commit"`
	ResourceCount       int    `json:"resource_count"`
	DataSourceCount     int    `json:"data_source_count"`
	EphemeralCount      int    `json:"ephemeral_resource_count"`
	ProviderConfigCount int    `json:"provider_config_count"`
}

// LocalInventory records the expected generated documentation inventory.
type LocalInventory struct {
	ResourceCount      int      `json:"resource_count"`
	DataSourceCount    int      `json:"data_source_count"`
	ResourceDocCount   int      `json:"resource_doc_count"`
	DataSourceDocCount int      `json:"data_source_doc_count"`
	EphemeralCount     int      `json:"ephemeral_resource_count"`
	EphemeralDocCount  int      `json:"ephemeral_resource_doc_count"`
	ResourceOnly       []string `json:"resource_only_data_source_gaps"`
	DataSourceOnly     []string `json:"data_source_only"`
}

// Surface is one published baseline surface. Provider configuration is one
// synthetic row because the retry knobs form one semantic configuration unit.
type Surface struct {
	Kind             string        `json:"kind"`
	Name             string        `json:"name"`
	Domain           string        `json:"domain"`
	Classification   string        `json:"classification"`
	LocalEquivalents []string      `json:"local_equivalents"`
	BaselineFields   []string      `json:"baseline_fields"`
	Evidence         FieldEvidence `json:"field_evidence"`
	SourceRefs       []string      `json:"source_refs"`
	FollowUp         string        `json:"follow_up,omitempty"`
	Blocker          string        `json:"blocker,omitempty"`
}

// FieldEvidence makes semantic and lifecycle differences explicit.
type FieldEvidence struct {
	Shared             string `json:"shared"`
	BaselineOnly       string `json:"baseline_only"`
	LocalOnly          string `json:"local_only"`
	Lifecycle          string `json:"lifecycle"`
	SensitiveWriteOnly string `json:"sensitive_write_only"`
}

// Load reads and validates the committed manifest.
func Load(root string) (*Manifest, error) {
	b, err := fs.ReadFile(os.DirFS(root), ManifestPath)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parse coverage manifest: %w", err)
	}
	if err := m.Validate(root); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate enforces pinned-baseline completeness and follow-up policy without
// requiring network access.
func (m *Manifest) Validate(root string) error {
	var problems []string
	registrations := localRegistrations()
	resourceDocs, resourceDocsErr := localDocNames(root, "resources")
	dataSourceDocs, dataSourceDocsErr := localDocNames(root, "data-sources")
	ephemeralDocs, ephemeralDocsErr := localDocNames(root, "ephemeral-resources")
	if resourceDocsErr != nil {
		problems = append(problems, "read resource docs: "+resourceDocsErr.Error())
	}
	if dataSourceDocsErr != nil {
		problems = append(problems, "read data-source docs: "+dataSourceDocsErr.Error())
	}
	if ephemeralDocsErr != nil {
		problems = append(problems, "read ephemeral-resource docs: "+ephemeralDocsErr.Error())
	}
	if resourceDocsErr == nil && len(resourceDocs) != m.LocalInventory.ResourceDocCount {
		problems = append(problems, fmt.Sprintf("declared resource_doc_count=%d, actual=%d", m.LocalInventory.ResourceDocCount, len(resourceDocs)))
	}
	if dataSourceDocsErr == nil && len(dataSourceDocs) != m.LocalInventory.DataSourceDocCount {
		problems = append(problems, fmt.Sprintf("declared data_source_doc_count=%d, actual=%d", m.LocalInventory.DataSourceDocCount, len(dataSourceDocs)))
	}
	if ephemeralDocsErr == nil && len(ephemeralDocs) != m.LocalInventory.EphemeralDocCount {
		problems = append(problems, fmt.Sprintf("declared ephemeral_resource_doc_count=%d, actual=%d", m.LocalInventory.EphemeralDocCount, len(ephemeralDocs)))
	}
	if len(registrations["resource"]) != m.LocalInventory.ResourceCount || len(registrations["data-source"]) != m.LocalInventory.DataSourceCount {
		problems = append(problems, fmt.Sprintf("declared local registrations drifted: resources=%d/%d data-sources=%d/%d", m.LocalInventory.ResourceCount, len(registrations["resource"]), m.LocalInventory.DataSourceCount, len(registrations["data-source"])))
	}
	if len(registrations["ephemeral-resource"]) != m.LocalInventory.EphemeralCount {
		problems = append(problems, fmt.Sprintf("declared ephemeral registrations drifted: ephemeral-resources=%d/%d", m.LocalInventory.EphemeralCount, len(registrations["ephemeral-resource"])))
	}
	if resourceDocsErr == nil && !sameNames(registrations["resource"], resourceDocs) {
		problems = append(problems, "registered resources do not match generated resource docs")
	}
	if dataSourceDocsErr == nil && !sameNames(registrations["data-source"], dataSourceDocs) {
		problems = append(problems, "registered data sources do not match generated data-source docs")
	}
	if ephemeralDocsErr == nil && !sameNames(registrations["ephemeral-resource"], ephemeralDocs) {
		problems = append(problems, "registered ephemeral resources do not match generated ephemeral-resource docs")
	}
	if m.SchemaVersion != 1 {
		problems = append(problems, "schema_version must be 1")
	}
	if m.Baseline.Source != "browningluke/opnsense" || m.Baseline.Version != "0.26.0" || m.Baseline.Commit != "33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba" {
		problems = append(problems, "baseline source/version/commit does not match the pinned audit baseline")
	}
	if m.Baseline.ResourceCount != 47 || m.Baseline.DataSourceCount != 45 || m.Baseline.EphemeralCount != 1 || m.Baseline.ProviderConfigCount != 1 {
		problems = append(problems, "baseline counts must be 47/45/1/1")
	}
	if m.OwnershipDirection == "" || !strings.Contains(strings.ToLower(m.OwnershipDirection), "one provider") {
		problems = append(problems, "ownership_direction must state the one-provider direction")
	}
	allowed := append([]string(nil), m.AllowedClassifications...)
	sort.Strings(allowed)
	wantAllowed := make([]string, 0, len(allowedClassifications))
	for name := range allowedClassifications {
		wantAllowed = append(wantAllowed, name)
	}
	sort.Strings(wantAllowed)
	if strings.Join(allowed, "\x00") != strings.Join(wantAllowed, "\x00") {
		problems = append(problems, "allowed_classifications drifted")
	}

	seen := make(map[string]bool, len(m.Surfaces))
	counts := map[string]int{}
	validatedFollowUps := map[string]bool{}
	for _, row := range m.Surfaces {
		key := row.Kind + ":" + row.Name
		if _, ok := expectedSurfaces[row.Kind]; !ok {
			problems = append(problems, fmt.Sprintf("surface %s has unknown or empty kind %q", key, row.Kind))
		}
		if seen[key] {
			problems = append(problems, "duplicate surface "+key)
		}
		seen[key] = true
		counts[row.Kind]++
		if !allowedClassifications[row.Classification] {
			problems = append(problems, fmt.Sprintf("surface %s has unknown classification %q", key, row.Classification))
		}
		if row.Domain == "" || len(row.BaselineFields) == 0 || len(row.SourceRefs) < 2 {
			problems = append(problems, "surface "+key+" lacks domain, fields, or source references")
		}
		hasUpstreamEvidence := false
		for _, ref := range row.SourceRefs {
			if !strings.HasPrefix(ref, "https://") {
				if _, err := fs.Stat(os.DirFS(root), ref); err != nil {
					problems = append(problems, "surface "+key+" has missing local source reference: "+ref)
				}
			} else {
				hasUpstreamEvidence = true
				if !strings.Contains(ref, m.Baseline.Commit) {
					problems = append(problems, "surface "+key+" has mutable or unpinned upstream source reference: "+ref)
				}
			}
		}
		if !hasUpstreamEvidence {
			problems = append(problems, "surface "+key+" lacks immutable upstream evidence")
		}
		if row.Evidence.Shared == "" || row.Evidence.BaselineOnly == "" || row.Evidence.LocalOnly == "" || row.Evidence.Lifecycle == "" || row.Evidence.SensitiveWriteOnly == "" {
			problems = append(problems, "surface "+key+" has incomplete field/lifecycle evidence")
		}
		if row.FollowUp != "" && !validatedFollowUps[row.FollowUp] {
			problems = append(problems, validateFollowUp(root, row.FollowUp)...)
			validatedFollowUps[row.FollowUp] = true
		}
		if isCovered(row.Classification) {
			if len(row.LocalEquivalents) == 0 {
				problems = append(problems, "covered surface "+key+" lacks a local equivalent")
			}
			for _, equivalent := range row.LocalEquivalents {
				if !registrations[row.Kind][equivalent] {
					problems = append(problems, fmt.Sprintf("covered surface %s names unregistered local equivalent %q", key, equivalent))
				}
				if row.Kind == "provider-configuration" {
					content, err := fs.ReadFile(os.DirFS(root), "docs/index.md")
					if err != nil || !strings.Contains(string(content), "`"+equivalent+"`") {
						problems = append(problems, fmt.Sprintf("covered provider setting %s lacks provider docs", equivalent))
					}
					continue
				}
				docs := resourceDocs
				docKind := "resource"
				switch row.Kind {
				case "data-source":
					docs = dataSourceDocs
					docKind = "data-source"
				case "ephemeral-resource":
					docs = ephemeralDocs
					docKind = "ephemeral-resource"
				}
				if !docs[equivalent] {
					problems = append(problems, fmt.Sprintf("covered surface %s local equivalent %q lacks %s docs", key, equivalent, docKind))
				}
			}
		}
		if row.Classification == "missing-buildable" {
			if row.FollowUp == "" {
				problems = append(problems, "buildable surface "+key+" lacks follow_up")
			}
			if row.Blocker != "" {
				problems = append(problems, "buildable surface "+key+" must not declare blocker")
			}
		} else if row.FollowUp != "" {
			problems = append(problems, fmt.Sprintf("surface %s classification %q must not declare follow_up", key, row.Classification))
		}
		if row.Classification == "missing-upstream-blocked" {
			if err := validateBlocker(root, row.Blocker); err != nil {
				problems = append(problems, "blocked surface "+key+": "+err.Error())
			}
		} else if row.Blocker != "" {
			problems = append(problems, fmt.Sprintf("surface %s classification %q must not declare blocker", key, row.Classification))
		}
		if row.Classification == "needs-field-audit" && strings.TrimSpace(row.Evidence.BaselineOnly) == "" {
			problems = append(problems, "needs-field-audit surface "+key+" lacks unresolved proof")
		}
	}
	if counts["resource"] != 47 || counts["data-source"] != 45 || counts["ephemeral-resource"] != 1 || counts["provider-configuration"] != 1 {
		problems = append(problems, fmt.Sprintf("surface counts are resource=%d data-source=%d ephemeral=%d provider-configuration=%d", counts["resource"], counts["data-source"], counts["ephemeral-resource"], counts["provider-configuration"]))
	}
	for kind, names := range expectedSurfaces {
		for _, name := range strings.Fields(names) {
			if !seen[kind+":"+name] {
				problems = append(problems, "missing pinned baseline surface "+kind+":"+name)
			}
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "\n"))
	}
	return nil
}

func isCovered(classification string) bool {
	return classification == "covered" || classification == "covered-different-name" || classification == "covered-broader"
}

func localRegistrations() map[string]map[string]bool {
	registrations := map[string]map[string]bool{
		"resource": {}, "data-source": {}, "ephemeral-resource": {}, "provider-configuration": {},
	}
	p := providerimpl.New("coverage-audit")()
	var schema provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schema)
	for name := range schema.Schema.Attributes {
		registrations["provider-configuration"][name] = true
	}
	for _, factory := range p.Resources(context.Background()) {
		var resp resource.MetadataResponse
		factory().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "opnsense"}, &resp)
		registrations["resource"][strings.TrimPrefix(resp.TypeName, "opnsense_")] = true
	}
	for _, factory := range p.DataSources(context.Background()) {
		var resp datasource.MetadataResponse
		factory().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "opnsense"}, &resp)
		registrations["data-source"][strings.TrimPrefix(resp.TypeName, "opnsense_")] = true
	}
	withEphemeral, ok := p.(provider.ProviderWithEphemeralResources)
	if ok {
		for _, factory := range withEphemeral.EphemeralResources(context.Background()) {
			var resp ephemeral.MetadataResponse
			factory().Metadata(context.Background(), ephemeral.MetadataRequest{ProviderTypeName: "opnsense"}, &resp)
			registrations["ephemeral-resource"][strings.TrimPrefix(resp.TypeName, "opnsense_")] = true
		}
	}
	return registrations
}

func localDocNames(root, kind string) (map[string]bool, error) {
	entries, err := fs.ReadDir(os.DirFS(root), filepath.Join("docs", kind))
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			names[strings.TrimSuffix(entry.Name(), ".md")] = true
		}
	}
	return names, nil
}

func sameNames(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for name := range left {
		if !right[name] {
			return false
		}
	}
	return true
}

func validateFollowUp(root, path string) []string {
	var problems []string
	if !strings.HasPrefix(path, "_bmad-output/implementation-artifacts/") || filepath.Ext(path) != ".md" {
		problems = append(problems, "follow_up "+path+" must be an implementation story")
	}
	story, err := fs.ReadFile(os.DirFS(root), path)
	if err != nil {
		return []string{"buildable follow_up is missing: " + path}
	}
	storyText := string(story)
	if !actionablePriority.MatchString(storyText) {
		problems = append(problems, "follow_up "+path+" must declare P0 or P1 priority")
	}
	acceptanceCriteria := ""
	if start := strings.Index(storyText, "## Acceptance Criteria"); start >= 0 {
		acceptanceCriteria = storyText[start+len("## Acceptance Criteria"):]
		if end := strings.Index(acceptanceCriteria, "\n## "); end >= 0 {
			acceptanceCriteria = acceptanceCriteria[:end]
		}
	}
	lowerStory := strings.ToLower(acceptanceCriteria)
	for _, keyword := range []string{"given", "when", "then"} {
		if !strings.Contains(lowerStory, keyword) {
			problems = append(problems, "follow_up "+path+" lacks Given/When/Then acceptance criteria")
			break
		}
	}
	status, err := fs.ReadFile(os.DirFS(root), "_bmad-output/implementation-artifacts/sprint-status.yaml")
	if err != nil {
		return append(problems, "read sprint-status.yaml: "+err.Error())
	}
	storyID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	storyPosition := sprintStoryPosition(string(status), storyID)
	positioningPosition := sprintStoryPosition(string(status), positioningStoryID)
	if storyPosition < 0 {
		problems = append(problems, "follow_up "+path+" is missing from sprint-status.yaml")
	} else if positioningPosition < 0 || storyPosition >= positioningPosition {
		problems = append(problems, "follow_up "+path+" must appear before Story 31.1 in sprint-status.yaml")
	}
	return problems
}

func sprintStoryPosition(status, storyID string) int {
	needle := "  " + storyID + ":"
	for offset, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, needle) {
			return offset
		}
	}
	return -1
}

func validateBlocker(root, blocker string) error {
	path, symbol, ok := strings.Cut(blocker, "#")
	if !ok || path != "docs/upstream-blocked.md" || strings.TrimSpace(symbol) == "" {
		return errors.New("blocker must use docs/upstream-blocked.md#<register-symbol>")
	}
	register, err := fs.ReadFile(os.DirFS(root), path)
	if err != nil {
		return fmt.Errorf("read blocker register: %w", err)
	}
	want := strings.ToLower(strings.ReplaceAll(symbol, "-", " "))
	if !strings.Contains(strings.ToLower(string(register)), want) {
		return fmt.Errorf("blocker register symbol %q does not exist", symbol)
	}
	return nil
}

// Render produces the deterministic human-readable audit.
func (m *Manifest) Render() []byte {
	var b bytes.Buffer
	fmt.Fprintln(&b, "# Provider Ecosystem Coverage Audit")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Pinned baseline: `%s` v%s at `%s` (%d resources, %d data sources, %d ephemeral resource, and provider retry configuration).\n\n", m.Baseline.Source, m.Baseline.Version, m.Baseline.Commit, m.Baseline.ResourceCount, m.Baseline.DataSourceCount, m.Baseline.EphemeralCount)
	fmt.Fprintf(&b, "Current local inventory: **%d resources and %d data sources**, with %d/%d generated docs. There are **%d resource-only same-name data-source gaps** and the data-only surfaces are `%s`.\n\n", m.LocalInventory.ResourceCount, m.LocalInventory.DataSourceCount, m.LocalInventory.ResourceDocCount, m.LocalInventory.DataSourceDocCount, len(m.LocalInventory.ResourceOnly), strings.Join(m.LocalInventory.DataSourceOnly, "`, `"))
	fmt.Fprintln(&b, "## Ownership Direction")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, m.OwnershipDirection)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Classification Summary")
	fmt.Fprintln(&b)
	classCounts := map[string]int{}
	for _, row := range m.Surfaces {
		classCounts[row.Classification]++
	}
	for _, class := range m.AllowedClassifications {
		fmt.Fprintf(&b, "- `%s`: %d\n", class, classCounts[class])
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Domain Summary")
	fmt.Fprintln(&b)
	domainCounts := map[string]map[string]int{}
	for _, row := range m.Surfaces {
		if domainCounts[row.Domain] == nil {
			domainCounts[row.Domain] = map[string]int{}
		}
		domainCounts[row.Domain][row.Classification]++
	}
	domains := make([]string, 0, len(domainCounts))
	for domain := range domainCounts {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	for _, domain := range domains {
		var parts []string
		total := 0
		for _, class := range m.AllowedClassifications {
			count := domainCounts[domain][class]
			if count > 0 {
				parts = append(parts, fmt.Sprintf("`%s`=%d", class, count))
				total += count
			}
		}
		fmt.Fprintf(&b, "- **%s:** %d surfaces (%s)\n", domain, total, strings.Join(parts, ", "))
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Surface Audit")
	fmt.Fprintln(&b)
	for _, kind := range []string{"provider-configuration", "resource", "data-source", "ephemeral-resource"} {
		fmt.Fprintf(&b, "### %s\n\n", title(kind))
		for _, row := range m.Surfaces {
			if row.Kind != kind {
				continue
			}
			local := "none"
			if len(row.LocalEquivalents) > 0 {
				local = "`" + strings.Join(row.LocalEquivalents, "`, `") + "`"
			}
			fmt.Fprintf(&b, "#### `%s:%s` — %s\n\n", row.Kind, row.Name, row.Classification)
			fmt.Fprintf(&b, "- **Domain / local mapping:** %s / %s\n", row.Domain, local)
			fmt.Fprintf(&b, "- **Baseline fields:** %s\n", strings.Join(row.BaselineFields, ", "))
			fmt.Fprintf(&b, "- **Shared semantics:** %s\n", row.Evidence.Shared)
			fmt.Fprintf(&b, "- **Baseline-only:** %s\n", row.Evidence.BaselineOnly)
			fmt.Fprintf(&b, "- **Local-only:** %s\n", row.Evidence.LocalOnly)
			fmt.Fprintf(&b, "- **Lifecycle:** %s\n", row.Evidence.Lifecycle)
			fmt.Fprintf(&b, "- **Sensitive/write-only:** %s\n", row.Evidence.SensitiveWriteOnly)
			fmt.Fprintf(&b, "- **Sources:** %s\n", strings.Join(row.SourceRefs, "; "))
			if row.FollowUp != "" {
				fmt.Fprintf(&b, "- **Follow-up:** `%s`\n", row.FollowUp)
			}
			if row.Blocker != "" {
				fmt.Fprintf(&b, "- **Blocker:** `%s`\n", row.Blocker)
			}
			fmt.Fprintln(&b)
		}
	}
	fmt.Fprintln(&b, "## Local Inventory Exceptions")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Resource-only same-name data-source gaps (%d): `%s`.\n\n", len(m.LocalInventory.ResourceOnly), strings.Join(m.LocalInventory.ResourceOnly, "`, `"))
	fmt.Fprintf(&b, "Data-only: `%s`.\n", strings.Join(m.LocalInventory.DataSourceOnly, "`, `"))
	return b.Bytes()
}

func title(s string) string {
	parts := strings.Split(s, "-")
	for i := range parts {
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}
