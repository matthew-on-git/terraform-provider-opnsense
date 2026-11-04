// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

var (
	frrObservationCounter atomic.Uint64
	frrCollectorSHA       atomic.Value
)

func frrOracleBase() string {
	return strings.TrimSuffix(os.Getenv("OPNSENSE_FRR_ORACLE_URI"), "/snapshot")
}

func frrRequiredBinding(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Fatal("missing required isolated runtime binding:", name)
	}
	return value
}

type frrObservation struct {
	Identity             struct{ Name, UUID string } `json:"identity"`
	Core, Plugin, Daemon frrCommand
	Config               frrCommand `json:"frr_config"`
	PF                   frrCommand `json:"pf_rules"`
	AllPF                frrCommand `json:"pf_all_rules"`
	Rendered             frrCommand `json:"pf_rendered"`
	Routes               frrCommand `json:"routes"`
	Summary              frrCommand `json:"bgp_summary"`
	Advertised4          frrCommand `json:"advertised_ipv4"`
	Advertised6          frrCommand `json:"advertised_ipv6"`
	Collector            struct {
		Active          bool     `json:"active"`
		Prefixes        []string `json:"prefixes"`
		EverPrefixes    []string `json:"ever_prefixes"`
		DefaultObserved bool     `json:"default_observed"`
		Sessions        int      `json:"sessions"`
	} `json:"collector"`
	CollectorSHA string `json:"collector_sha256"`
}
type frrCommand struct {
	Status         int
	Stdout, Stderr string
}

func frrObserve(t *testing.T) frrObservation {
	t.Helper()
	var result frrObservation
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, frrOracleBase()+"/snapshot", nil)
	if err != nil {
		t.Fatal("invalid observation URI")
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPNSENSE_FRR_ORACLE_TOKEN"))
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal("independent observation transport failed")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != 200 || json.NewDecoder(response.Body).Decode(&result) != nil {
		t.Fatal("independent observation failed")
	}
	if result.Identity.Name != frrRequiredBinding(t, "OPNSENSE_DISPOSABLE_VM_NAME") || result.Identity.UUID != frrRequiredBinding(t, "OPNSENSE_DISPOSABLE_VM_UUID") {
		t.Fatal("observation VM identity mismatch")
	}
	if result.Core.Status != 0 || frrVersionToken(result.Core.Stdout, "OPNsense") != frrRequiredBinding(t, "OPNSENSE_FRR_EXPECTED_CORE") || result.Plugin.Status != 0 || strings.TrimSpace(result.Plugin.Stdout) != frrRequiredBinding(t, "OPNSENSE_FRR_EXPECTED_PLUGIN") || result.Daemon.Status != 0 || frrVersionToken(result.Daemon.Stdout, "bgpd version") != frrRequiredBinding(t, "OPNSENSE_FRR_EXPECTED_DAEMON") {
		t.Fatal("independent core/plugin/daemon version mismatch")
	}
	if initial := frrCollectorSHA.Load(); initial == nil {
		frrCollectorSHA.Store(result.CollectorSHA)
	} else if initial.(string) != result.CollectorSHA {
		t.Fatal("independent collector source changed during proof")
	}
	if len(result.CollectorSHA) != 64 || result.Config.Status != 0 || result.PF.Status != 0 || result.AllPF.Status != 0 || result.Rendered.Status != 0 || result.Routes.Status != 0 {
		t.Fatal("independent generated/config/PF/routing observation incomplete")
	}
	if dir := os.Getenv("OPNSENSE_FRR_ARTIFACT_DIR"); dir != "" {
		root, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal("cannot retain independent observations")
		}
		defer func() { _ = root.Close() }()
		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatal("cannot encode independent observations")
		}
		if root.WriteFile("observation-"+strconv.FormatUint(frrObservationCounter.Add(1), 10)+".json", encoded, 0o600) != nil {
			t.Fatal("cannot retain independent observations")
		}
	}
	return result
}

