// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	errUnknownIntent = errors.New("trust operation inputs must be known before mutation")
	errTrustMissing  = errors.New("trust object is affirmatively absent")
)

func nullableTrustString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// trustRecord accepts only the released monad contract. An empty/default editor
// is not evidence of deletion; malformed responses must retain managed state.
// Trust object identities include both UUIDs and legacy hexadecimal refids.
// Keep them single unencoded path segments; never reinterpret caller input as a URL.
func trustSafeID(id string) bool {
	if id == "" {
		return false
	}
	for _, ch := range id {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_' {
			continue
		}
		return false
	}
	return true
}

func trustRecord(ctx context.Context, c *opnsense.Client, kind, id string) (map[string]string, error) {
	if !trustSafeID(id) {
		return nil, errors.New("invalid Trust object identity")
	}
	if err := c.AcquireRead(ctx); err != nil {
		return nil, errors.New("trust read admission failed")
	}
	defer c.ReleaseRead()
	raw, err := settingsRequest(ctx, c, http.MethodGet, "/api/trust/"+kind+"/get/"+id, nil)
	if err != nil {
		return nil, errors.New("trust read failed; managed state retained")
	}
	envelope, parseErr := settingsObject(raw, []string{kind})
	if parseErr != nil {
		return nil, errors.New("invalid Trust read envelope")
	}
	monad, ok := envelope[kind]
	if !ok {
		return nil, errors.New("missing Trust read monad")
	}
	if string(monad) == "[]" {
		return nil, errTrustMissing
	}
	values, parseErr := settingsObject(monad, nil)
	if parseErr != nil || len(values) == 0 {
		return nil, errors.New("invalid or empty Trust editor response")
	}
	// These strings are always present in a released object editor. Missing
	// fields are not an affirmative empty value and must not erase managed state.
	for _, key := range []string{"descr", "crt_payload", "prv_payload"} {
		var value string
		rawValue, present := values[key]
		if !present || string(rawValue) == "null" || json.Unmarshal(rawValue, &value) != nil {
			return nil, errors.New("incomplete Trust object editor")
		}
	}
	result := make(map[string]string)
	for key, value := range values {
		if key == "in_use" || key == "is_user" {
			v := strings.Trim(string(value), `"`)
			if v != "0" && v != "1" {
				return nil, errors.New("invalid Trust operational boolean")
			}
			result[key] = v
			continue
		}
		if key == "valid_from" || key == "valid_to" {
			var timestamp string
			if string(value) == "null" {
				return nil, errors.New("invalid Trust validity timestamp")
			}
			if json.Unmarshal(value, &timestamp) != nil {
				var number json.Number
				if json.Unmarshal(value, &number) != nil {
					return nil, errors.New("invalid Trust validity timestamp")
				}
				timestamp = number.String()
			}
			if timestamp != "" {
				n, err := strconv.ParseInt(timestamp, 10, 64)
				if err != nil || n < 0 {
					return nil, errors.New("invalid Trust validity timestamp")
				}
			}
			result[key] = timestamp
			continue
		}
		if string(value) == "null" {
			switch key {
			case "refid", "descr", "commonname", "crt_payload", "prv_payload", "csr_payload", "name", "rfc3280_purpose", "city", "state", "organization", "organizationalunit", "email", "ocsp_uri", "altnames_dns", "altnames_ip", "altnames_uri", "altnames_email", "lifetime", "serial", "refcount":
				return nil, errors.New("invalid null Trust string field")
			}
		}
		var s string
		if json.Unmarshal(value, &s) == nil {
			if key == "lifetime" || key == "serial" || key == "refcount" {
				if s != "" {
					n, err := strconv.ParseInt(s, 10, 64)
					if err != nil || n < 0 {
						return nil, errors.New("invalid Trust numeric metadata")
					}
				}
			}
			result[key] = s
			continue
		}
		if key == "country" || key == "caref" || key == "action" || key == "key_type" || key == "digest" || key == "cert_type" || key == "private_key_location" {
			if string(value) == "[]" || string(value) == "null" {
				continue
			}
			var entries map[string]struct {
				Selected json.Number `json:"selected"`
			}
			if json.Unmarshal(value, &entries) != nil {
				return nil, errors.New("invalid Trust selection field")
			}
			selected := 0
			for option, entry := range entries {
				if entry.Selected.String() != "0" && entry.Selected.String() != "1" {
					return nil, errors.New("invalid Trust selection marker")
				}
				if entry.Selected.String() == "1" {
					result[key] = option
					selected++
				}
			}
			if selected > 1 {
				return nil, errors.New("ambiguous Trust selection field")
			}
			continue
		}
		// Unknown appliance fields are not exposed. Known public/private strings
		// must never silently degrade to an empty successful observation.
		switch key {
		case "refid", "descr", "commonname", "crt_payload", "prv_payload", "csr_payload", "name", "rfc3280_purpose", "city", "state", "organization", "organizationalunit", "email", "ocsp_uri", "altnames_dns", "altnames_ip", "altnames_uri", "altnames_email", "lifetime", "serial", "refcount":
			return nil, errors.New("invalid Trust string field")
		}
	}
	if result["refid"] == "" {
		return nil, errors.New("trust editor lacks confirmed object identity")
	}
	return result, nil
}

