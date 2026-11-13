// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package coverageaudit_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/coverageaudit"
	providerimpl "github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

func root(t *testing.T) string {
	t.Helper()
	r, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestManifestAndReportAreCurrent(t *testing.T) {
	r := root(t)
	audit, err := coverageaudit.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	want, err := fs.ReadFile(os.DirFS(r), coverageaudit.ReportPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(audit.Render())
	if got != string(want) {
		t.Fatal("generated coverage report is stale; run go generate ./...")
	}
	if !strings.Contains(got, "## Classification Summary") || !strings.Contains(got, "## Domain Summary") {
		t.Fatal("generated coverage report must retain classification and domain summaries")
	}
	for _, surface := range audit.Surfaces {
		marker := "#### `" + surface.Kind + ":" + surface.Name + "` — " + surface.Classification
		markerIndex := strings.Index(got, marker)
		if count := strings.Count(got, marker); count != 1 {
			t.Fatalf("rendered surface %s:%s count = %d, want 1", surface.Kind, surface.Name, count)
		}
		section := got[markerIndex:]
		if end := strings.Index(section[len(marker):], "\n#### "); end >= 0 {
			section = section[:len(marker)+end]
		}
		for label, evidence := range map[string]string{
			"Shared semantics":     surface.Evidence.Shared,
			"Baseline-only":        surface.Evidence.BaselineOnly,
			"Local-only":           surface.Evidence.LocalOnly,
			"Lifecycle":            surface.Evidence.Lifecycle,
			"Sensitive/write-only": surface.Evidence.SensitiveWriteOnly,
			"Sources":              strings.Join(surface.SourceRefs, "; "),
		} {
			if !strings.Contains(section, "- **"+label+":** "+evidence) {
				t.Fatalf("rendered surface %s:%s lacks %s evidence", surface.Kind, surface.Name, label)
			}
		}
	}
}

func TestMatrixClassificationExamples(t *testing.T) {
	r := root(t)
	audit, err := coverageaudit.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.Surfaces) != 94 {
		t.Fatalf("surface count = %d, want 94", len(audit.Surfaces))
	}

	if audit.LocalInventory.ResourceCount != 113 || audit.LocalInventory.DataSourceCount != 110 || len(audit.LocalInventory.ResourceOnly) != 9 {
		t.Fatal("dhcpv4 settings lookup inventory must be 113/110 with 9 lookup gaps")
	}
	for _, kind := range []string{"resource", "data-source"} {
		surface := findSurface(t, audit, kind, "trust_settings")
		if surface.Classification != "missing-buildable" || surface.FollowUp == "" {
			t.Fatal("settings increment must not claim full Trust audit parity")
		}
	}

	route := findSurface(t, audit, "resource", "route")
	if route.Classification != "covered-different-name" || strings.Join(route.LocalEquivalents, ",") != "system_route" {
		t.Fatalf("route mapping = %q %v", route.Classification, route.LocalEquivalents)
	}
	keaAlias := findSurface(t, audit, "resource", "kea_peer")
	if keaAlias.Classification != "intentionally-different" || strings.Join(keaAlias.LocalEquivalents, ",") != "kea_ha_peer" {
		t.Fatalf("deprecated Kea alias mapping = %q %v", keaAlias.Classification, keaAlias.LocalEquivalents)
	}
	interfaceData := findSurface(t, audit, "data-source", "interface")
	if interfaceData.Classification != "covered" || interfaceData.FollowUp != "" {
		t.Fatalf("interface data-source classification = %q, follow-up = %q", interfaceData.Classification, interfaceData.FollowUp)
	}
	openVPNKey := findSurface(t, audit, "ephemeral-resource", "openvpn_generate_key")
	if openVPNKey.Classification != "covered" || strings.Join(openVPNKey.LocalEquivalents, ",") != "openvpn_generate_key" || openVPNKey.FollowUp != "" {
		t.Fatalf("OpenVPN ephemeral classification = %q, equivalents = %v, follow-up = %q", openVPNKey.Classification, openVPNKey.LocalEquivalents, openVPNKey.FollowUp)
	}
}

func TestManifestRejectsInvalidMatrixRows(t *testing.T) {
	r := root(t)
	tests := map[string]struct {
		mutate func(*coverageaudit.Manifest)
		want   string
	}{
		"duplicate": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces = append(audit.Surfaces, audit.Surfaces[0])
		}, "duplicate surface"},
		"unknown classification": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].Classification = "unknown"
		}, "unknown classification"},
		"unknown kind": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].Kind = "unknown"
		}, "unknown or empty kind"},
		"empty kind": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].Kind = ""
		}, "unknown or empty kind"},
		"missing baseline row": {func(audit *coverageaudit.Manifest) {
			for i, surface := range audit.Surfaces {
				if surface.Kind == "resource" && surface.Name == "route" {
					audit.Surfaces = append(audit.Surfaces[:i], audit.Surfaces[i+1:]...)
					return
				}
			}
		}, "missing pinned baseline surface resource:route"},
		"buildable without follow-up": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].Classification = "missing-buildable"
			audit.Surfaces[0].FollowUp = ""
		}, "lacks follow_up"},
		"nonexistent nonempty follow-up": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].FollowUp = "_bmad-output/implementation-artifacts/does-not-exist.md"
		}, "follow_up is missing"},
		"nonactionable follow-up": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].FollowUp = "go.mod"
		}, "must declare P0 or P1 priority"},
		"follow-up ordered at story 31.1": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].FollowUp = "_bmad-output/implementation-artifacts/31-1-complete-provider-positioning-docs.md"
		}, "must appear before Story 31.1"},
		"follow-up without Given When Then": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].FollowUp = "_bmad-output/implementation-artifacts/31-1-complete-provider-positioning-docs.md"
		}, "lacks Given/When/Then acceptance criteria"},
		"covered unregistered equivalent": {func(audit *coverageaudit.Manifest) {
			mutateSurface(t, audit, "resource", "route", func(surface *coverageaudit.Surface) {
				surface.LocalEquivalents = []string{"not_registered"}
			})
		}, "names unregistered local equivalent"},
		"covered with follow-up": {func(audit *coverageaudit.Manifest) {
			mutateSurface(t, audit, "resource", "route", func(surface *coverageaudit.Surface) {
				surface.FollowUp = "_bmad-output/implementation-artifacts/31-0a-firewall-nat-field-parity.md"
			})
		}, "must not declare follow_up"},
		"covered with blocker": {func(audit *coverageaudit.Manifest) {
			mutateSurface(t, audit, "resource", "route", func(surface *coverageaudit.Surface) {
				surface.Blocker = "docs/upstream-blocked.md#gateway-group"
			})
		}, "must not declare blocker"},
		"buildable with blocker": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].Blocker = "docs/upstream-blocked.md#gateway-group"
		}, "must not declare blocker"},
		"blocked with arbitrary file": {func(audit *coverageaudit.Manifest) {
			mutateSurface(t, audit, "resource", "route", func(surface *coverageaudit.Surface) {
				surface.Classification = "missing-upstream-blocked"
				surface.Blocker = "go.mod#module"
			})
		}, "blocker must use docs/upstream-blocked.md"},
		"blocked with missing register symbol": {func(audit *coverageaudit.Manifest) {
			mutateSurface(t, audit, "resource", "route", func(surface *coverageaudit.Surface) {
				surface.Classification = "missing-upstream-blocked"
				surface.Blocker = "docs/upstream-blocked.md#does-not-exist"
			})
		}, "blocker register symbol \"does-not-exist\" does not exist"},
		"mutable upstream source": {func(audit *coverageaudit.Manifest) {
			audit.Surfaces[0].SourceRefs[0] = "https://github.com/browningluke/terraform-provider-opnsense/blob/main/provider/provider.go"
		}, "mutable or unpinned upstream source reference"},
		"declared resource docs drift": {func(audit *coverageaudit.Manifest) {
			audit.LocalInventory.ResourceDocCount++
		}, "declared resource_doc_count"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			audit, err := coverageaudit.Load(r)
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(audit)
			if err := audit.Validate(r); err == nil {
				t.Fatal("expected validation error")
			} else if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validation error = %q, want substring %q", err, test.want)
			}
		})
	}
}