func frrLabOperation(t *testing.T, operation string) {
	t.Helper()
	uri := frrOracleBase() + "/filter-reload"
	if operation != "filter-reload" {
		t.Fatal("unsupported lab operation")
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, uri, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPNSENSE_FRR_ORACLE_TOKEN"))
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		t.Fatal("explicit lab filter activation failed")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != 200 {
		t.Fatal("explicit lab filter activation rejected")
	}
	t.Logf("FRR_LAB_OPERATION operation=%s source=explicit_lab_not_provider", operation)
}

func frrLabFilterReload(t *testing.T) {
	t.Helper()
	frrLabOperation(t, "filter-reload")
	t.Log("FRR_FIREWALL_ACTIVATION source=explicit_lab_filter_reload provider_reconfigure=FRR_only")
}

func frrLabDisabledConfigSync(t *testing.T) {
	t.Helper()
	for _, kind := range []string{"general", "bgp"} {
		if string(frrRawSettings(t, kind)["enabled"]) != `"0"` {
			t.Fatal("lab disabled-config sync refuses enabled settings")
		}
	}
	root, err := os.OpenRoot(os.Getenv("OPNSENSE_FRR_ARTIFACT_DIR"))
	if err != nil {
		t.Fatal("cannot open private host handshake directory")
	}
	defer func() { _ = root.Close() }()
	if root.WriteFile("disabled-config-sync.request", []byte(frrRequiredBinding(t, "OPNSENSE_DISPOSABLE_VM_UUID")), 0o600) != nil {
		t.Fatal("cannot request host-side disabled config sync")
	}
	deadline := time.Now().Add(30 * time.Second)
	for {
		raw, err := root.ReadFile("disabled-config-sync.done")
		if err == nil {
			if string(raw) != "ok" {
				t.Fatal("host-side disabled config sync failed")
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("host-side disabled config sync timed out")
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Log("FRR_LAB_OPERATION operation=disabled-config-sync source=explicit_host_probe_not_provider")
}

func frrRaw(t *testing.T, method, path string, body any) map[string]json.RawMessage {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	client := acctest.TestClient(t)
	req, err := http.NewRequestWithContext(context.Background(), method, client.BaseURL()+path, &payload)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := client.HTTPClient().Do(req)
	if err != nil {
		t.Fatal("raw API oracle transport failed")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != 200 {
		t.Fatalf("raw API oracle %s HTTP %d", path, response.StatusCode)
	}
	var result map[string]json.RawMessage
	if json.NewDecoder(response.Body).Decode(&result) != nil {
		t.Fatalf("raw API oracle %s malformed response", path)
	}
	return result
}

func frrRawSettings(t *testing.T, kind string) map[string]json.RawMessage {
	t.Helper()
	outer := frrRaw(t, "GET", "/api/quagga/"+kind+"/get", nil)
	var result map[string]json.RawMessage
	if json.Unmarshal(outer[kind], &result) != nil || len(result) == 0 {
		t.Fatalf("raw %s settings absent", kind)
	}
	return result
}

func frrRawSelection(raw json.RawMessage) (string, error) {
	var scalar *string
	if json.Unmarshal(raw, &scalar) == nil {
		if scalar == nil {
			return "", fmt.Errorf("null raw selector")
		}
		return *scalar, nil
	}
	if string(raw) == "[]" {
		return "", nil
	}
	var options map[string]struct {
		Selected json.RawMessage `json:"selected"`
	}
	if json.Unmarshal(raw, &options) != nil || options == nil {
		return "", fmt.Errorf("malformed raw selector")
	}
	values := []string{}
	for id, entry := range options {
		flag := strings.Trim(string(entry.Selected), "\"")
		if flag != "0" && flag != "1" {
			return "", fmt.Errorf("malformed raw selected flag")
		}
		if flag == "1" && id != "" {
			values = append(values, id)
		}
	}
	sort.Strings(values)
	return strings.Join(values, ","), nil
}

func frrRestoreSettings(t *testing.T, kind string, snapshot map[string]json.RawMessage) {
	t.Helper()
	payload := map[string]string{}
	for key, raw := range snapshot {
		if strings.HasPrefix(key, "neighbors") || key == "aspaths" || key == "prefixlists" || key == "communitylists" || key == "routemaps" || key == "peergroups" || key == "redistributions" {
			continue
		}
		scalar, err := frrRawSelection(raw)
		if err != nil {
			t.Fatalf("cannot restore raw %s field %s", kind, key)
		}
		payload[key] = scalar
	}
	result := frrRaw(t, "POST", "/api/quagga/"+kind+"/set", map[string]any{kind: payload})
	if string(result["result"]) != `"saved"` {
		t.Fatalf("restoration save for %s rejected", kind)
	}
}

func frrTagged(t *testing.T, kind, marker string) []string {
	t.Helper()
	var ids []string
	total := -1
	for page := 1; total < 0 || len(ids) < total; page++ {
		body := frrRaw(t, "POST", "/api/quagga/bgp/search_"+kind, map[string]any{"current": page, "rowCount": 500, "searchPhrase": marker})
		var rows []struct {
			UUID string `json:"uuid"`
		}
		var count int
		if json.Unmarshal(body["rows"], &rows) != nil || json.Unmarshal(body["total"], &count) != nil || count < 0 {
			t.Fatal("raw tagged search malformed")
		}
		if total >= 0 && count != total {
			t.Fatal("raw tagged search changed during pagination")
		}
		total = count
		if len(rows) == 0 && len(ids) < total {
			t.Fatal("raw tagged search incomplete")
		}
		for _, row := range rows {
			if row.UUID == "" {
				t.Fatal("raw tagged search missing identity")
			}
			ids = append(ids, row.UUID)
		}
	}
	return ids
}

// Separate standard testing cleanups keep later phases running after Fatal/Goexit.
func frrRegisterCleanup(t interface{ Cleanup(func()) }, phases ...func()) {
	for i := len(phases) - 1; i >= 0; i-- {
		t.Cleanup(phases[i])
	}
}

func frrAssertNamespaceZero(t *testing.T, kind, phase string) {
	t.Helper()
	ids := frrTagged(t, kind, "")
	if len(ids) != 0 {
		t.Errorf("%s %s has %d global leftovers before rollback", phase, kind, len(ids))
	}
	t.Logf("FRR_%s namespace=%s total=%d", phase, kind, len(ids))
}

func frrAssertZeros(t *testing.T, _, phase string) {
	t.Helper()
	for _, kind := range []string{"neighbor", "peergroup", "routemap", "prefixlist", "redistribution", "aspath", "communitylist"} {
		frrAssertNamespaceZero(t, kind, phase)
	}
}

func frrPreflight(t *testing.T) {
	t.Helper()
	acctest.PreCheck(t)
	if os.Getenv("OPNSENSE_URI") != "https://host.docker.internal:18444" {
		t.Fatal("FRR acceptance requires the exact isolated container forwarding endpoint")
	}
	if !regexp.MustCompile(`^http://host\.docker\.internal:[0-9]+/snapshot$`).MatchString(frrRequiredBinding(t, "OPNSENSE_FRR_ORACLE_URI")) {
		t.Fatal("requires the fixed private host observation oracle")
	}
	if !regexp.MustCompile(`^[0-9a-f-]{36}$`).MatchString(frrRequiredBinding(t, "OPNSENSE_DISPOSABLE_VM_UUID")) || frrRequiredBinding(t, "OPNSENSE_FRR_IDENTITY_ATTESTATION") != frrRequiredBinding(t, "OPNSENSE_DISPOSABLE_VM_NAME")+":"+os.Getenv("OPNSENSE_DISPOSABLE_VM_UUID") {
		t.Fatal("requires exact disposable identity")
	}
	endpointCheck := exec.Command("python3", "-c", `
import hashlib,re,socket,ssl,sys,urllib.parse
uri=urllib.parse.urlsplit(sys.stdin.readline().strip()); expected=sys.stdin.readline().strip()
if uri.scheme!="https" or not uri.hostname or not re.fullmatch(r"[0-9a-f]{64}",expected): raise SystemExit(1)
ctx=ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT);ctx.check_hostname=False;ctx.verify_mode=ssl.CERT_NONE
with socket.create_connection((uri.hostname,uri.port or 443),timeout=15) as sock:
 with ctx.wrap_socket(sock,server_hostname=uri.hostname) as tls:
  if hashlib.sha256(tls.getpeercert(binary_form=True)).hexdigest()!=expected: raise SystemExit(1)
`)
	endpointCheck.Stdin = strings.NewReader(os.Getenv("OPNSENSE_URI") + "\n" + os.Getenv("OPNSENSE_FRR_TLS_SHA256") + "\n")
	if endpointCheck.Run() != nil {
		t.Fatal("exact container API endpoint TLS attestation failed")
	}
	if os.Getenv("OPNSENSE_FRR_IPV6_READY_ATTESTATION") != frrRequiredBinding(t, "OPNSENSE_DISPOSABLE_VM_UUID")+":vtnet0:2001:db8:2::15:linklocal:ready" {
		t.Fatal("requires independently attested isolated WAN IPv6 next-hop readiness")
	}
	t.Log("FRR_LAB_IPV6_READY interface=vtnet0 global=2001:db8:2::15 linklocal=true provisioning=parent_lab_not_provider")
	observed := frrObserve(t)
	if observed.Collector.DefaultObserved {
		t.Fatal("collector already observed default before fixture")
	}
	t.Logf("FRR_IDENTITY core=%s plugin=%s daemon=%s collector_sha256=%s", frrVersionToken(observed.Core.Stdout, "OPNsense"), strings.TrimSpace(observed.Plugin.Stdout), frrVersionToken(observed.Daemon.Stdout, "bgpd version"), observed.CollectorSHA)
}

func TestAccFRRCompatibility(t *testing.T) {
	if os.Getenv("OPNSENSE_FRR_COMPATIBILITY") != "1" {
		t.Skip("FRR compatibility requires the digest-bound isolated runner")
	}
	frrPreflight(t)
	marker := "tf31h_" + fmt.Sprint(time.Now().UnixNano())
	general, bgp := frrRawSettings(t, "general"), frrRawSettings(t, "bgp")
	if string(general["enabled"]) != `"0"` || string(bgp["enabled"]) != `"0"` {
		t.Fatal("acceptance requires disabled baseline FRR and BGP")
	}
	baseline := frrObserve(t)
	if strings.Contains(baseline.Config.Stdout, "router bgp") {
		t.Fatal("requires parent-restored canonical disabled generated-config baseline")
	}
	if baseline.Collector.Active || len(baseline.Collector.Prefixes) > 0 {
		t.Fatal("collector baseline must have no active session or current prefixes")
	}
	workingDir, err := os.MkdirTemp(os.Getenv("OPNSENSE_FRR_ARTIFACT_DIR"), "terraform-core-")
	if err != nil {
		t.Fatal(err)
	}
	kinds := []string{"neighbor", "peergroup", "redistribution", "routemap", "prefixlist", "aspath", "communitylist"}
	for _, kind := range kinds {
		if len(frrTagged(t, kind, "")) != 0 {
			t.Fatal("isolated fixture requires every BGP object namespace initially empty")
		}
	}
	phases := []func(){func() { frrAssertZeros(t, marker, "BEFORE_ROLLBACK") }}
	for _, kind := range kinds {
		phases = append(phases, func() {
			for _, id := range frrTagged(t, kind, marker) {
				result := frrRaw(t, "POST", "/api/quagga/bgp/del_"+kind+"/"+id, nil)
				if string(result["result"]) != `"deleted"` && string(result["result"]) != `"saved"` {
					t.Errorf("best-effort cleanup %s rejected", kind)
				}
			}
		})
	}
	phases = append(phases,
		func() { frrRestoreSettings(t, "bgp", bgp) },
		func() { frrRestoreSettings(t, "general", general) },
		func() {
			if string(frrRaw(t, "POST", "/api/quagga/service/reconfigure", nil)["status"]) != `"ok"` {
				t.Error("baseline FRR reload failed")
			}
		},
		func() { frrLabDisabledConfigSync(t) },
		func() { frrLabFilterReload(t) },
		func() {
			restored := frrObserve(t)
			if !reflect.DeepEqual(general, frrRawSettings(t, "general")) || !reflect.DeepEqual(bgp, frrRawSettings(t, "bgp")) || baseline.Config.Stdout != restored.Config.Stdout || baseline.PF.Stdout != restored.PF.Stdout || baseline.AllPF.Stdout != restored.AllPF.Stdout || baseline.Rendered.Stdout != restored.Rendered.Stdout || baseline.Routes.Stdout != restored.Routes.Stdout {
				t.Error("exact settings/generated config/PF/routing baseline restoration failed")
			} else {
				t.Log("FRR_RESTORE settings_equal=true config_equal=true pf_equal=true rendered_pf_equal=true routes_equal=true")
			}
		},
	)
	for _, kind := range kinds {
		phases = append(phases, func() { frrAssertNamespaceZero(t, kind, "FINAL_CLEANUP") })
	}
	phases = append(phases, func() {
		if frrObserve(t).Collector.DefaultObserved {
			t.Error("collector observed forbidden default during fixture")
		}
	})
	frrRegisterCleanup(t, phases...)

	frrLegacyCoreUpgrade(t, marker)
	config := frrCompatibilityConfig(marker, true, false)
	var advancedID string
	sequenceIDs := map[string]string{}
	check := func(s *terraform.State) error {
		owner := s.RootModule().Resources["opnsense_quagga_bgp_neighbor.advanced"]
		if owner == nil {
			return fmt.Errorf("advanced owner missing")
		}
		if advancedID != "" && advancedID != owner.Primary.ID {
			return fmt.Errorf("advanced identity churn")
		}
		advancedID = owner.Primary.ID
		for _, name := range []string{"a", "b"} {
			object := s.RootModule().Resources["opnsense_quagga_route_map."+name]
			if object == nil {
				return fmt.Errorf("route-map owner missing")
			}
			if previous := sequenceIDs[name]; previous != "" && previous != object.Primary.ID {
				return fmt.Errorf("route-map identity changed")
			}
			sequenceIDs[name] = object.Primary.ID
		}
		if err := frrAdvancedLiveCheck(t, s); err != nil {
			return err
		}
		return frrLiveCheck(t, s, marker, true, baseline.Collector.Sessions)
	}
	steps := []resource.TestStep{{Config: config, Check: check}, {Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}}}
	for _, name := range []string{"opnsense_quagga_prefix_list.a", "opnsense_quagga_prefix_list.b", "opnsense_quagga_route_map.a", "opnsense_quagga_route_map.b", "opnsense_quagga_bgp_peergroup.policy", "opnsense_quagga_bgp_neighbor.peer", "opnsense_quagga_bgp_neighbor.advanced", "opnsense_quagga_bgp_neighbor.reflector", "opnsense_quagga_bgp_redistribution.policy"} {
		steps = append(steps, resource.TestStep{ResourceName: name, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"linked_prefixlist_out", "linked_routemap_out", "match_prefix", "route_map", "password"}, ImportStateCheck: frrCheckImportedPolicy(t, marker, name)})
	}
	reordered := strings.ReplaceAll(strings.ReplaceAll(config, "order = 10", "order = 30"), "order = 20", "order = 5")
	steps = append(steps, resource.TestStep{Config: reordered, Check: func(s *terraform.State) error {
		observation := frrObserve(t)
		generated := observation.Config.Stdout
		if err := frrRenderedSequenceValues(generated, marker+"_rm"); err != nil {
			return err
		}
		a, b := s.RootModule().Resources["opnsense_quagga_route_map.a"], s.RootModule().Resources["opnsense_quagga_route_map.b"]
		if a == nil || b == nil || a.Primary.ID == b.Primary.ID || a.Primary.ID != sequenceIDs["a"] || b.Primary.ID != sequenceIDs["b"] {
			return fmt.Errorf("sequence update lost route-map identities")
		}
		if err := frrCheckImportedPolicy(t, marker, "opnsense_quagga_bgp_peergroup.policy")([]*terraform.InstanceState{s.RootModule().Resources["opnsense_quagga_bgp_peergroup.policy"].Primary}); err != nil {
			return err
		}
		deadline := time.Now().Add(180 * time.Second)
		for {
			observation = frrObserve(t)
			if observation.Collector.DefaultObserved {
				return fmt.Errorf("reversed policy exported forbidden default")
			}
			if observation.Collector.Active && !slices.Contains(observation.Collector.Prefixes, "198.18.31.0/24") {
				break
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("reversed policy did not withdraw current approved prefix with active session")
			}
			time.Sleep(time.Second)
		}
		t.Log("FRR_ADVANCED_SEQUENCE rendered_sequence_values=true effective_numeric_order=true grouped_reference=true")
		return nil
	}}, resource.TestStep{Config: config, Check: check})
	steps = append(steps, resource.TestStep{Config: config, ExpectNonEmptyPlan: true, Check: func(s *terraform.State) error {
		owner := s.RootModule().Resources["opnsense_quagga_bgp_neighbor.peer"]
		result := frrRaw(t, "POST", "/api/quagga/bgp/set_neighbor/"+owner.Primary.ID, map[string]any{"neighbor": map[string]string{"description": marker + "_drift", "family": "ipv6"}})
		advanced := s.RootModule().Resources["opnsense_quagga_bgp_neighbor.advanced"]
		drift := frrRaw(t, "POST", "/api/quagga/bgp/set_neighbor/"+advanced.Primary.ID, map[string]any{"neighbor": map[string]string{"weight": "9", "attributeunchanged": "med"}})
		if string(drift["result"]) != `"saved"` {
			return fmt.Errorf("advanced drift rejected")
		}
		if string(result["result"]) != `"saved"` {
			return fmt.Errorf("raw drift mutation rejected")
		}
		return nil
	}}, resource.TestStep{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("opnsense_quagga_bgp_neighbor.peer", plancheck.ResourceActionUpdate)}}, Check: check}, resource.TestStep{Config: frrCompatibilityConfig(marker, false, true), Check: func(s *terraform.State) error { return frrLiveCheck(t, s, marker, false, baseline.Collector.Sessions) }}, resource.TestStep{Config: config, Check: check})
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, WorkingDir: workingDir, CheckDestroy: func(_ *terraform.State) error { frrAssertZeros(t, marker, "TERRAFORM_DESTROY"); return nil }, Steps: steps})
}

