// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const trustKeylessCAMinimumPackage = "26.7.8_2"

var (
	errTrustKeylessUnpatched     = errors.New("certificate-only CA import and keyless CA metadata updates are not supported by this provider; the original OPNsense 26.7.8 keyless validation was corrected in 26.7.8_2, but provider support remains deferred; a matching private key is required, never regenerate as a workaround")
	errTrustKeylessDiscriminator = errors.New("certificate-only CA import and keyless CA metadata updates require the installed opnsense package revision from GET /api/core/firmware/info; product_version alone cannot distinguish original 26.7.8 from patched 26.7.8_2; a matching private key is required until that discriminator is available, never regenerate as a workaround")
	stableOPNsensePackage        = regexp.MustCompile(`^[0-9]+\.[0-9]+(?:\.[0-9]+)?(?:_[0-9]+)?$`)
)

type trustPackageVersion struct {
	parts [3]uint64
	rev   uint64
}

func trustCAKeylessRequest(p map[string]string, kind string, create, explicit bool) bool {
	if kind != "ca" || strings.TrimSpace(p["prv_payload"]) != "" {
		return false
	}
	action := p["action"]
	if create && action == "" {
		action = "internal"
	}
	if !create && !explicit {
		action = "existing"
	}
	return action == "existing"
}

func trustPrepareCAAction(ctx context.Context, client *opnsense.Client, p map[string]string, create, explicit bool) error {
	keyless := false
	if trustCAKeylessRequest(p, "ca", create, explicit) {
		supported, err := trustKeylessCASupported(ctx, client)
		if err != nil {
			return err
		}
		if !supported {
			return errTrustKeylessUnpatched
		}
		keyless = true
	}
	return trustAction(p, "ca", create, explicit, keyless)
}

func trustKeylessCASupported(ctx context.Context, client *opnsense.Client) (bool, error) {
	version, err := trustInstalledOPNsensePackage(ctx, client)
	if err != nil {
		return false, err
	}
	ok, err := trustPackageVersionAtLeast(version, trustKeylessCAMinimumPackage)
	if err != nil {
		return false, errTrustKeylessDiscriminator
	}
	return ok, nil
}

func trustInstalledOPNsensePackage(ctx context.Context, client *opnsense.Client) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := client.AcquireRead(ctx); err != nil {
		return "", errTrustKeylessDiscriminator
	}
	defer client.ReleaseRead()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.BaseURL()+"/api/core/firmware/info", nil)
	if err != nil {
		return "", errTrustKeylessDiscriminator
	}
	resp, err := client.HTTPClient().Do(req)
	if err != nil {
		return "", errTrustKeylessDiscriminator
	}
	defer func() { _ = resp.Body.Close() }()
	if err := opnsense.CheckHTTPError(resp.StatusCode, "/api/core/firmware/info"); err != nil {
		return "", errTrustKeylessDiscriminator
	}
	decoder := json.NewDecoder(resp.Body)
	var info struct {
		Packages []struct {
			Name      string `json:"name"`
			Version   string `json:"version"`
			Installed string `json:"installed"`
		} `json:"package"`
	}
	if err := decoder.Decode(&info); err != nil {
		return "", errTrustKeylessDiscriminator
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return "", errTrustKeylessDiscriminator
	}
	found := ""
	for _, pkg := range info.Packages {
		if pkg.Name != "opnsense" {
			continue
		}
		if found != "" || pkg.Installed != "1" || !stableOPNsensePackage.MatchString(pkg.Version) {
			return "", errTrustKeylessDiscriminator
		}
		found = pkg.Version
	}
	if found == "" {
		return "", errTrustKeylessDiscriminator
	}
	return found, nil
}

func parseTrustPackageVersion(value string) (trustPackageVersion, error) {
	var out trustPackageVersion
	if !stableOPNsensePackage.MatchString(value) {
		return out, fmt.Errorf("malformed or non-stable OPNsense package version %q", value)
	}
	base, rev, hasRev := strings.Cut(value, "_")
	if hasRev {
		number, err := strconv.ParseUint(rev, 10, 63)
		if err != nil {
			return out, fmt.Errorf("malformed OPNsense package revision %q", value)
		}
		out.rev = number
	}
	for index, part := range strings.Split(base, ".") {
		number, err := strconv.ParseUint(part, 10, 63)
		if err != nil {
			return out, fmt.Errorf("malformed OPNsense package version %q", value)
		}
		out.parts[index] = number
	}
	return out, nil
}

func trustPackageVersionAtLeast(actual, minimum string) (bool, error) {
	parsedActual, err := parseTrustPackageVersion(actual)
	if err != nil {
		return false, err
	}
	parsedMinimum, err := parseTrustPackageVersion(minimum)
	if err != nil {
		return false, err
	}
	for index := range parsedActual.parts {
		if parsedActual.parts[index] != parsedMinimum.parts[index] {
			return parsedActual.parts[index] > parsedMinimum.parts[index], nil
		}
	}
	return parsedActual.rev >= parsedMinimum.rev, nil
}