func TestLocalInventoryMatchesRegistrationsAndDocs(t *testing.T) {
	r := root(t)
	audit, err := coverageaudit.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	p := providerimpl.New("test")()
	resources := registeredResources(t, p)
	dataSources := registeredDataSources(t, p)
	resourceDocs := docNames(t, filepath.Join(r, "docs", "resources"))
	dataSourceDocs := docNames(t, filepath.Join(r, "docs", "data-sources"))
	if len(resources) != audit.LocalInventory.ResourceCount || len(dataSources) != audit.LocalInventory.DataSourceCount {
		t.Fatalf("registered inventory drift: resources=%d data-sources=%d", len(resources), len(dataSources))
	}
	if len(resourceDocs) != audit.LocalInventory.ResourceDocCount || len(dataSourceDocs) != audit.LocalInventory.DataSourceDocCount {
		t.Fatalf("declared docs inventory drift: resources=%d/%d data-sources=%d/%d", audit.LocalInventory.ResourceDocCount, len(resourceDocs), audit.LocalInventory.DataSourceDocCount, len(dataSourceDocs))
	}
	if strings.Join(resources, "\x00") != strings.Join(resourceDocs, "\x00") {
		t.Fatal("registered resource names do not match generated resource docs")
	}
	if strings.Join(dataSources, "\x00") != strings.Join(dataSourceDocs, "\x00") {
		t.Fatal("registered data-source names do not match generated data-source docs")
	}
	resourceOnly := difference(resources, dataSources)
	dataOnly := difference(dataSources, resources)
	if strings.Join(resourceOnly, "\x00") != strings.Join(audit.LocalInventory.ResourceOnly, "\x00") {
		t.Fatalf("resource-only gaps drifted: %v", resourceOnly)
	}
	if strings.Join(dataOnly, "\x00") != strings.Join(audit.LocalInventory.DataSourceOnly, "\x00") {
		t.Fatalf("data-only inventory drifted: %v", dataOnly)
	}
}