// Imports may canonicalize a sibling UUID within a name/number group. Verify
// every ignored reference independently through the authoritative raw API.
func frrCheckImportedPolicy(t *testing.T, marker, name string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 || states[0].ID == "" {
			return fmt.Errorf("import did not return one identified object")
		}
		expected := map[string]string{
			"opnsense_quagga_route_map.a":               "match_prefix.#",
			"opnsense_quagga_bgp_peergroup.policy":      "linked_routemap_out.#",
			"opnsense_quagga_bgp_neighbor.peer":         "peer_group",
			"opnsense_quagga_bgp_redistribution.policy": "route_map",
		}[name]
		if expected != "" && (states[0].Attributes[expected] == "" || states[0].Attributes[expected] == "0") {
			return fmt.Errorf("import lost configured policy reference")
		}
		kind := map[string]string{"opnsense_quagga_route_map.a": "routemap", "opnsense_quagga_bgp_peergroup.policy": "peergroup", "opnsense_quagga_bgp_neighbor.peer": "neighbor", "opnsense_quagga_bgp_redistribution.policy": "redistribution"}[name]
		if kind != "" {
			raw := frrRaw(t, "GET", "/api/quagga/bgp/get_"+kind+"/"+states[0].ID, nil)
			var owner map[string]json.RawMessage
			if json.Unmarshal(raw[kind], &owner) != nil {
				return fmt.Errorf("canonical owner read failed")
			}
			if err := frrCanonicalReferences(states[0].Attributes, owner); err != nil {
				return err
			}
		}
		for key, id := range states[0].Attributes {
			kind, group := "", ""
			switch {
			case strings.HasPrefix(key, "match_prefix.") && key != "match_prefix.#", strings.HasPrefix(key, "linked_prefixlist_out.") && key != "linked_prefixlist_out.#":
				kind, group = "prefixlist", marker+"_pl"
			case strings.HasPrefix(key, "linked_routemap_out.") && key != "linked_routemap_out.#", key == "route_map":
				kind, group = "routemap", marker+"_rm"
			}
			if kind == "" || id == "" {
				continue
			}
			raw := frrRaw(t, "GET", "/api/quagga/bgp/get_"+kind+"/"+id, nil)
			var object struct {
				Name string `json:"name"`
			}
			if json.Unmarshal(raw[kind], &object) != nil || object.Name != group {
				return fmt.Errorf("imported policy reference lost authoritative group identity")
			}
		}
		return nil
	}
}

