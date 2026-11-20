// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/tfconv"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &generalResource{}
	_ resource.ResourceWithImportState = &generalResource{}
)

const generalID = "unbound-general"

const unboundPendingActivationKey = "unbound_pending_activation_v1"

var (
	unboundNumericString = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)
	unboundCacheSize     = regexp.MustCompile(`(?i)^[0-9]+[kmg]?$`)
)

var generalReqOpts = opnsense.ReqOpts{
	GetEndpoint:               "/api/unbound/settings/get",
	UpdateEndpoint:            "/api/unbound/settings/set",
	ReconfigureEndpoint:       "/api/unbound/service/reconfigure_general",
	StrictReconfigureResponse: true,
	Monad:                     "unbound",
}

var legacyGeneralReqOpts = opnsense.ReqOpts{
	GetEndpoint:         "/api/unbound/settings/get",
	UpdateEndpoint:      "/api/unbound/settings/set",
	ReconfigureEndpoint: "/api/unbound/service/reconfigure",
	Monad:               "unbound",
}

type generalResource struct{ client *opnsense.Client }

func newGeneralResource() resource.Resource { return &generalResource{} }

func (r *generalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_general"
}

func (r *generalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	r.client = client
}

type unboundSettingsEnvelope struct {
	General    json.RawMessage `json:"general"`
	Advanced   json.RawMessage `json:"advanced"`
	ACLs       json.RawMessage `json:"acls"`
	DNSBL      json.RawMessage `json:"dnsbl"`
	Forwarding json.RawMessage `json:"forwarding"`
}

func getGeneralSettings(ctx context.Context, client *opnsense.Client) (*generalAPIResponse, bool, error) {
	settings, err := opnsense.GetSingleton[unboundSettingsEnvelope](ctx, client, generalReqOpts)
	if err != nil {
		return nil, false, err
	}
	result, err := flattenGeneralSettings(settings)
	if err == nil {
		return result, true, nil
	}
	supported, _, compatibilityErr := unboundParitySupport(ctx, client)
	if compatibilityErr != nil {
		return nil, false, fmt.Errorf("%w (compatibility check also failed: %v)", err, compatibilityErr)
	}
	if supported {
		return nil, true, err
	}
	legacy, legacyErr := flattenLegacyGeneralSettings(settings.General)
	if legacyErr != nil {
		return nil, false, legacyErr
	}
	return legacy, false, nil
}

func flattenLegacyGeneralSettings(raw json.RawMessage) (*generalAPIResponse, error) {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return nil, errors.New("malformed legacy Unbound settings response: missing or empty general branch")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("malformed legacy Unbound settings response: decode general branch: %w", err)
	}
	if err := validateUnboundBranchFields("general", fields); err != nil {
		return nil, err
	}
	body, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	var result generalAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("malformed legacy Unbound settings response: %w", err)
	}
	return &result, nil
}

func flattenGeneralSettings(settings *unboundSettingsEnvelope) (*generalAPIResponse, error) {
	branches := []struct {
		name string
		raw  json.RawMessage
	}{
		{"general", settings.General},
		{"advanced", settings.Advanced},
		{"acls", settings.ACLs},
		{"forwarding", settings.Forwarding},
	}
	flat := make(map[string]json.RawMessage)
	for _, branch := range branches {
		if len(branch.raw) == 0 || string(branch.raw) == "null" || string(branch.raw) == "{}" {
			return nil, fmt.Errorf("malformed Unbound settings response: missing or empty %s branch", branch.name)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(branch.raw, &fields); err != nil {
			return nil, fmt.Errorf("malformed Unbound settings response: decode %s branch: %w", branch.name, err)
		}
		if branch.name == "advanced" {
			for _, key := range []string{"logverbosity", "valloglevel"} {
				if raw, ok := fields[key]; ok {
					normalized, err := normalizeIndexedChoice(raw)
					if err != nil {
						return nil, fmt.Errorf("malformed Unbound settings response: advanced.%s: %w", key, err)
					}
					fields[key] = normalized
				}
			}
		}
		if err := validateUnboundBranchFields(branch.name, fields); err != nil {
			return nil, err
		}
		for key, value := range fields {
			if branch.name == "forwarding" && key == "enabled" {
				key = "forwarding_enabled"
			}
			flat[key] = value
		}
	}
	if _, ok := flat["port"]; !ok {
		return nil, fmt.Errorf("malformed Unbound settings response: general.port is absent")
	}
	if _, ok := flat["default_action"]; !ok {
		return nil, fmt.Errorf("malformed Unbound settings response: acls.default_action is absent")
	}
	body, err := json.Marshal(flat)
	if err != nil {
		return nil, err
	}
	var result generalAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("malformed Unbound settings response: %w", err)
	}
	return &result, nil
}