func registeredResources(t *testing.T, p provider.Provider) []string {
	t.Helper()
	var names []string
	for _, factory := range p.Resources(context.Background()) {
		var resp resource.MetadataResponse
		factory().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "opnsense"}, &resp)
		names = append(names, strings.TrimPrefix(resp.TypeName, "opnsense_"))
	}
	sort.Strings(names)
	return names
}

func registeredDataSources(t *testing.T, p provider.Provider) []string {
	t.Helper()
	var names []string
	for _, factory := range p.DataSources(context.Background()) {
		var resp datasource.MetadataResponse
		factory().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "opnsense"}, &resp)
		names = append(names, strings.TrimPrefix(resp.TypeName, "opnsense_"))
	}
	sort.Strings(names)
	return names
}

func docNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			names = append(names, strings.TrimSuffix(entry.Name(), ".md"))
		}
	}
	sort.Strings(names)
	return names
}

func difference(left, right []string) []string {
	have := make(map[string]bool, len(right))
	for _, name := range right {
		have[name] = true
	}
	var result []string
	for _, name := range left {
		if !have[name] {
			result = append(result, name)
		}
	}
	return result
}

func findSurface(t *testing.T, audit *coverageaudit.Manifest, kind, name string) coverageaudit.Surface {
	t.Helper()
	for _, surface := range audit.Surfaces {
		if surface.Kind == kind && surface.Name == name {
			return surface
		}
	}
	t.Fatalf("missing surface %s:%s", kind, name)
	return coverageaudit.Surface{}
}

func mutateSurface(t *testing.T, audit *coverageaudit.Manifest, kind, name string, mutate func(*coverageaudit.Surface)) {
	t.Helper()
	for i := range audit.Surfaces {
		if audit.Surfaces[i].Kind == kind && audit.Surfaces[i].Name == name {
			mutate(&audit.Surfaces[i])
			return
		}
	}
	t.Fatalf("missing surface %s:%s", kind, name)
}

func TestAdvancedFRRCandidateRetainsProofBoundary(t *testing.T) {
	audit, err := coverageaudit.Load(root(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, surface := range audit.Surfaces {
		if surface.Kind != "resource" || surface.Name != "quagga_bgp_neighbor" {
			continue
		}
		if surface.Classification != "missing-buildable" || !strings.Contains(surface.Evidence.BaselineOnly, "candidate") || !strings.Contains(surface.Evidence.BaselineOnly, "pending") {
			t.Fatal("advanced implementation was promoted to full proof/release")
		}
		if !strings.Contains(surface.Evidence.SensitiveWriteOnly, "never remote-adopted") || !strings.Contains(surface.Evidence.SensitiveWriteOnly, "not encryption") {
			t.Fatal("authentication boundary missing from audit")
		}
		return
	}
	t.Fatal("neighbor audit row missing")
}