func frrLiveCheck(t *testing.T, s *terraform.State, marker string, firewall bool, baselineSessions int) error {
	t.Helper()
	frrLabFilterReload(t)
	deadline := time.Now().Add(180 * time.Second)
	var observation frrObservation
	for {
		observation = frrObserve(t)
		if observation.Collector.Active && observation.Collector.Sessions > baselineSessions && strings.Contains(strings.Join(observation.Collector.Prefixes, ","), "198.18.31.0/24") {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("collector did not receive positive approved prefix")
		}
		time.Sleep(time.Second)
	}
	for _, prefix := range observation.Collector.Prefixes {
		if prefix != "198.18.31.0/24" {
			return fmt.Errorf("collector observed a non-allowlisted prefix")
		}
	}
	if observation.Collector.DefaultObserved {
		return fmt.Errorf("collector observed forbidden default")
	}
	config := observation.Config.Stdout
	for _, fragment := range []string{"address-family ipv4 unicast", "address-family ipv6 unicast", "route-map " + marker + "_rm permit 10", "route-map " + marker + "_rm deny 20", "neighbor " + marker + "_pg next-hop-self", "neighbor " + marker + "_pg route-map " + marker + "_rm out"} {
		if !strings.Contains(config, fragment) {
			return fmt.Errorf("independent generated configuration missing %s", fragment)
		}
	}
	if err := frrAFPolicyProof(config, marker, !firewall); err != nil {
		return err
	}
	if !strings.Contains(config, "neighbor 10.0.2.2 peer-group "+marker+"_pg") || strings.Contains(config, "neighbor 10.0.2.2 route-map ") {
		return fmt.Errorf("active peer does not exclusively inherit group policy")
	}
	if err := frrBGPFirewallProof(observation.AllPF.Stdout, observation.Rendered.Stdout, firewall); err != nil {
		return err
	}
	if strings.Contains(config, "neighbor 10.0.2.2 default-originate") {
		return fmt.Errorf("active collector peer unexpectedly originates default")
	}
	// Data sources may return canonical sibling UUIDs; all other state must
	// agree with the managed owner and both reference forms must resolve to the
	// configured authoritative group, not merely be nonempty.
	for _, name := range []string{"opnsense_quagga_bgp_neighbor.peer", "opnsense_quagga_bgp_peergroup.policy", "opnsense_quagga_route_map.a"} {
		managed, lookup := s.RootModule().Resources[name], s.RootModule().Resources["data."+name]
		if managed == nil || lookup == nil {
			return fmt.Errorf("resource/data-source fixture missing")
		}
		if err := frrCompareAdvancedPublicAttributes(managed.Primary.Attributes, lookup.Primary.Attributes, "linked_routemap_out.", "match_prefix."); err != nil {
			return err
		}
		if err := frrCheckImportedPolicy(t, marker, name)([]*terraform.InstanceState{lookup.Primary}); err != nil {
			return err
		}
	}
	owner := s.RootModule().Resources["opnsense_quagga_bgp_neighbor.peer"]
	raw := frrRaw(t, "GET", "/api/quagga/bgp/get_neighbor/"+owner.Primary.ID, nil)
	var neighbor map[string]json.RawMessage
	if json.Unmarshal(raw["neighbor"], &neighbor) != nil {
		return fmt.Errorf("raw neighbor read missing")
	}
	membership, err := frrRawSelection(neighbor["peergroup"])
	if err != nil {
		return err
	}
	group := s.RootModule().Resources["opnsense_quagga_bgp_peergroup.policy"]
	if membership == "" || membership != group.Primary.ID || owner.Primary.Attributes["peer_group"] != membership {
		return fmt.Errorf("raw active peer-group membership differs")
	}
	var description string
	_ = json.Unmarshal(neighbor["description"], &description)
	if description != marker {
		return fmt.Errorf("drift repair did not restore description")
	}
	for _, name := range []string{"v4", "v6", "dual", "cross"} {
		owner := s.RootModule().Resources["opnsense_quagga_bgp_neighbor."+name]
		if owner == nil {
			return fmt.Errorf("legacy/explicit neighbor fixture missing")
		}
		expected := map[string]string{"v4": "ipv4", "v6": "ipv6", "dual": "ipv4,ipv6", "cross": "ipv6"}[name]
		family := []string{}
		for k, v := range owner.Primary.Attributes {
			if strings.HasPrefix(k, "family.") && k != "family.#" {
				family = append(family, v)
			}
		}
		sort.Strings(family)
		fixture := frrRaw(t, "GET", "/api/quagga/bgp/get_neighbor/"+owner.Primary.ID, nil)
		var rawFixture map[string]json.RawMessage
		if json.Unmarshal(fixture["neighbor"], &rawFixture) != nil {
			return fmt.Errorf("raw disabled family fixture missing")
		}
		selected, err := frrRawSelection(rawFixture["family"])
		if err != nil {
			return err
		}
		selectedFamilies := strings.Split(selected, ",")
		sort.Strings(selectedFamilies)
		if strings.Join(selectedFamilies, ",") != expected || string(rawFixture["enabled"]) != `"0"` {
			return fmt.Errorf("raw disabled family fixture differs for %s", name)
		}
		if strings.Join(family, ",") != expected {
			return fmt.Errorf("legacy or explicit family state wrong for %s", name)
		}
	}
	t.Logf("FRR_EFFECTIVE_CONFIG policy_order=true peer_group_per_af=true positive_prefix=198.18.31.0/24 default_export=false firewall_rules=%t activation=explicit_lab_filter_reload", firewall)
	return nil
}