func normalizeIndexedChoice(raw json.RawMessage) (json.RawMessage, error) {
	var scalar string
	if err := json.Unmarshal(raw, &scalar); err == nil {
		return raw, nil
	}
	var options []struct {
		Selected json.Number `json:"selected"`
	}
	if err := json.Unmarshal(raw, &options); err != nil {
		return nil, fmt.Errorf("must be a string or selected-option array: %w", err)
	}
	selected := -1
	for index, option := range options {
		switch option.Selected.String() {
		case "0":
		case "1":
			if selected >= 0 {
				return nil, errors.New("selected-option array contains multiple selections")
			}
			selected = index
		default:
			return nil, fmt.Errorf("selected-option array has invalid selection %q", option.Selected.String())
		}
	}
	if selected < 0 {
		return nil, errors.New("selected-option array contains no selection")
	}
	return json.Marshal(strconv.Itoa(selected))
}

func validateUnboundBranchFields(branch string, fields map[string]json.RawMessage) error {
	required := map[string][]string{
		"general":  {"enabled", "port", "stats", "active_interface", "dnssec", "dns64", "dns64prefix", "noarecords", "regdhcp", "regdhcpdomain", "regdhcpstatic", "noreglladdr6", "noregrecords", "txtsupport", "cacheflush", "local_zone_type", "outgoing_interface", "enable_wpad"},
		"advanced": {"hideidentity", "hideversion", "prefetchkey", "dnssecstripped", "aggressivensec", "qnameminstrict", "outgoingnumtcp", "incomingnumtcp", "numqueriesperthread", "outgoingrange", "jostletimeout", "discardtimeout", "privatedomain", "privateaddress", "insecuredomain", "prefetch", "unwantedreplythreshold", "msgcachesize", "rrsetcachesize", "cachemaxttl", "cachemaxnegativettl", "cacheminttl", "infrahostttl", "infrakeepprobing", "infracachenumhosts", "extendedstatistics", "logqueries", "logreplies", "logtagqueryreply", "loglocalactions", "logservfail", "logverbosity", "valloglevel", "serveexpired", "serveexpiredreplyttl", "serveexpiredttl", "serveexpiredttlreset", "serveexpiredclienttimeout"},
		"acls":     {"default_action"}, "forwarding": {"enabled"},
	}
	for _, key := range required[branch] {
		if _, ok := fields[key]; !ok {
			return fmt.Errorf("malformed Unbound settings response: %s.%s is absent", branch, key)
		}
	}
	boolFields := map[string]bool{
		"enabled": true, "stats": true, "dnssec": true, "dns64": true, "noarecords": true,
		"regdhcp": true, "regdhcpstatic": true, "noreglladdr6": true, "noregrecords": true,
		"txtsupport": true, "cacheflush": true, "enable_wpad": true, "safesearch": true,
		"hideidentity": true, "hideversion": true, "prefetchkey": true, "dnssecstripped": true,
		"aggressivensec": true, "qnameminstrict": true, "prefetch": true, "infrakeepprobing": true,
		"extendedstatistics": true, "logqueries": true, "logreplies": true, "logtagqueryreply": true,
		"loglocalactions": true, "logservfail": true, "serveexpired": true, "serveexpiredttlreset": true,
	}
	intFields := map[string]bool{
		"outgoingnumtcp": true, "incomingnumtcp": true, "numqueriesperthread": true, "outgoingrange": true,
		"jostletimeout": true, "discardtimeout": true, "unwantedreplythreshold": true, "cachemaxttl": true,
		"cachemaxnegativettl": true, "cacheminttl": true, "infrahostttl": true, "infracachenumhosts": true,
		"logverbosity": true, "valloglevel": true,
	}
	numericStringFields := map[string]bool{
		"serveexpiredreplyttl": true, "serveexpiredttl": true, "serveexpiredclienttimeout": true,
	}
	for key, raw := range fields {
		if !boolFields[key] && !intFields[key] && !numericStringFields[key] {
			continue
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("malformed Unbound settings response: %s.%s must be a string: %w", branch, key, err)
		}
		if boolFields[key] && value != "0" && value != "1" {
			return fmt.Errorf("malformed Unbound settings response: %s.%s has invalid boolean %q", branch, key, value)
		}
		if (intFields[key] || numericStringFields[key]) && string(raw) == "null" {
			return fmt.Errorf("malformed Unbound settings response: %s.%s must be a string", branch, key)
		}
		if intFields[key] && value != "" {
			number, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("malformed Unbound settings response: %s.%s has invalid integer %q", branch, key, value)
			}
			maximum := int64(^uint64(0) >> 1)
			switch key {
			case "logverbosity":
				maximum = 5
			case "valloglevel":
				maximum = 2
			}
			if number < 0 || number > maximum {
				return fmt.Errorf("malformed Unbound settings response: %s.%s integer %d is outside 0..%d", branch, key, number, maximum)
			}
		}
		if numericStringFields[key] && value != "" {
			if !unboundNumericString.MatchString(value) {
				return fmt.Errorf("malformed Unbound settings response: %s.%s has invalid numeric value %q", branch, key, value)
			}
			number, err := strconv.ParseFloat(value, 64)
			if err != nil || math.IsInf(number, 0) || math.IsNaN(number) || number < -99999999999999 || number > 99999999999999 {
				return fmt.Errorf("malformed Unbound settings response: %s.%s numeric value %q is outside -99999999999999..99999999999999", branch, key, value)
			}
		}
	}
	if branch == "advanced" {
		for _, key := range []string{"msgcachesize", "rrsetcachesize"} {
			if string(fields[key]) == "null" {
				return fmt.Errorf("malformed Unbound settings response: advanced.%s must be a string", key)
			}
			var value string
			if err := json.Unmarshal(fields[key], &value); err != nil {
				return fmt.Errorf("malformed Unbound settings response: advanced.%s must be a string: %w", key, err)
			}
			if value != "" && !unboundCacheSize.MatchString(value) {
				return fmt.Errorf("malformed Unbound settings response: advanced.%s has invalid cache size %q", key, value)
			}
		}
	}
	if branch == "general" {
		var port string
		if err := json.Unmarshal(fields["port"], &port); err != nil {
			return fmt.Errorf("malformed Unbound settings response: general.port must be a string: %w", err)
		}
		number, err := strconv.ParseInt(port, 10, 64)
		if err != nil || number < 1 || number > 65535 {
			return fmt.Errorf("malformed Unbound settings response: general.port %q is outside 1..65535", port)
		}
	}
	selectedFields := map[string]bool{}
	switch branch {
	case "general":
		selectedFields = map[string]bool{
			"active_interface": false, "outgoing_interface": false, "local_zone_type": true,
		}
	case "advanced":
		selectedFields = map[string]bool{
			"privatedomain": false, "privateaddress": false, "insecuredomain": false,
		}
	case "acls":
		selectedFields = map[string]bool{"default_action": true}
	}
	for key, requireOne := range selectedFields {
		raw, ok := fields[key]
		if !ok {
			continue
		}
		if err := validateSelectedMap(raw, requireOne); err != nil {
			return fmt.Errorf("malformed Unbound settings response: %s.%s: %w", branch, key, err)
		}
		if requireOne {
			var selected opnsense.SelectedMap
			if err := json.Unmarshal(raw, &selected); err != nil {
				return fmt.Errorf("malformed Unbound settings response: %s.%s: %w", branch, key, err)
			}
			allowed := map[string]map[string]bool{
				"general.local_zone_type": {"transparent": true, "always_nxdomain": true, "always_refuse": true, "always_transparent": true, "deny": true, "refuse": true, "nodefault": true, "typetransparent": true, "inform": true, "inform_deny": true, "static": true},
				"acls.default_action":     {"allow": true, "deny": true, "refuse": true},
			}
			if choices := allowed[branch+"."+key]; choices != nil && !choices[string(selected)] {
				return fmt.Errorf("malformed Unbound settings response: %s.%s has unsupported choice %q", branch, key, selected)
			}
		}
	}
	return nil
}