func trustWrite(ctx context.Context, c *opnsense.Client, kind, operation, id string, payload map[string]string) (string, string, error) {
	if id != "" && !trustSafeID(id) {
		return "", "", errors.New("invalid Trust object identity")
	}
	endpoint := "/api/trust/" + kind + "/" + operation
	if id != "" {
		endpoint += "/" + id
	}
	var body any
	if payload != nil {
		body = map[string]any{kind: payload}
	}
	raw, err := settingsRequest(ctx, c, http.MethodPost, endpoint, body)
	if err != nil {
		return "", "", errors.New("trust mutation response unavailable; do not automatically replay the operation")
	}
	if _, err := settingsObject(raw, []string{"result", "uuid", "private_key"}); err != nil {
		return "", "", errors.New("invalid Trust mutation envelope; do not automatically replay")
	}
	var response struct {
		Result     string `json:"result"`
		UUID       string `json:"uuid"`
		PrivateKey string `json:"private_key"`
	}
	if json.Unmarshal(raw, &response) != nil {
		return "", "", errors.New("invalid Trust mutation response; do not automatically replay")
	}
	expected := "saved"
	if operation == "del" {
		expected = "deleted"
	}
	if response.Result != expected {
		return "", "", errors.New("trust mutation was not confirmed; inspect appliance validation without exposing private payloads")
	}
	if operation == "add" && !trustSafeID(response.UUID) {
		return "", "", errors.New("trust add did not return a confirmed UUID; do not automatically replay")
	}
	return response.UUID, response.PrivateKey, nil
}

func trustOption(options, value string) bool {
	for _, option := range strings.Split(strings.Trim(options, "|"), "|") {
		if option == value {
			return true
		}
	}
	return false
}

func trustAction(p map[string]string, kind string, create, explicit bool) error {
	action := p["action"]
	if _, provided := p["action"]; create && provided && strings.TrimSpace(action) == "" {
		return errors.New("trust action must be nonempty when configured")
	}
	if action == "" {
		if kind == "ca" {
			action = "internal"
		} else {
			action = "import"
		}
	}
	if !create && !explicit {
		if kind == "ca" {
			action = "existing"
		} else {
			action = "manual"
		}
	}
	p["action"] = action
	if strings.TrimSpace(p["descr"]) == "" {
		return errors.New("trust description is required")
	}
	allowed := "|internal|existing|ocsp|"
	if kind == "cert" {
		allowed = "|internal|external|import|sign_csr|import_csr|reissue|manual|"
	}
	if !trustOption(allowed, action) {
		return errors.New("unsupported Trust action")
	}
	for field, allowed := range map[string]string{"key_type": "|512|1024|2048|3072|4096|7680|8192|prime256v1|secp384r1|secp521r1|", "digest": "|sha1|sha224|sha256|sha384|sha512|", "cert_type": "|usr_cert|server_cert|combined_server_client|v3_ca|", "private_key_location": "|firewall|local|"} {
		if value, ok := p[field]; ok && !trustOption(allowed, value) {
			return errors.New("invalid Trust generation option: " + field)
		}
	}
	if value, ok := p["lifetime"]; ok {
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil || n < 1 {
			return errors.New("trust lifetime must be positive days")
		}
	}
	required := []string{}
	switch action {
	case "existing":
		if p["prv_payload"] == "" {
			return errors.New("certificate-only CA import and keyless CA metadata updates are not supported by this provider; the original OPNsense 26.7.8 keyless validation was corrected in 26.7.8_2, but provider support remains deferred; a matching private key is required, never regenerate as a workaround")
		}
		required = []string{"crt_payload", "prv_payload"}
	case "internal": // Released subject fields are optional, including common name.
	case "ocsp":
		required = []string{"caref"}
	case "external": // The model supplies country and cryptographic defaults.

	case "import", "import_csr":
		required = []string{"crt_payload"}
	case "sign_csr":
		required = []string{"caref", "csr_payload"}
	case "reissue":
		required = []string{"caref", "prv_payload"}
	}
	for _, field := range required {
		if strings.TrimSpace(p[field]) == "" {
			return errors.New("trust action requires " + field)
		}
	}
	if !create && action == "internal" && p["private_key_location"] == "local" {
		return errors.New("local-only key generation is supported by add only; set cannot return its key")
	}
	trimTrustPayload(p, action)
	if err := validateTrustPEM(p); err != nil {
		return err
	}
	return nil
}

// Restrict operations to released action-specific fields. In particular, a
// metadata edit must not resend generation material or a retained local key.
func trimTrustPayload(p map[string]string, action string) {
	allowed := "|action|descr|caref|"
	switch action {
	case "internal", "ocsp", "external", "sign_csr", "reissue":
		allowed += "commonname|country|state|city|organization|organizationalunit|email|key_type|digest|lifetime|cert_type|ocsp_uri|altnames_dns|altnames_ip|altnames_uri|altnames_email|"
		switch action {
		case "internal":
			allowed += "private_key_location|"
		case "sign_csr":
			allowed += "csr_payload|"
		case "reissue":
			allowed += "prv_payload|"
		}
	default:
		allowed += "crt_payload|prv_payload|"
	}
	for key := range p {
		if !strings.Contains(allowed, "|"+key+"|") {
			delete(p, key)
		}
	}
}