func frrCompatibilityConfig(marker string, firewall, defaultOriginate bool) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_general" "lab" {
 enabled = true
 firewall_rules = %[2]t
 depends_on = [opnsense_quagga_bgp_global.lab]
}
resource "opnsense_quagga_bgp_global" "lab" {
 enabled = true
 as_number = 65111
 router_id = "198.18.31.1"
 networks = ["198.18.31.0/24", "0.0.0.0/0"]
}
resource "opnsense_quagga_prefix_list" "a" {
 name = "%[1]s_pl"
 version = "IPv4"
 sequence = 10
 action = "permit"
 network = "198.18.31.0/24"
}
resource "opnsense_quagga_prefix_list" "b" {
 name = "%[1]s_pl"
 version = "IPv4"
 sequence = 20
 action = "deny"
 network = "0.0.0.0/0 le 32"
 depends_on = [opnsense_quagga_prefix_list.a]
}
resource "opnsense_quagga_route_map" "a" {
 name = "%[1]s_rm"
 action = "permit"
 order = 10
 match_prefix = [opnsense_quagga_prefix_list.b.id]
}
resource "opnsense_quagga_route_map" "b" {
 name = "%[1]s_rm"
 action = "deny"
 order = 20
 depends_on = [opnsense_quagga_route_map.a]
}
resource "opnsense_quagga_bgp_redistribution" "policy" {
 enabled = false
 description = "%[1]s_redistribution"
 redistribute = "connected"
 route_map = opnsense_quagga_route_map.b.id
}
resource "opnsense_quagga_bgp_peergroup" "policy" {
 name = "%[1]s_pg"
 remote_as = 65222
 family = ["ipv4","ipv6"]
 next_hop_self = true
 default_originate = false
 linked_routemap_out = [opnsense_quagga_route_map.b.id]
}
resource "opnsense_quagga_bgp_peergroup" "unused" {
 name = "%[1]s_unused"
 listen_ranges = ["192.0.2.0/24","2001:db8:31::/64"]
 remote_as = 65222
 family = ["ipv4","ipv6"]
 default_originate = %[3]t
}
resource "opnsense_quagga_bgp_neighbor" "peer" {
 address = "10.0.2.2"
 remote_as = 65222
 description = "%[1]s"
 family = ["ipv4","ipv6"]
 peer_group = opnsense_quagga_bgp_peergroup.policy.id
 depends_on = [opnsense_quagga_general.lab]
}
resource "opnsense_quagga_bgp_neighbor" "advanced" {
 address = "2001:db8::34"
 description = "%[1]s_advanced"
 family = ["ipv6"]
 remote_as_mode = "external"
 local_as = 64513
 password = "acceptance-only-input"
 weight = 0
 local_ip = "2001:db8:2::15"
 link_local_interface = "wan"
 next_hop_self = true
 next_hop_self_all = true
 multihop = true
 route_reflector_client = false
 soft_reconfiguration_inbound = true
 bfd = true
 bfd_strict = true
 connect_timer = 30
 default_originate = true
 as_override = true
 remove_private_as = "remove-private-AS all replace-AS"
 allow_as_in = "origin"
 disable_connected_check = true
 attribute_unchanged = "as-path"
 capabilities = ["extended-nexthop","dynamic","fqdn","link-local"]
 depends_on = [opnsense_quagga_general.lab]
}
resource "opnsense_quagga_bgp_neighbor" "reflector" {
 address = "2001:db8::35"
 description = "%[1]s_reflector"
 remote_as_mode = "internal"
 family = ["ipv6"]
 route_reflector_client = true
 depends_on = [opnsense_quagga_general.lab]
}
resource "opnsense_quagga_bgp_neighbor" "v4" {
 enabled=false
 address="192.0.2.31"
 remote_as=65222
 multi_protocol=false
 description="%[1]s_v4"
}
resource "opnsense_quagga_bgp_neighbor" "v6" {
 enabled=false
 address="2001:db8::31"
 remote_as=65222
 multi_protocol=false
 description="%[1]s_v6"
}
resource "opnsense_quagga_bgp_neighbor" "dual" {
 enabled=false
 address="192.0.2.32"
 remote_as=65222
 multi_protocol=true
 description="%[1]s_dual"
}
resource "opnsense_quagga_bgp_neighbor" "cross" {
 enabled=false
 address="192.0.2.33"
 remote_as=65222
 family=["ipv6"]
 description="%[1]s_cross"
}
data "opnsense_quagga_bgp_neighbor" "peer" {id=opnsense_quagga_bgp_neighbor.peer.id}
data "opnsense_quagga_bgp_neighbor" "advanced" {id=opnsense_quagga_bgp_neighbor.advanced.id}
data "opnsense_quagga_bgp_peergroup" "policy" {id=opnsense_quagga_bgp_peergroup.policy.id}
data "opnsense_quagga_route_map" "a" {id=opnsense_quagga_route_map.a.id}
`, marker, firewall, defaultOriginate)
}

// Real Terraform Core obtains version-0 state through the unmodified baseline
// provider's supported import, then upgrades with the frozen candidate. No
// manual state editing or replacement of a user's resource address is involved.
func frrLegacyCoreUpgrade(t *testing.T, marker string) {
	t.Helper()
	dir, err := os.MkdirTemp(os.Getenv("OPNSENSE_FRR_ARTIFACT_DIR"), "core-v0-")
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	config := fmt.Sprintf(`terraform {
 required_providers { opnsense = { source = "matthew-on-git/opnsense" } }
}
provider "opnsense" {}
resource "opnsense_quagga_bgp_neighbor" "legacy" {
 enabled = false
 address = "2001:db8::40"
 remote_as = 65222
 multi_protocol = false
 description = %q
}
`, marker+"_core_v0")
	if root.WriteFile("main.tf", []byte(config), 0o600) != nil {
		t.Fatal("cannot write Core migration fixture")
	}
	artifactDir := os.Getenv("OPNSENSE_FRR_ARTIFACT_DIR")
	for _, kind := range []string{"legacy", "current"} {
		cli := fmt.Sprintf("provider_installation {\n dev_overrides {\n  \"registry.terraform.io/matthew-on-git/opnsense\" = %q\n }\n direct {}\n}\n", artifactDir+"/bin/"+kind)
		if root.WriteFile(kind+".tfrc", []byte(cli), 0o600) != nil {
			t.Fatal("cannot write Core provider override")
		}
	}
	var id string
	run := func(kind, action string) {
		t.Helper()
		var cmd *exec.Cmd
		switch action {
		case "import":
			cmd = exec.Command("terraform", "import", "-no-color", "opnsense_quagga_bgp_neighbor.legacy")
			cmd.Args = append(cmd.Args, id)
		case "refresh":
			cmd = exec.Command("terraform", "apply", "-refresh-only", "-auto-approve", "-no-color")
		case "saved-plan":
			cmd = exec.Command("terraform", "plan", "-detailed-exitcode", "-out=candidate.tfplan", "-no-color")
		case "apply":
			cmd = exec.Command("terraform", "apply", "-auto-approve", "-no-color", "candidate.tfplan")
		case "plan":
			cmd = exec.Command("terraform", "plan", "-detailed-exitcode", "-no-color")
		case "destroy":
			cmd = exec.Command("terraform", "destroy", "-auto-approve", "-no-color")
		default:
			t.Fatal("unsupported Core fixture command")
		}
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "TF_CLI_CONFIG_FILE="+dir+"/"+kind+".tfrc", "TF_IN_AUTOMATION=1")
		log, err := root.OpenFile("core-migration.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			t.Fatal("cannot retain Core log")
		}
		cmd.Stdout = log
		cmd.Stderr = log
		runErr := cmd.Run()
		_ = log.Close()
		if runErr != nil {
			t.Fatalf("Terraform Core %s %s failed; private scanned log retained", kind, action)
		}
	}
	added := frrRaw(t, "POST", "/api/quagga/bgp/add_neighbor", map[string]any{"neighbor": map[string]string{"enabled": "0", "address": "2001:db8::40", "remoteas": "65222", "family": "ipv6", "description": marker + "_core_v0"}})
	if string(added["result"]) != `"saved"` || json.Unmarshal(added["uuid"], &id) != nil || !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).MatchString(id) {
		t.Fatal("legacy import fixture create did not affirm identity")
	}
	run("legacy", "import")
	type instance struct {
		SchemaVersion int                        `json:"schema_version"`
		Attributes    map[string]json.RawMessage `json:"attributes"`
	}
	type state struct {
		Resources []struct {
			Type, Name string
			Instances  []instance
		} `json:"resources"`
	}
	readState := func() instance {
		raw, err := root.ReadFile("terraform.tfstate")
		if err != nil {
			t.Fatal("Core state unavailable")
		}
		var decoded state
		if json.Unmarshal(raw, &decoded) != nil {
			t.Fatal("Core state malformed")
		}
		for _, resource := range decoded.Resources {
			if resource.Type == "opnsense_quagga_bgp_neighbor" && resource.Name == "legacy" && len(resource.Instances) == 1 {
				return resource.Instances[0]
			}
		}
		t.Fatal("Core migration resource address or identity lost")
		return instance{}
	}
	before := readState()
	if before.SchemaVersion != 0 || string(before.Attributes["id"]) != fmt.Sprintf("%q", id) {
		t.Fatal("baseline provider did not persist actual version-0 identity")
	}
	run("current", "refresh")
	after := readState()
	var upgradedFamilies []string
	familyErr := json.Unmarshal(after.Attributes["family"], &upgradedFamilies)
	if after.SchemaVersion != 1 || string(after.Attributes["id"]) != fmt.Sprintf("%q", id) || (familyErr != nil || len(upgradedFamilies) != 1 || upgradedFamilies[0] != "ipv6") {
		t.Fatal("real Core migration did not preserve IPv6 intent and UUID")
	}
	run("current", "saved-plan")
	run("current", "apply")
	run("current", "plan")
	run("current", "destroy")
	t.Log("FRR_CORE_V0_UPGRADE actual_legacy_import=true offline_upgrade=true uuid_preserved=true ipv6_only=true no_state_edit=true no_change_plan=true terraform_delete=true")
}

func frrBGPFirewallProof(rules, rendered string, enabled bool) error {
	in4, out4, in6, out6 := false, false, false, false
	count := 0
	for _, line := range strings.Split(rules, "\n") {
		if !strings.Contains(line, "proto tcp") || (!regexp.MustCompile(`\bport = (bgp|179)(\s|$)`).MatchString(line)) {
			continue
		}
		count++
		ipv4, ipv6 := strings.Contains(line, " inet "), strings.Contains(line, " inet6 ")
		if !ipv4 && !ipv6 {
			ipv4 = true
			ipv6 = true
		}
		if strings.Contains(line, " on ") {
			continue
		}
		toSelf := strings.Contains(line, "from any to (self) port = ") || strings.Contains(line, "from any to self port = ")
		fromSelf := strings.Contains(line, "from (self) to any port = ") || strings.Contains(line, "from self to any port = ")
		if strings.HasPrefix(line, "pass in ") && toSelf {
			in4 = in4 || ipv4
			in6 = in6 || ipv6
		}
		if strings.HasPrefix(line, "pass out ") && fromSelf {
			out4 = out4 || ipv4
			out6 = out6 || ipv6
		}
	}
	if !enabled {
		if count != 0 || strings.Contains(rendered, "# Pass BGP") {
			return fmt.Errorf("effective/rendered BGP firewall opt-out failed")
		}
		return nil
	}
	broadIn4, broadOut4, broadIn6, broadOut6 := false, false, false, false
	for _, line := range strings.Split(rendered, "\n") {
		if !strings.Contains(line, "# Pass BGP") || !strings.Contains(line, "proto tcp") || !strings.Contains(line, "179") {
			continue
		}
		if strings.Contains(line, " on ") {
			continue
		}
		// The core renderer may expand source inet46 into separate inet/inet6
		// rules. Require the complete union in both directions, not one syntax.
		ipv4, ipv6 := !strings.Contains(line, " inet6 "), !strings.Contains(line, " inet ")
		if strings.HasPrefix(line, "pass in ") && strings.Contains(line, "from {any} to {(self)} port {179}") {
			broadIn4 = broadIn4 || ipv4
			broadIn6 = broadIn6 || ipv6
		}
		if strings.HasPrefix(line, "pass out ") && strings.Contains(line, "from {(self)} to {any} port {179}") {
			broadOut4 = broadOut4 || ipv4
			broadOut6 = broadOut6 || ipv6
		}
	}
	if !in4 || !out4 || !in6 || !out6 || !broadIn4 || !broadOut4 || !broadIn6 || !broadOut6 {
		return fmt.Errorf("independent generated/effective inet46 TCP179 to/from-self firewall proof incomplete")
	}
	return nil
}

func frrCanonicalReferences(attrs map[string]string, raw map[string]json.RawMessage) error {
	for field, wire := range map[string]string{"linked_routemap_out": "linkedRoutemapOut", "match_prefix": "match2", "route_map": "linkedRoutemap", "peer_group": "peergroup"} {
		if _, present := attrs[field]; !present {
			if _, present = attrs[field+".#"]; !present {
				continue
			}
		}
		selected, err := frrRawSelection(raw[wire])
		if err != nil {
			return err
		}
		want := []string{}
		if value, scalar := attrs[field]; scalar {
			if value != "" {
				want = append(want, value)
			}
		} else {
			for key, value := range attrs {
				if strings.HasPrefix(key, field+".") && key != field+".#" {
					want = append(want, value)
				}
			}
		}
		got := []string{}
		if selected != "" {
			got = strings.Split(selected, ",")
		}
		sort.Strings(got)
		sort.Strings(want)
		if !reflect.DeepEqual(want, got) {
			return fmt.Errorf("fresh import/data-source reference %s is not raw canonical selection", field)
		}
	}
	return nil
}

func frrAFPolicyProof(config, marker string, unusedDefault bool) error {
	blocks := map[string][]string{}
	af := ""
	for _, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "address-family ") {
			af = strings.TrimPrefix(line, "address-family ")
			continue
		}
		if line == "exit-address-family" {
			af = ""
			continue
		}
		if line == "neighbor "+marker+"_pg default-originate" {
			return fmt.Errorf("active group may not originate default")
		}
		if af == "" && (line == "neighbor "+marker+"_pg next-hop-self" || strings.HasPrefix(line, "neighbor "+marker+"_pg route-map ") || line == "neighbor "+marker+"_unused default-originate") {
			return fmt.Errorf("group AF policy outside address-family block")
		}
		if af != "" {
			blocks[af] = append(blocks[af], line)
		}
	}
	for _, family := range []string{"ipv4 unicast", "ipv6 unicast"} {
		lines := blocks[family]
		for _, expected := range []string{"neighbor " + marker + "_pg next-hop-self", "neighbor " + marker + "_pg route-map " + marker + "_rm out"} {
			count := 0
			for _, line := range lines {
				if line == expected {
					count++
				}
			}
			if count != 1 {
				return fmt.Errorf("group policy missing or duplicated in %s", family)
			}
		}
		active, unused := 0, 0
		for _, line := range lines {
			if line == "neighbor "+marker+"_pg default-originate" {
				active++
			}
			if line == "neighbor "+marker+"_unused default-originate" {
				unused++
			}
		}
		if active != 0 || unused != map[bool]int{true: 1, false: 0}[unusedDefault] {
			return fmt.Errorf("per-AF default-originate mismatch in %s", family)
		}
	}
	return nil
}

// Advanced observations never print generated configuration or credentials.
func frrAdvancedLiveCheck(t *testing.T, s *terraform.State) error {
	t.Helper()
	owner := s.RootModule().Resources["opnsense_quagga_bgp_neighbor.advanced"]
	lookup := s.RootModule().Resources["data.opnsense_quagga_bgp_neighbor.advanced"]
	if owner == nil || lookup == nil {
		return fmt.Errorf("advanced owner/lookup missing")
	}
	if _, exists := lookup.Primary.Attributes["password"]; exists {
		return fmt.Errorf("lookup exposed authentication")
	}
	if owner.Primary.Attributes["password"] != "acceptance-only-input" || owner.Primary.Attributes["weight"] != "0" || owner.Primary.Attributes["attribute_unchanged"] != "as-path" {
		return fmt.Errorf("advanced configured state did not converge")
	}
	if err := frrCompareAdvancedPublicAttributes(owner.Primary.Attributes, lookup.Primary.Attributes); err != nil {
		return err
	}
	observation := frrObserve(t)
	config := observation.Config.Stdout
	for _, fragment := range []string{"neighbor 2001:db8::34 remote-as external", "neighbor 2001:db8::34 local-as 64513", "neighbor 2001:db8::34 weight 0", "neighbor 2001:db8::34 interface vtnet0", "neighbor 2001:db8::34 bfd strict", "neighbor 2001:db8::34 ebgp-multihop 255", "neighbor 2001:db8::34 timers connect 30", "neighbor 2001:db8::34 attribute-unchanged as-path", "neighbor 2001:db8::34 capability dynamic", "neighbor 2001:db8::34 disable-connected-check"} {
		if !strings.Contains(config, fragment) {
			return fmt.Errorf("advanced generated transport/listen settings incomplete")
		}
	}
	af := ""
	reflectorCount := 0
	for _, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "address-family ") {
			af = strings.TrimPrefix(line, "address-family ")
		}
		if line == "exit-address-family" {
			af = ""
		}
		if line == "neighbor 2001:db8::35 route-reflector-client" {
			reflectorCount++
			if af != "ipv6 unicast" {
				return fmt.Errorf("reflector escaped IPv6 address family")
			}
		}
		if strings.HasPrefix(line, "neighbor 2001:db8::34 ") && (strings.Contains(line, "next-hop-self") || strings.Contains(line, "route-reflector-client") || strings.Contains(line, "soft-reconfiguration") || strings.Contains(line, "default-originate") || strings.Contains(line, "as-override") || strings.Contains(line, "remove-private-AS") || strings.Contains(line, "allowas-in")) {
			if af != "ipv6 unicast" {
				return fmt.Errorf("advanced AF flag escaped selected IPv6 family")
			}
		}
	}
	if reflectorCount != 1 {
		return fmt.Errorf("reflector AF setting missing or duplicated")
	}
	for _, capability := range []string{"extended-nexthop", "dynamic", "fqdn", "link-local"} {
		if !strings.Contains(config, "neighbor 2001:db8::34 capability "+capability+"\n") {
			return fmt.Errorf("configured capability missing from generated settings")
		}
	}
	group := s.RootModule().Resources["opnsense_quagga_bgp_peergroup.unused"]
	if group == nil {
		return fmt.Errorf("listen-range owner missing")
	}
	for _, prefix := range []string{"192.0.2.0/24", "2001:db8:31::/64"} {
		if !strings.Contains(config, "bgp listen range "+prefix+" peer-group "+group.Primary.Attributes["name"]+"\n") {
			return fmt.Errorf("listen range not bound to intended peer-group")
		}
	}
	for _, suffix := range []string{"next-hop-self all", "soft-reconfiguration inbound", "default-originate", "as-override", "remove-private-AS all replace-AS", "allowas-in origin"} {
		if !strings.Contains(config, "neighbor 2001:db8::34 "+suffix) {
			return fmt.Errorf("advanced generated AF settings incomplete")
		}
	}
	if !strings.Contains(config, "neighbor 2001:db8::35 remote-as internal") || !strings.Contains(config, "neighbor 2001:db8::35 route-reflector-client") {
		return fmt.Errorf("mode-only iBGP reflector generation incomplete")
	}
	t.Log("FRR_ADVANCED public_fields=19 password_lookup=false numeric_zero=true generated_transport=true generated_ipv6=true listen_ranges=2 identity_stable=true")
	return nil
}

func frrVersionToken(output, label string) string {
	fields := strings.Fields(output)
	labelFields := strings.Fields(label)
	for i := 0; i+len(labelFields) < len(fields); i++ {
		if slices.Equal(fields[i:i+len(labelFields)], labelFields) {
			token := fields[i+len(labelFields)]
			if label == "OPNsense" {
				match := regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)+(?:_[0-9]+)?)(?:-(?:amd64|aarch64))?$`).FindStringSubmatch(token)
				if match == nil {
					return ""
				}
				return match[1]
			}
			return token
		}
	}
	return ""
}