func validateSelectedMap(raw json.RawMessage, requireOne bool) error {
	var scalar string
	if err := json.Unmarshal(raw, &scalar); err == nil {
		if requireOne && scalar == "" {
			return errors.New("selected field is empty")
		}
		return nil
	}
	// OPNsense uses []/null for an unset selector on some releases. Preserve
	// that compatibility while still rejecting malformed populated maps.
	if string(raw) == "[]" || string(raw) == "null" {
		return nil
	}
	var options map[string]struct {
		Selected json.Number `json:"selected"`
	}
	if err := json.Unmarshal(raw, &options); err != nil {
		return fmt.Errorf("must be a string or selected-option map: %w", err)
	}
	selected := 0
	for _, option := range options {
		switch option.Selected.String() {
		case "0":
		case "1":
			selected++
		default:
			return fmt.Errorf("selected-option map has invalid selection %q", option.Selected.String())
		}
	}
	if requireOne && selected != 1 {
		return fmt.Errorf("selected-option map contains %d selections; expected exactly one", selected)
	}
	return nil
}

func putString(dst map[string]any, key string, value types.String) {
	if !value.IsNull() && !value.IsUnknown() {
		dst[key] = value.ValueString()
	}
}

func putBool(dst map[string]any, key string, value types.Bool) {
	if !value.IsNull() && !value.IsUnknown() {
		dst[key] = opnsense.BoolToString(value.ValueBool())
	}
}

