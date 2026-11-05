// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
	"golang.org/x/sync/semaphore"
)

// Keep discovery and owner mutations/recovery together. The client write mutex
// additionally serializes the actual save/reload with unrelated provider owners.
var bgpOwnerMu = semaphore.NewWeighted(1)

type frrCapability struct {
	Version, Core    string
	Grouped          bool
	Fields           map[string]bool
	InterfaceOptions map[string]bool
}

// Package revisions do not change the plugin release. Keep 2.0 revisions
// eligible for modern-model validation without accepting 2.1 or other 2.x APIs.
var frrPackageVersion = regexp.MustCompile(`^(?:1\.(?:4[6-9]|5[0-5])(?:_[0-9]+)?|2\.0(?:_[0-9]+)?)$`)

func frrRead(ctx context.Context, c *opnsense.Client, endpoint string, out any) error {
	// Firmware package metadata can block behind the appliance's pkg/configd
	// backend. Bound this read-only gate, including retries/read-slot waiting,
	// rather than holding every BGP owner indefinitely. Earlier caller
	// cancellation/deadlines still take precedence.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := c.AcquireRead(ctx); err != nil {
		return err
	}
	defer c.ReleaseRead()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL()+endpoint, nil)
	if err != nil {
		return err
	}
	response, err := c.HTTPClient().Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("FRR read-only discovery %s: timed out; check firmware/package backend health before retrying", endpoint)
		}
		return fmt.Errorf("FRR read-only discovery %s: transport failure", endpoint)
	}
	defer func() { _ = response.Body.Close() }()
	if err := opnsense.CheckHTTPError(response.StatusCode, endpoint); err != nil {
		return err
	}
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("FRR read-only discovery %s: malformed JSON response", endpoint)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("FRR discovery decoder failed")
	}
	return nil
}

func discoverFRR(ctx context.Context, c *opnsense.Client) (frrCapability, error) {
	var result frrCapability
	var info struct {
		Core     string                                      `json:"product_version"`
		Packages []struct{ Name, Version, Installed string } `json:"package"`
	}
	if err := frrRead(ctx, c, "/api/core/firmware/info", &info); err != nil {
		return result, err
	}
	result.Core = info.Core
	found := false
	for _, p := range info.Packages {
		if p.Name != "os-frr" {
			continue
		}
		if found || p.Installed != "1" || !frrPackageVersion.MatchString(p.Version) {
			return result, fmt.Errorf("FRR compatibility requires installed stable os-frr 1.46–1.55 or 2.0 (including package revisions); missing, contradictory or unsupported package metadata")
		}
		found = true
		result.Version = p.Version
		result.Grouped = strings.HasPrefix(p.Version, "2.0")
	}
	if !found {
		return result, fmt.Errorf("FRR compatibility: installed os-frr package is missing; install and independently verify the plugin before managing BGP")
	}
	var template struct {
		Neighbor map[string]json.RawMessage `json:"neighbor"`
	}
	if err := frrRead(ctx, c, "/api/quagga/bgp/get_neighbor", &template); err != nil {
		return result, err
	}
	result.Fields = map[string]bool{}
	for wire := range template.Neighbor {
		result.Fields[wire] = validAdvancedTemplateField(wire, template.Neighbor[wire])
	}
	if raw, exists := template.Neighbor["linklocalinterface"]; exists {
		var options map[string]json.RawMessage
		if selected, err := decodeFRRSelection(raw); err != nil || len(selected) > 1 {
			return result, fmt.Errorf("FRR link-local interface selected metadata is malformed")
		}
		if string(raw) != `""` && string(raw) != "[]" && json.Unmarshal(raw, &options) != nil {
			return result, fmt.Errorf("FRR link-local interface capability is malformed")
		}
		result.InterfaceOptions = map[string]bool{}
		for option, rawOption := range options {
			var entry struct {
				Value *string `json:"value"`
			}
			if json.Unmarshal(rawOption, &entry) != nil || entry.Value == nil {
				return result, fmt.Errorf("FRR link-local interface option metadata is malformed")
			}
			result.InterfaceOptions[option] = true
		}
	}
	family, hasFamily := template.Neighbor["family"]
	multi, hasMulti := template.Neighbor["multiprotocol"]
	if result.Grouped {
		var options map[string]struct {
			Value    string `json:"value"`
			Selected *int   `json:"selected"`
		}
		if !hasFamily || hasMulti || json.Unmarshal(family, &options) != nil || len(options) != 2 {
			return result, fmt.Errorf("FRR compatibility: installed package contradicts the neighbor address-family capability")
		}
		for _, key := range []string{"ipv4", "ipv6"} {
			v, ok := options[key]
			if !ok || v.Value == "" || (v.Selected == nil || (*v.Selected != 0 && *v.Selected != 1)) {
				return result, fmt.Errorf("FRR compatibility: malformed neighbor family options")
			}
		}
	} else {
		var value string
		if hasFamily || !hasMulti || json.Unmarshal(multi, &value) != nil || (value != "0" && value != "1") {
			return result, fmt.Errorf("FRR compatibility: legacy package contradicts the neighbor multiprotocol capability")
		}
	}
	return result, nil
}

func getFRRSingleton[K any](ctx context.Context, c *opnsense.Client, opts opnsense.ReqOpts) (*K, error) {
	var envelope map[string]json.RawMessage
	if err := frrRead(ctx, c, opts.GetEndpoint, &envelope); err != nil {
		return nil, err
	}
	raw, exists := envelope[opts.Monad]
	if !exists || len(raw) < 3 || raw[0] != '{' {
		return nil, fmt.Errorf("FRR singleton read returned missing or empty settings; state was retained")
	}
	var result K
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("FRR singleton read returned malformed settings")
	}
	return &result, nil
}