func TestFRRVersionTokensRejectPrefixes(t *testing.T) {
	for _, tc := range []struct{ output, label, want string }{{"OPNsense 26.7.8_20-amd64", "OPNsense", "26.7.8_2"}, {"bgpd version 10.7.10", "bgpd version", "10.7.1"}} {
		if frrVersionToken(tc.output, tc.label) == tc.want {
			t.Fatal("version prefix falsely accepted")
		}
	}
	if frrVersionToken("OPNsense 26.7.8_2 (amd64)", "OPNsense") != "26.7.8_2" || frrVersionToken("OPNsense 26.7.8_2-amd64", "OPNsense") != "26.7.8_2" {
		t.Fatal("exact core token not parsed")
	}
}

// The top-level flatmap % is synthetic: the resource has one more schema
// property (password) than the deliberately password-free lookup. All real
// public values, including set membership/counts and explicit empty strings,
// must still agree in both directions.
func frrCompareAdvancedPublicAttributes(owner, lookup map[string]string, ignoredPrefixes ...string) error {
	if _, exists := lookup["password"]; exists {
		return fmt.Errorf("lookup exposed authentication")
	}
	ignored := func(key string) bool {
		if key == "%" || key == "password" {
			return true
		}
		for _, prefix := range ignoredPrefixes {
			if strings.HasPrefix(key, prefix) {
				return true
			}
		}
		return false
	}

	for _, pair := range [][2]map[string]string{{owner, lookup}, {lookup, owner}} {
		for key, value := range pair[0] {
			if ignored(key) {
				continue
			}
			if got, exists := pair[1][key]; !exists || got != value {
				return fmt.Errorf("advanced public lookup differs for %s", key)
			}
		}
	}
	return nil
}