func putInt(dst map[string]any, key string, value types.Int64) {
	if !value.IsNull() && !value.IsUnknown() {
		dst[key] = opnsense.Int64ToString(value.ValueInt64())
	}
}

func putSet(ctx context.Context, dst map[string]any, key string, value types.Set) {
	if !value.IsNull() && !value.IsUnknown() {
		dst[key] = tfconv.SetToCSV(ctx, value)
	}
}

func generalSettingsPayload(ctx context.Context, m GeneralResourceModel) map[string]any {
	general := make(map[string]any)
	putBool(general, "enabled", m.Enabled)
	putString(general, "port", m.Port)
	putBool(general, "stats", m.Stats)
	putSet(ctx, general, "active_interface", m.ActiveInterface)
	putBool(general, "dnssec", m.DNSSEC)
	putBool(general, "dns64", m.DNS64)
	putString(general, "dns64prefix", m.DNS64Prefix)
	putBool(general, "noarecords", m.NoAAAARecords)
	putBool(general, "regdhcp", m.RegisterDHCP)
	putString(general, "regdhcpdomain", m.RegisterDHCPDomain)
	putBool(general, "regdhcpstatic", m.RegisterDHCPStatic)
	putBool(general, "noreglladdr6", m.NoRegisterLinkLocal6)
	putBool(general, "noregrecords", m.NoRegisterRecords)
	putBool(general, "txtsupport", m.TXTSupport)
	putBool(general, "cacheflush", m.CacheFlush)
	putString(general, "local_zone_type", m.LocalZoneType)
	putSet(ctx, general, "outgoing_interface", m.OutgoingInterface)
	putBool(general, "enable_wpad", m.EnableWPAD)
	advanced := make(map[string]any)
	putBool(advanced, "hideidentity", m.HideIdentity)
	putBool(advanced, "hideversion", m.HideVersion)
	putBool(advanced, "prefetchkey", m.EnablePrefetchKey)
	putBool(advanced, "dnssecstripped", m.DNSSECStripped)
	putBool(advanced, "aggressivensec", m.AggressiveNSEC)
	putBool(advanced, "qnameminstrict", m.QnameMinStrict)
	putInt(advanced, "outgoingnumtcp", m.OutgoingNumTCP)
	putInt(advanced, "incomingnumtcp", m.IncomingNumTCP)
	putInt(advanced, "numqueriesperthread", m.NumQueriesPerThread)
	putInt(advanced, "outgoingrange", m.OutgoingRange)
	putInt(advanced, "jostletimeout", m.JostleTimeout)
	putInt(advanced, "discardtimeout", m.DiscardTimeout)
	putSet(ctx, advanced, "privatedomain", m.PrivateDomains)
	putSet(ctx, advanced, "privateaddress", m.PrivateAddresses)
	putSet(ctx, advanced, "insecuredomain", m.InsecureDomains)
	putBool(advanced, "prefetch", m.EnablePrefetch)
	putInt(advanced, "unwantedreplythreshold", m.UnwantedReplyThreshold)
	putString(advanced, "msgcachesize", m.MsgCacheSize)
	putString(advanced, "rrsetcachesize", m.RRSetCacheSize)
	putInt(advanced, "cachemaxttl", m.MaxTTL)
	putInt(advanced, "cachemaxnegativettl", m.MaxNegativeTTL)
	putInt(advanced, "cacheminttl", m.MinTTL)
	putInt(advanced, "infrahostttl", m.HostTTL)
	putBool(advanced, "infrakeepprobing", m.KeepProbingHosts)
	putInt(advanced, "infracachenumhosts", m.NumHosts)
	putBool(advanced, "extendedstatistics", m.ExtendedStatistics)
	putBool(advanced, "logqueries", m.LogQueries)
	putBool(advanced, "logreplies", m.LogReplies)
	putBool(advanced, "logtagqueryreply", m.TagQueryReply)
	putBool(advanced, "loglocalactions", m.LogLocalActions)
	putBool(advanced, "logservfail", m.LogServfail)
	putInt(advanced, "logverbosity", m.VerbosityLevel)
	putInt(advanced, "valloglevel", m.ValidationLevel)
	putBool(advanced, "serveexpired", m.ServeExpiredEnabled)
	putString(advanced, "serveexpiredreplyttl", m.ServeExpiredRecordReplyTTL)
	putString(advanced, "serveexpiredttl", m.ServeExpiredTTL)
	putBool(advanced, "serveexpiredttlreset", m.ServeExpiredResetTTL)
	putString(advanced, "serveexpiredclienttimeout", m.ServeExpiredClientResponseTimeout)

	acls := make(map[string]any)
	putString(acls, "default_action", m.ACLDefaultAction)
	forwarding := make(map[string]any)
	putBool(forwarding, "enabled", m.ForwardingEnabled)
	return map[string]any{"general": general, "advanced": advanced, "acls": acls, "forwarding": forwarding}
}

type unboundPrivateState interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

func getUnboundPending(ctx context.Context, p unboundPrivateState) (bool, diag.Diagnostics) {
	if p == nil {
		return false, nil
	}
	v, d := p.GetKey(ctx, unboundPendingActivationKey)
	if len(v) == 0 || d.HasError() {
		return false, d
	}
	var pending bool
	if err := json.Unmarshal(v, &pending); err != nil {
		d.AddError("Invalid private pending state", err.Error())
	}
	return pending, d
}

func setUnboundPending(ctx context.Context, p unboundPrivateState, pending bool) diag.Diagnostics {
	if p == nil {
		return nil
	}
	if pending {
		return p.SetKey(ctx, unboundPendingActivationKey, []byte("true"))
	}
	return p.SetKey(ctx, unboundPendingActivationKey, nil)
}

func generalPayloadMatchesRemote(ctx context.Context, payload map[string]any, remote GeneralResourceModel) bool {
	actual := generalSettingsPayload(ctx, remote)
	for branchName, desiredBranchValue := range payload {
		desiredBranch, ok := desiredBranchValue.(map[string]any)
		if !ok {
			return false
		}
		actualBranch, ok := actual[branchName].(map[string]any)
		if !ok {
			return false
		}
		for key, desired := range desiredBranch {
			if !reflect.DeepEqual(actualBranch[key], desired) {
				return false
			}
		}
	}
	return true
}