func TestFRRAdvancedPublicFlatmapComparison(t *testing.T) {
	owner := map[string]string{"%": "36", "password": "configured-test-input", "id": "stable", "weight": "0", "multihop": "false", "attribute_unchanged": "", "capabilities.#": "1", "capabilities.123": "dynamic"}
	lookup := map[string]string{"%": "35", "id": "stable", "weight": "0", "multihop": "false", "attribute_unchanged": "", "capabilities.#": "1", "capabilities.123": "dynamic"}
	if err := frrCompareAdvancedPublicAttributes(owner, lookup); err != nil {
		t.Fatal("synthetic schema counts caused false mismatch", err)
	}
	for key := range lookup {
		if key == "%" {
			continue
		}
		t.Run(key, func(t *testing.T) {
			changed := map[string]string{}
			for k, v := range lookup {
				changed[k] = v
			}
			changed[key] = "different"
			if frrCompareAdvancedPublicAttributes(owner, changed) == nil {
				t.Fatal("real public mismatch was ignored", key)
			}
		})
	}
	missing := map[string]string{}
	for key, value := range lookup {
		if key != "attribute_unchanged" {
			missing[key] = value
		}
	}
	if frrCompareAdvancedPublicAttributes(owner, missing) == nil {
		t.Fatal("missing explicit empty public value was ignored")
	}
	lookup["unexpected_public_field"] = "value"
	if frrCompareAdvancedPublicAttributes(owner, lookup) == nil {
		t.Fatal("lookup-only public value was ignored")
	}
}