func generalChangedPayload(ctx context.Context, plan, prior GeneralResourceModel) map[string]any {
	desired := generalSettingsPayload(ctx, plan)
	previous := generalSettingsPayload(ctx, prior)
	changed := make(map[string]any)
	for branch, value := range desired {
		if !reflect.DeepEqual(value, previous[branch]) {
			changed[branch] = value
		}
	}
	return changed
}

func minimalGeneralState() GeneralResourceModel {
	return GeneralResourceModel{ID: types.StringValue(generalID), ActiveInterface: types.SetNull(types.StringType), OutgoingInterface: types.SetNull(types.StringType), PrivateDomains: types.SetNull(types.StringType), PrivateAddresses: types.SetNull(types.StringType), InsecureDomains: types.SetNull(types.StringType)}
}

func clearGeneralParityState(state *GeneralResourceModel) {
	state.HideIdentity = types.BoolNull()
	state.HideVersion = types.BoolNull()
	state.EnablePrefetchKey = types.BoolNull()
	state.DNSSECStripped = types.BoolNull()
	state.AggressiveNSEC = types.BoolNull()
	state.QnameMinStrict = types.BoolNull()
	state.OutgoingNumTCP = types.Int64Null()
	state.IncomingNumTCP = types.Int64Null()
	state.NumQueriesPerThread = types.Int64Null()
	state.OutgoingRange = types.Int64Null()
	state.JostleTimeout = types.Int64Null()
	state.DiscardTimeout = types.Int64Null()
	state.PrivateDomains = types.SetNull(types.StringType)
	state.PrivateAddresses = types.SetNull(types.StringType)
	state.InsecureDomains = types.SetNull(types.StringType)
	state.EnablePrefetch = types.BoolNull()
	state.UnwantedReplyThreshold = types.Int64Null()
	state.MsgCacheSize = types.StringNull()
	state.RRSetCacheSize = types.StringNull()
	state.MaxTTL = types.Int64Null()
	state.MaxNegativeTTL = types.Int64Null()
	state.MinTTL = types.Int64Null()
	state.HostTTL = types.Int64Null()
	state.KeepProbingHosts = types.BoolNull()
	state.NumHosts = types.Int64Null()
	state.ExtendedStatistics = types.BoolNull()
	state.LogQueries = types.BoolNull()
	state.LogReplies = types.BoolNull()
	state.TagQueryReply = types.BoolNull()
	state.LogLocalActions = types.BoolNull()
	state.LogServfail = types.BoolNull()
	state.VerbosityLevel = types.Int64Null()
	state.ValidationLevel = types.Int64Null()
	state.ServeExpiredEnabled = types.BoolNull()
	state.ServeExpiredRecordReplyTTL = types.StringNull()
	state.ServeExpiredTTL = types.StringNull()
	state.ServeExpiredResetTTL = types.BoolNull()
	state.ServeExpiredClientResponseTimeout = types.StringNull()
	state.ACLDefaultAction = types.StringNull()
	state.ForwardingEnabled = types.BoolNull()
}

func generalUsesParity(state GeneralResourceModel) bool {
	values := []interface {
		IsNull() bool
		IsUnknown() bool
	}{
		state.HideIdentity, state.HideVersion, state.EnablePrefetchKey, state.DNSSECStripped, state.AggressiveNSEC, state.QnameMinStrict,
		state.OutgoingNumTCP, state.IncomingNumTCP, state.NumQueriesPerThread, state.OutgoingRange, state.JostleTimeout, state.DiscardTimeout,
		state.PrivateDomains, state.PrivateAddresses, state.InsecureDomains, state.EnablePrefetch, state.UnwantedReplyThreshold,
		state.MsgCacheSize, state.RRSetCacheSize, state.MaxTTL, state.MaxNegativeTTL, state.MinTTL, state.HostTTL, state.KeepProbingHosts,
		state.NumHosts, state.ExtendedStatistics, state.LogQueries, state.LogReplies, state.TagQueryReply, state.LogLocalActions,
		state.LogServfail, state.VerbosityLevel, state.ValidationLevel, state.ServeExpiredEnabled, state.ServeExpiredRecordReplyTTL,
		state.ServeExpiredTTL, state.ServeExpiredResetTTL, state.ServeExpiredClientResponseTimeout, state.ACLDefaultAction, state.ForwardingEnabled,
	}
	for _, value := range values {
		if !value.IsNull() && !value.IsUnknown() {
			return true
		}
	}
	return false
}

func authoritativeGeneralState(ctx context.Context, client *opnsense.Client) (*GeneralResourceModel, error) {
	result, parity, err := getGeneralSettings(ctx, client)
	if err != nil {
		return nil, err
	}
	var state GeneralResourceModel
	state.fromAPI(ctx, result, generalID)
	normalizeUnsetGeneralIntegers(&state, result)
	if !parity {
		clearGeneralParityState(&state)
	}
	return &state, nil
}

func (r *generalResource) apply(ctx context.Context, prior *GeneralResourceModel, payload map[string]any, opts opnsense.ReqOpts, private unboundPrivateState, stateSetter func(context.Context, any) diag.Diagnostics, diagnostics *diag.Diagnostics, action string) {
	if len(payload) == 0 {
		result, err := authoritativeGeneralState(ctx, r.client)
		if err != nil {
			diagnostics.AddError("Error reading the Unbound DNS settings during "+action, err.Error())
			return
		}
		*diagnostics = append(*diagnostics, stateSetter(ctx, result)...)
		return
	}
	err := opnsense.UpdateSingleton(ctx, r.client, opts, &payload)
	if err != nil {
		var activationErr *opnsense.MutationReconfigureError
		var validationErr *opnsense.ValidationError
		committed := errors.As(err, &activationErr)
		var confirmed *GeneralResourceModel
		if !errors.As(err, &validationErr) {
			remote, readErr := authoritativeGeneralState(ctx, r.client)
			if readErr != nil {
				diagnostics.AddError("Unable to confirm Unbound settings mutation outcome", readErr.Error())
			} else {
				confirmed = remote
				if !committed {
					committed = generalPayloadMatchesRemote(ctx, payload, *remote)
				}
			}
		}
		if committed {
			*diagnostics = append(*diagnostics, setUnboundPending(ctx, private, true)...)
			switch {
			case confirmed != nil:
				*diagnostics = append(*diagnostics, stateSetter(ctx, confirmed)...)
			case prior != nil:
				*diagnostics = append(*diagnostics, stateSetter(ctx, prior)...)
			default:
				minimal := minimalGeneralState()
				*diagnostics = append(*diagnostics, stateSetter(ctx, &minimal)...)
			}
		}
		diagnostics.AddError("Error "+action+" the Unbound DNS settings", err.Error())
		return
	}
	result, err := authoritativeGeneralState(ctx, r.client)
	if err != nil {
		if prior != nil {
			*diagnostics = append(*diagnostics, stateSetter(ctx, prior)...)
		} else {
			minimal := minimalGeneralState()
			*diagnostics = append(*diagnostics, stateSetter(ctx, &minimal)...)
		}
		diagnostics.AddError("Error reading the Unbound DNS settings after "+action, err.Error())
		return
	}
	*diagnostics = append(*diagnostics, setUnboundPending(ctx, private, false)...)
	*diagnostics = append(*diagnostics, stateSetter(ctx, result)...)
}