func TestFRRGenericPeerPublicFlatmapComparison(t *testing.T) {
	owner := map[string]string{"%": "36", "password": "", "id": "peer", "weight": "0", "description": "", "family.#": "1", "family.123": "ipv4", "peer_group": "group", "linked_routemap_out.#": "2", "linked_routemap_out.1": "sibling-a", "linked_routemap_out.2": "sibling-b", "match_prefix.#": "2", "match_prefix.1": "prefix-a"}
	lookup := map[string]string{"%": "35", "id": "peer", "weight": "0", "description": "", "family.#": "1", "family.123": "ipv4", "peer_group": "group", "linked_routemap_out.#": "1", "linked_routemap_out.3": "canonical", "match_prefix.#": "1", "match_prefix.2": "canonical-prefix"}
	compare := func(values map[string]string) error {
		return frrCompareAdvancedPublicAttributes(owner, values, "linked_routemap_out.", "match_prefix.")
	}
	if err := compare(lookup); err != nil {
		t.Fatal("generic peer metadata/canonical relation false mismatch", err)
	}
	for _, key := range []string{"id", "weight", "description", "family.#", "family.123", "peer_group"} {
		t.Run(key, func(t *testing.T) {
			changed := map[string]string{}
			for k, v := range lookup {
				changed[k] = v
			}
			changed[key] = "different"
			if compare(changed) == nil {
				t.Fatal("generic peer real field mismatch ignored", key)
			}
		})
	}
	delete(lookup, "description")
	if compare(lookup) == nil {
		t.Fatal("generic peer missing empty real field ignored")
	}
	lookup["description"] = ""
	lookup["unexpected_public_field"] = "value"
	if compare(lookup) == nil {
		t.Fatal("generic peer lookup-only real field ignored")
	}
	delete(lookup, "unexpected_public_field")
	lookup["password"] = ""
	if compare(lookup) == nil {
		t.Fatal("generic peer password exposure ignored")
	}
}

// Released bgpd.conf at 5ac2959c8514b4f60e94d65439be5f629f10521f
// uses helpers.sortDictList(..., 'name', 'id'): textual rendering may put "30"
// before "5". FRR precedence is numeric, independently proved by the mandatory
// active-session current-prefix withdrawal check before the sequence marker.
func frrRenderedSequenceValues(config, name string) error {
	counts := map[string]int{}
	for _, line := range strings.Split(config, "\n") {
		counts[strings.TrimSpace(line)]++
	}
	if counts["route-map "+name+" permit 30"] != 1 || counts["route-map "+name+" deny 5"] != 1 {
		return fmt.Errorf("generated route-map numeric sequence values missing or duplicated")
	}
	return nil
}

func TestFRRRenderedSequenceValuesAllowReleasedLexicalOrder(t *testing.T) {
	// Source: net/frr/src/opnsense/service/templates/OPNsense/Quagga/Config/bgpd.conf
	// at the released revision above. Lexical ID order is not effective policy order.
	lexical := "route-map fixture_rm permit 30\n match ip address prefix-list fixture_pl\nroute-map fixture_rm deny 5\n"
	if err := frrRenderedSequenceValues(lexical, "fixture_rm"); err != nil {
		t.Fatal("valid released lexical rendering rejected", err)
	}
	for _, invalid := range []string{
		"route-map fixture_rm permit 30\n",
		"route-map fixture_rm deny 5\n",
		"route-map fixture_rm permit 300\nroute-map fixture_rm deny 5\n",
		"route-map fixture_rm permit 30\nroute-map fixture_rm deny 50\n",
		"route-map fixture_rm permit 30\nroute-map fixture_rm deny 5\nroute-map fixture_rm deny 5\n",
	} {
		if frrRenderedSequenceValues(invalid, "fixture_rm") == nil {
			t.Fatal("missing, wrong or duplicated numeric sequence accepted")
		}
	}
}