func (r *generalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts := generalReqOpts
	payload := generalSettingsPayload(ctx, plan)
	if req.Config.Schema != nil {
		var config GeneralResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if resp.Diagnostics.HasError() {
			return
		}
		supported, version, err := unboundParitySupport(ctx, r.client)
		if err != nil {
			resp.Diagnostics.AddError("Unable to check OPNsense compatibility", err.Error())
			return
		}
		if generalUsesParity(config) && !supported {
			resp.Diagnostics.AddError("Unsupported OPNsense version", fmt.Sprintf("Unbound parity fields require stable OPNsense %s or newer; appliance reports %s", unboundParityMinimumVersion, version))
			return
		}
		if !supported {
			opts = legacyGeneralReqOpts
			payload = map[string]any{"general": payload["general"]}
		}
	}
	r.apply(ctx, nil, payload, opts, resp.Private, resp.State.Set, &resp.Diagnostics, "applying")
}

func (r *generalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GeneralResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pending, d := getUnboundPending(ctx, req.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pending {
		opts, err := unboundCompatibilityOpts(ctx, r.client, generalReqOpts, legacyGeneralReqOpts)
		if err != nil {
			resp.Diagnostics.AddError("Unable to check OPNsense compatibility", err.Error())
			return
		}
		if err := opnsense.RetryReconfigure(ctx, r.client, opts); err != nil {
			resp.Diagnostics.AddError("Error activating pending Unbound settings mutation", err.Error())
			return
		}
		resp.Diagnostics.Append(setUnboundPending(ctx, resp.Private, false)...)
	}
	result, parity, err := getGeneralSettings(ctx, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error reading the Unbound DNS settings", err.Error())
		return
	}
	state.fromAPI(ctx, result, generalID)
	normalizeUnsetGeneralIntegers(&state, result)
	if !parity {
		clearGeneralParityState(&state)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *generalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GeneralResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts := generalReqOpts
	legacy := false
	if req.Config.Schema != nil {
		var config GeneralResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if resp.Diagnostics.HasError() {
			return
		}
		supported, version, err := unboundParitySupport(ctx, r.client)
		if err != nil {
			resp.Diagnostics.AddError("Unable to check OPNsense compatibility", err.Error())
			return
		}
		if generalUsesParity(config) && !supported {
			resp.Diagnostics.AddError("Unsupported OPNsense version", fmt.Sprintf("Unbound parity fields require stable OPNsense %s or newer; appliance reports %s", unboundParityMinimumVersion, version))
			return
		}
		if !supported {
			opts = legacyGeneralReqOpts
			legacy = true
		}
	}
	var prior GeneralResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pending, d := getUnboundPending(ctx, req.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pending {
		if err := opnsense.RetryReconfigure(ctx, r.client, opts); err != nil {
			resp.Diagnostics.AddError("Error activating pending Unbound settings mutation", err.Error())
			return
		}
		resp.Diagnostics.Append(setUnboundPending(ctx, resp.Private, false)...)
		remote, err := authoritativeGeneralState(ctx, r.client)
		if err != nil {
			resp.Diagnostics.AddError("Error reconciling pending Unbound settings update", err.Error())
			return
		}
		if generalPayloadMatchesRemote(ctx, generalSettingsPayload(ctx, plan), *remote) {
			resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
			return
		}
		prior = *remote
	}
	payload := generalChangedPayload(ctx, plan, prior)
	if legacy {
		payload = map[string]any{"general": generalSettingsPayload(ctx, plan)["general"]}
	}
	r.apply(ctx, &prior, payload, opts, resp.Private, resp.State.Set, &resp.Diagnostics, "updating")
}

// Delete is deliberately state-only: singleton settings are never reset.
func (r *generalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	pending, diagnostics := getUnboundPending(ctx, req.Private)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() || !pending {
		return
	}
	opts, err := unboundCompatibilityOpts(ctx, r.client, generalReqOpts, legacyGeneralReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to check OPNsense compatibility", err.Error())
		return
	}
	if err := opnsense.RetryReconfigure(ctx, r.client, opts); err != nil {
		resp.Diagnostics.AddError("Error activating pending Unbound settings mutation before state removal", err.Error())
		return
	}
	resp.Diagnostics.Append(setUnboundPending(ctx, resp.Private, false)...)
}

func (r *generalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != generalID {
		resp.Diagnostics.AddError("Invalid import identifier", fmt.Sprintf("use %q for this singleton, got %q", generalID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(generalID))...)
}
