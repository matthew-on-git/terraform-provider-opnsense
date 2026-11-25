// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/tfconv"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// generalLookupFields is the public allowlist. It is the resource refresh
// contract, not every field the stable settings GET returns. Released core
// 26.7.9 and 26.7.10 keep Unbound.xml e81b0746f78b. getAction returns
// getNodes() for the whole unbound model. Sibling nodes and fields the
// resource never refreshes are not lookup state. The nodes have no password
// or write-only secret. A missing or blank required field fails closed and
// is not replaced with its model default. An optional BooleanField or
// NumericField with no Required and no Default stays null when blank.
var generalLookupFields = []string{
	"id",
	"enabled",
	"port",
	"stats",
	"active_interface",
	"dnssec",
	"dns64",
	"dns64_prefix",
	"no_aaaa_records",
	"register_dhcp",
	"register_dhcp_domain",
	"register_dhcp_static",
	"no_register_link_local6",
	"no_register_records",
	"txt_support",
	"cache_flush",
	"local_zone_type",
	"outgoing_interface",
	"enable_wpad",
	"hide_identity",
	"hide_version",
	"enable_prefetch_key",
	"dnssec_stripped",
	"aggressive_nsec",
	"qname_min_strict",
	"outgoing_num_tcp",
	"incoming_num_tcp",
	"num_queries_per_thread",
	"outgoing_range",
	"jostle_timeout",
	"discard_timeout",
	"private_domains",
	"private_addresses",
	"insecure_domains",
	"enable_prefetch",
	"unwanted_reply_threshold",
	"msg_cache_size",
	"rrset_cache_size",
	"max_ttl",
	"max_negative_ttl",
	"min_ttl",
	"host_ttl",
	"keep_probing_hosts",
	"num_hosts",
	"extended_statistics",
	"log_queries",
	"log_replies",
	"tag_query_reply",
	"log_local_actions",
	"log_servfail",
	"verbosity_level",
	"validation_level",
	"serve_expired_enabled",
	"serve_expired_record_reply_ttl",
	"serve_expired_ttl",
	"serve_expired_reset_ttl",
	"serve_expired_client_response_timeout",
	"acl_default_action",
	"forwarding_enabled",
}

var generalZoneOptions = optionSet(
	"transparent", "always_nxdomain", "always_refuse", "always_transparent",
	"deny", "inform", "inform_deny", "nodefault", "refuse", "static", "typetransparent",
)

var generalACLActions = optionSet("allow", "deny", "refuse")

var generalVerbosityOptions = optionSet("0", "1", "2", "3", "4", "5")

var generalValidationOptions = optionSet("0", "1", "2")

const generalNotFoundSummary = "Unbound general settings not found"

var _ datasource.DataSource = &generalDataSource{}

type generalDataSource struct{ client *opnsense.Client }

func newGeneralDataSource() datasource.DataSource { return &generalDataSource{} }

func (d *generalDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_general"
}

func (d *generalDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads Unbound general settings. The only accepted identity is `unbound-general`. This data source never writes, deletes, or reconfigures. Released core 26.7.9 and 26.7.10 return the whole Unbound model from `GET /api/unbound/settings/get`. The public allowlist is the resource refresh contract: general, advanced, the ACL default action, and forwarding enablement. `safesearch` belongs to the DNSBL lookup. `belownxdomain`, ACL items, and sibling host, alias, forward, and blocklist nodes are not lookup state. The nodes have no password or write-only secret. A missing or blank required field fails closed and is not replaced with its model default. Optional booleans and integers stay null when blank. An empty optional string stays empty. Literal zero is kept when the model minimum is not above zero.",
		Attributes:          generalLookupAttributes(),
	}
}

func (d *generalDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	d.client = client
}

func (d *generalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config GeneralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != generalID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid Unbound general settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", generalID, config.ID.ValueString()),
		)
		return
	}
	settings, err := opnsense.GetSingleton[unboundSettingsEnvelope](ctx, d.client, generalReqOpts)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			resp.Diagnostics.AddError(
				generalNotFoundSummary,
				fmt.Sprintf("The settings reader returned %s. An empty required object is not found and is not replaced with model defaults.", notFound.Error()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading Unbound general settings", err.Error())
		return
	}
	state, err := generalLookupFromEnvelope(settings)
	if err != nil {
		resp.Diagnostics.AddError("Malformed Unbound general settings response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func generalLookupFromEnvelope(settings *unboundSettingsEnvelope) (GeneralResourceModel, error) {
	if settings == nil {
		return GeneralResourceModel{}, errors.New("response is missing")
	}
	general, err := generalBranchFields("general", settings.General)
	if err != nil {
		return GeneralResourceModel{}, err
	}
	advanced, err := generalBranchFields("advanced", settings.Advanced)
	if err != nil {
		return GeneralResourceModel{}, err
	}
	acls, err := generalBranchFields("acls", settings.ACLs)
	if err != nil {
		return GeneralResourceModel{}, err
	}
	forwarding, err := generalBranchFields("forwarding", settings.Forwarding)
	if err != nil {
		return GeneralResourceModel{}, err
	}

	enabled, err := generalRequiredBool(general, "enabled")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	port, err := generalRequiredPort(general)
	if err != nil {
		return GeneralResourceModel{}, err
	}
	zone, err := generalRequiredOption(general, "local_zone_type", generalZoneOptions)
	if err != nil {
		return GeneralResourceModel{}, err
	}
	aggressive, err := generalRequiredBool(advanced, "aggressivensec")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	verbosity, err := generalRequiredOption(advanced, "logverbosity", generalVerbosityOptions)
	if err != nil {
		return GeneralResourceModel{}, err
	}
	validation, err := generalRequiredOption(advanced, "valloglevel", generalValidationOptions)
	if err != nil {
		return GeneralResourceModel{}, err
	}
	privateAddresses, err := generalRequiredNetworks(advanced, "privateaddress")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	action, err := generalRequiredOption(acls, "default_action", generalACLActions)
	if err != nil {
		return GeneralResourceModel{}, err
	}

	stats, err := generalOptionalBool(general, "stats")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	dnssec, err := generalOptionalBool(general, "dnssec")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	dns64, err := generalOptionalBool(general, "dns64")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	prefix, err := generalOptionalPrefix(general, "dns64prefix")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	noAAAA, err := generalOptionalBool(general, "noarecords")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	registerDHCP, err := generalOptionalBool(general, "regdhcp")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	registerDomain, err := generalOptionalText(general, "regdhcpdomain")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	registerStatic, err := generalOptionalBool(general, "regdhcpstatic")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	noLinkLocal, err := generalOptionalBool(general, "noreglladdr6")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	noRecords, err := generalOptionalBool(general, "noregrecords")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	txtSupport, err := generalOptionalBool(general, "txtsupport")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	cacheFlush, err := generalOptionalBool(general, "cacheflush")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	active, err := generalOptionalSet(general, "active_interface")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	outgoing, err := generalOptionalSet(general, "outgoing_interface")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	wpad, err := generalOptionalBool(general, "enable_wpad")
	if err != nil {
		return GeneralResourceModel{}, err
	}

	hideIdentity, err := generalOptionalBool(advanced, "hideidentity")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	hideVersion, err := generalOptionalBool(advanced, "hideversion")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	prefetch, err := generalOptionalBool(advanced, "prefetch")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	prefetchKey, err := generalOptionalBool(advanced, "prefetchkey")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	stripped, err := generalOptionalBool(advanced, "dnssecstripped")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	qname, err := generalOptionalBool(advanced, "qnameminstrict")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	outgoingTCP, err := generalOptionalInt(advanced, "outgoingnumtcp")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	incomingTCP, err := generalOptionalInt(advanced, "incomingnumtcp")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	queries, err := generalOptionalInt(advanced, "numqueriesperthread")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	outgoingRange, err := generalOptionalInt(advanced, "outgoingrange")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	jostle, err := generalOptionalInt(advanced, "jostletimeout")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	discard, err := generalOptionalInt(advanced, "discardtimeout")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	privateDomains, err := generalOptionalSet(advanced, "privatedomain")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	insecureDomains, err := generalOptionalSet(advanced, "insecuredomain")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	msgCache, err := generalOptionalCache(advanced, "msgcachesize")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	rrsetCache, err := generalOptionalCache(advanced, "rrsetcachesize")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	maxTTL, err := generalOptionalInt(advanced, "cachemaxttl")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	maxNegative, err := generalOptionalInt(advanced, "cachemaxnegativettl")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	minTTL, err := generalOptionalInt(advanced, "cacheminttl")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	hostTTL, err := generalOptionalInt(advanced, "infrahostttl")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	keepProbing, err := generalOptionalBool(advanced, "infrakeepprobing")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	numHosts, err := generalOptionalInt(advanced, "infracachenumhosts")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	threshold, err := generalOptionalInt(advanced, "unwantedreplythreshold")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	extended, err := generalOptionalBool(advanced, "extendedstatistics")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	logQueries, err := generalOptionalBool(advanced, "logqueries")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	logReplies, err := generalOptionalBool(advanced, "logreplies")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	tagReply, err := generalOptionalBool(advanced, "logtagqueryreply")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	logLocal, err := generalOptionalBool(advanced, "loglocalactions")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	logServfail, err := generalOptionalBool(advanced, "logservfail")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	serveExpired, err := generalOptionalBool(advanced, "serveexpired")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	serveReply, err := generalOptionalNumericText(advanced, "serveexpiredreplyttl")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	serveTTL, err := generalOptionalNumericText(advanced, "serveexpiredttl")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	serveReset, err := generalOptionalBool(advanced, "serveexpiredttlreset")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	serveClient, err := generalOptionalNumericText(advanced, "serveexpiredclienttimeout")
	if err != nil {
		return GeneralResourceModel{}, err
	}
	forwardingEnabled, err := generalOptionalBool(forwarding, "enabled")
	if err != nil {
		return GeneralResourceModel{}, err
	}

	verbosityValue, err := strconv.ParseInt(verbosity, 10, 64)
	if err != nil {
		return GeneralResourceModel{}, fmt.Errorf("logverbosity token %q is not an integer", verbosity)
	}
	validationValue, err := strconv.ParseInt(validation, 10, 64)
	if err != nil {
		return GeneralResourceModel{}, fmt.Errorf("valloglevel token %q is not an integer", validation)
	}

	return GeneralResourceModel{
		ID:                                types.StringValue(generalID),
		Enabled:                           types.BoolValue(enabled),
		Port:                              types.StringValue(port),
		Stats:                             stats,
		ActiveInterface:                   tfconv.SliceToSet(active),
		DNSSEC:                            dnssec,
		DNS64:                             dns64,
		DNS64Prefix:                       types.StringValue(prefix),
		NoAAAARecords:                     noAAAA,
		RegisterDHCP:                      registerDHCP,
		RegisterDHCPDomain:                types.StringValue(registerDomain),
		RegisterDHCPStatic:                registerStatic,
		NoRegisterLinkLocal6:              noLinkLocal,
		NoRegisterRecords:                 noRecords,
		TXTSupport:                        txtSupport,
		CacheFlush:                        cacheFlush,
		LocalZoneType:                     types.StringValue(zone),
		OutgoingInterface:                 tfconv.SliceToSet(outgoing),
		EnableWPAD:                        wpad,
		HideIdentity:                      hideIdentity,
		HideVersion:                       hideVersion,
		EnablePrefetchKey:                 prefetchKey,
		DNSSECStripped:                    stripped,
		AggressiveNSEC:                    types.BoolValue(aggressive),
		QnameMinStrict:                    qname,
		OutgoingNumTCP:                    outgoingTCP,
		IncomingNumTCP:                    incomingTCP,
		NumQueriesPerThread:               queries,
		OutgoingRange:                     outgoingRange,
		JostleTimeout:                     jostle,
		DiscardTimeout:                    discard,
		PrivateDomains:                    tfconv.SliceToSet(privateDomains),
		PrivateAddresses:                  tfconv.SliceToSet(privateAddresses),
		InsecureDomains:                   tfconv.SliceToSet(insecureDomains),
		EnablePrefetch:                    prefetch,
		UnwantedReplyThreshold:            threshold,
		MsgCacheSize:                      types.StringValue(msgCache),
		RRSetCacheSize:                    types.StringValue(rrsetCache),
		MaxTTL:                            maxTTL,
		MaxNegativeTTL:                    maxNegative,
		MinTTL:                            minTTL,
		HostTTL:                           hostTTL,
		KeepProbingHosts:                  keepProbing,
		NumHosts:                          numHosts,
		ExtendedStatistics:                extended,
		LogQueries:                        logQueries,
		LogReplies:                        logReplies,
		TagQueryReply:                     tagReply,
		LogLocalActions:                   logLocal,
		LogServfail:                       logServfail,
		VerbosityLevel:                    types.Int64Value(verbosityValue),
		ValidationLevel:                   types.Int64Value(validationValue),
		ServeExpiredEnabled:               serveExpired,
		ServeExpiredRecordReplyTTL:        types.StringValue(serveReply),
		ServeExpiredTTL:                   types.StringValue(serveTTL),
		ServeExpiredResetTTL:              serveReset,
		ServeExpiredClientResponseTimeout: types.StringValue(serveClient),
		ACLDefaultAction:                  types.StringValue(action),
		ForwardingEnabled:                 forwardingEnabled,
	}, nil
}

func generalLookupAttributes() map[string]dsschema.Attribute {
	boolAttr := func(desc string) dsschema.Attribute {
		return dsschema.BoolAttribute{Computed: true, MarkdownDescription: desc}
	}
	stringAttr := func(desc string) dsschema.Attribute {
		return dsschema.StringAttribute{Computed: true, MarkdownDescription: desc}
	}
	intAttr := func(desc string) dsschema.Attribute {
		return dsschema.Int64Attribute{Computed: true, MarkdownDescription: desc}
	}
	setAttr := func(desc string) dsschema.Attribute {
		return dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: desc}
	}
	return map[string]dsschema.Attribute{
		"id":                                    dsschema.StringAttribute{Required: true, MarkdownDescription: "Synthetic singleton ID. Must be `unbound-general`."},
		"enabled":                               boolAttr("Enable the resolver. A disabled flag stays false. A missing or blank value fails closed and is not replaced with the model default."),
		"port":                                  stringAttr("Listen port. A missing or blank value fails closed and is not replaced with 53."),
		"stats":                                 boolAttr("Collect query statistics. A blank optional flag stays null."),
		"active_interface":                      setAttr("Selected listen interfaces. Unselected keys are not state. An empty selection stays empty."),
		"dnssec":                                boolAttr("Enable DNSSEC validation. A blank optional flag stays null."),
		"dns64":                                 boolAttr("Enable DNS64 synthesis. A blank optional flag stays null."),
		"dns64_prefix":                          stringAttr("DNS64 prefix. An empty value stays empty and is not replaced with a well-known prefix."),
		"no_aaaa_records":                       boolAttr("Remove AAAA records. A blank optional flag stays null."),
		"register_dhcp":                         boolAttr("Register DHCP leases. A blank optional flag stays null."),
		"register_dhcp_domain":                  stringAttr("Domain appended to registered DHCP names. An empty value stays empty."),
		"register_dhcp_static":                  boolAttr("Register static DHCP mappings. A blank optional flag stays null."),
		"no_register_link_local6":               boolAttr("Do not register IPv6 link-local addresses. A blank optional flag stays null."),
		"no_register_records":                   boolAttr("Do not register system records. A blank optional flag stays null."),
		"txt_support":                           boolAttr("Serve TXT records for registered hosts. A blank optional flag stays null."),
		"cache_flush":                           boolAttr("Flush the cache on reconfigure. A blank optional flag stays null."),
		"local_zone_type":                       stringAttr("Local zone type. A missing value is not replaced with `transparent`."),
		"outgoing_interface":                    setAttr("Selected outgoing interfaces. Unselected keys are not state. An empty selection stays empty."),
		"enable_wpad":                           boolAttr("Serve WPAD records. A blank optional flag stays null."),
		"hide_identity":                         boolAttr("Refuse identity queries. A blank optional flag stays null."),
		"hide_version":                          boolAttr("Refuse version queries. A blank optional flag stays null."),
		"enable_prefetch_key":                   boolAttr("Prefetch DNSKEY records. A blank optional flag stays null."),
		"dnssec_stripped":                       boolAttr("Harden against DNSSEC-stripped responses. A blank optional flag stays null."),
		"aggressive_nsec":                       boolAttr("Use aggressive NSEC caching. A disabled flag stays false. A missing value is not replaced with the model default."),
		"qname_min_strict":                      boolAttr("Use strict QNAME minimisation. A blank optional flag stays null."),
		"outgoing_num_tcp":                      intAttr("Maximum outgoing TCP buffers. A blank value stays null. Literal 0 is kept."),
		"incoming_num_tcp":                      intAttr("Maximum incoming TCP buffers. A blank value stays null. Literal 0 is kept."),
		"num_queries_per_thread":                intAttr("Maximum simultaneous queries per thread. A blank value stays null. Literal 0 is kept."),
		"outgoing_range":                        intAttr("Outgoing port range. A blank value stays null. Literal 0 is kept."),
		"jostle_timeout":                        intAttr("Jostle timeout in milliseconds. A blank value stays null. Literal 0 is kept."),
		"discard_timeout":                       intAttr("Discard timeout in milliseconds. A blank value stays null. Literal 0 is kept."),
		"private_domains":                       setAttr("Domains allowed to contain private addresses. An empty selection stays empty."),
		"private_addresses":                     setAttr("Private address ranges. A missing or empty value fails closed and is not replaced with the model default."),
		"insecure_domains":                      setAttr("Domains excluded from DNSSEC validation. An empty selection stays empty."),
		"enable_prefetch":                       boolAttr("Prefetch cache entries before expiry. A blank optional flag stays null."),
		"unwanted_reply_threshold":              intAttr("Unwanted-reply threshold. A blank value stays null. Literal 0 is kept."),
		"msg_cache_size":                        stringAttr("Message cache size. An empty value stays empty."),
		"rrset_cache_size":                      stringAttr("RRset cache size. An empty value stays empty."),
		"max_ttl":                               intAttr("Maximum cache TTL. A blank value stays null. Literal 0 is kept."),
		"max_negative_ttl":                      intAttr("Maximum negative cache TTL. A blank value stays null. Literal 0 is kept."),
		"min_ttl":                               intAttr("Minimum cache TTL. A blank value stays null. Literal 0 is kept."),
		"host_ttl":                              intAttr("Infrastructure host cache TTL. A blank value stays null. Literal 0 is kept."),
		"keep_probing_hosts":                    boolAttr("Keep probing hosts that are down. A blank optional flag stays null."),
		"num_hosts":                             intAttr("Infrastructure host cache size. A blank value stays null. Literal 0 is kept."),
		"extended_statistics":                   boolAttr("Collect extended statistics. A blank optional flag stays null."),
		"log_queries":                           boolAttr("Log queries. A blank optional flag stays null."),
		"log_replies":                           boolAttr("Log replies. A blank optional flag stays null."),
		"tag_query_reply":                       boolAttr("Tag query and reply log lines. A blank optional flag stays null."),
		"log_local_actions":                     boolAttr("Log local-zone actions. A blank optional flag stays null."),
		"log_servfail":                          boolAttr("Log SERVFAIL responses. A blank optional flag stays null."),
		"verbosity_level":                       intAttr("Log verbosity, 0 through 5. A missing value is not replaced with the model default. Literal 0 is kept."),
		"validation_level":                      intAttr("DNSSEC validation log level, 0 through 2. A missing value fails closed. Literal 0 is kept."),
		"serve_expired_enabled":                 boolAttr("Serve expired records while refreshing. A blank optional flag stays null."),
		"serve_expired_record_reply_ttl":        stringAttr("TTL returned for expired records. An empty value stays empty."),
		"serve_expired_ttl":                     stringAttr("Maximum time to serve expired records. An empty value stays empty."),
		"serve_expired_reset_ttl":               boolAttr("Reset an expired record TTL after reply. A blank optional flag stays null."),
		"serve_expired_client_response_timeout": stringAttr("Client timeout before serving expired data. An empty value stays empty."),
		"acl_default_action":                    stringAttr("Default action for clients not matched by an ACL. A missing value is not replaced with `allow`."),
		"forwarding_enabled":                    boolAttr("Enable query forwarding. A blank optional flag stays null."),
	}
}

func generalBranchFields(name string, raw json.RawMessage) (map[string]json.RawMessage, error) {
	trimmed := bytesTrim(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return map[string]json.RawMessage{}, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return nil, fmt.Errorf("%s branch is malformed: %w", name, err)
	}
	if fields == nil {
		return map[string]json.RawMessage{}, nil
	}
	return fields, nil
}

func generalRequiredBool(fields map[string]json.RawMessage, key string) (bool, error) {
	token, present, err := generalStringToken(fields, key)
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}
	if !present || token == "" {
		return false, fmt.Errorf("%s is missing or blank; the model default is not copied", key)
	}
	switch token {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("%s token %q is not 0 or 1", key, token)
	}
}

func generalOptionalBool(fields map[string]json.RawMessage, key string) (types.Bool, error) {
	token, present, err := generalStringToken(fields, key)
	if err != nil {
		return types.BoolNull(), fmt.Errorf("%s: %w", key, err)
	}
	if !present || token == "" {
		return types.BoolNull(), nil
	}
	switch token {
	case "0":
		return types.BoolValue(false), nil
	case "1":
		return types.BoolValue(true), nil
	default:
		return types.BoolNull(), fmt.Errorf("%s token %q is not blank, 0, or 1", key, token)
	}
}

func generalRequiredPort(fields map[string]json.RawMessage) (string, error) {
	token, present, err := generalStringToken(fields, "port")
	if err != nil {
		return "", fmt.Errorf("port: %w", err)
	}
	if !present || token == "" {
		return "", errors.New("port is missing or blank; the model default is not copied")
	}
	number, err := strconv.ParseInt(token, 10, 64)
	if err != nil || number < 1 || number > 65535 {
		return "", fmt.Errorf("port token %q is not a port from 1 through 65535", token)
	}
	return token, nil
}

func generalOptionalText(fields map[string]json.RawMessage, key string) (string, error) {
	token, _, err := generalStringToken(fields, key)
	if err != nil {
		return "", fmt.Errorf("%s: %w", key, err)
	}
	if strings.ContainsAny(token, "\r\n") {
		return "", fmt.Errorf("%s contains a line break", key)
	}
	return token, nil
}

func generalOptionalPrefix(fields map[string]json.RawMessage, key string) (string, error) {
	token, err := generalOptionalText(fields, key)
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", nil
	}
	if !generalIPv6Prefix(token) {
		return "", fmt.Errorf("%s token %q is not an IPv6 prefix", key, token)
	}
	return token, nil
}

func generalOptionalCache(fields map[string]json.RawMessage, key string) (string, error) {
	token, err := generalOptionalText(fields, key)
	if err != nil {
		return "", err
	}
	if token != "" && !unboundCacheSize.MatchString(token) {
		return "", fmt.Errorf("%s token %q is not a cache size", key, token)
	}
	return token, nil
}

func generalOptionalNumericText(fields map[string]json.RawMessage, key string) (string, error) {
	token, err := generalOptionalText(fields, key)
	if err != nil {
		return "", err
	}
	if token != "" && !unboundNumericString.MatchString(token) {
		return "", fmt.Errorf("%s token %q is not numeric", key, token)
	}
	return token, nil
}

func generalOptionalInt(fields map[string]json.RawMessage, key string) (types.Int64, error) {
	token, present, err := generalStringToken(fields, key)
	if err != nil {
		return types.Int64Null(), fmt.Errorf("%s: %w", key, err)
	}
	if !present || token == "" {
		return types.Int64Null(), nil
	}
	number, err := strconv.ParseInt(token, 10, 64)
	if err != nil || number < 0 {
		return types.Int64Null(), fmt.Errorf("%s token %q is not a non-negative integer", key, token)
	}
	return types.Int64Value(number), nil
}

func generalRequiredOption(fields map[string]json.RawMessage, key string, allowed map[string]struct{}) (string, error) {
	raw, ok := fields[key]
	if !ok || len(bytesTrim(raw)) == 0 || string(bytesTrim(raw)) == "null" {
		return "", fmt.Errorf("%s is missing or blank; the model default is not copied", key)
	}
	token, err := generalOptionToken(raw, allowed)
	if err != nil {
		return "", fmt.Errorf("%s: %w", key, err)
	}
	if token == "" {
		return "", fmt.Errorf("%s is missing or blank; the model default is not copied", key)
	}
	return token, nil
}

func generalOptionToken(raw json.RawMessage, allowed map[string]struct{}) (string, error) {
	trimmed := bytesTrim(raw)
	switch trimmed[0] {
	case '"':
		var value string
		if err := json.Unmarshal(trimmed, &value); err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return "", nil
		}
		if _, ok := allowed[value]; !ok {
			return "", fmt.Errorf("token %q is outside the released option list", value)
		}
		return value, nil
	case '[':
		normalized, err := normalizeIndexedChoice(trimmed)
		if err != nil {
			return "", err
		}
		var value string
		if err := json.Unmarshal(normalized, &value); err != nil {
			return "", err
		}
		if _, ok := allowed[value]; !ok {
			return "", fmt.Errorf("token %q is outside the released option list", value)
		}
		return value, nil
	case '{':
		var rawMap map[string]struct {
			Selected json.Number `json:"selected"`
		}
		if err := json.Unmarshal(trimmed, &rawMap); err != nil {
			return "", err
		}
		selected := ""
		for key, entry := range rawMap {
			if entry.Selected.String() != "1" {
				continue
			}
			if selected != "" {
				return "", errors.New("option map contains multiple selections")
			}
			selected = key
		}
		if selected == "" {
			return "", nil
		}
		if _, ok := allowed[selected]; !ok {
			return "", fmt.Errorf("token %q is outside the released option list", selected)
		}
		return selected, nil
	default:
		return "", errors.New("option value is not a string, selected map, or selected-option array")
	}
}

func generalOptionalSet(fields map[string]json.RawMessage, key string) ([]string, error) {
	raw, ok := fields[key]
	if !ok || len(bytesTrim(raw)) == 0 || string(bytesTrim(raw)) == "null" {
		return []string{}, nil
	}
	var selected opnsense.CSVSelectedMapList
	if err := json.Unmarshal(raw, &selected); err != nil {
		return nil, fmt.Errorf("%s is malformed: %w", key, err)
	}
	return generalDropEmpty(selected), nil
}

func generalRequiredNetworks(fields map[string]json.RawMessage, key string) ([]string, error) {
	raw, ok := fields[key]
	if !ok || len(bytesTrim(raw)) == 0 || string(bytesTrim(raw)) == "null" {
		return nil, fmt.Errorf("%s is missing or empty; the model default is not copied", key)
	}
	var selected opnsense.CSVSelectedMapList
	if err := json.Unmarshal(raw, &selected); err != nil {
		return nil, fmt.Errorf("%s is malformed: %w", key, err)
	}
	values := generalDropEmpty(selected)
	if len(values) == 0 {
		return nil, fmt.Errorf("%s is missing or empty; the model default is not copied", key)
	}
	for _, value := range values {
		if !generalPrefix(value) {
			return nil, fmt.Errorf("%s token %q is not an IP prefix", key, value)
		}
	}
	return values, nil
}

func generalStringToken(fields map[string]json.RawMessage, key string) (string, bool, error) {
	raw, ok := fields[key]
	if !ok {
		return "", false, nil
	}
	trimmed := bytesTrim(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return "", true, nil
	}
	if trimmed[0] != '"' {
		return "", true, errors.New("value is not a string token")
	}
	var value string
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return "", true, err
	}
	return strings.TrimSpace(value), true, nil
}

func generalDropEmpty(values []string) []string {
	kept := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		kept = append(kept, value)
	}
	return kept
}

func generalIPv6Prefix(value string) bool {
	if strings.Count(value, "/") != 1 || strings.Contains(value, "%") {
		return false
	}
	ip, network, err := net.ParseCIDR(value)
	if err != nil || ip.To4() != nil || network.IP.To4() != nil {
		return false
	}
	return true
}

func generalPrefix(value string) bool {
	if strings.Count(value, "/") != 1 || strings.Contains(value, "%") {
		return false
	}
	address, prefix, ok := strings.Cut(value, "/")
	if !ok || prefix == "" {
		return false
	}
	if strings.Contains(address, ":") {
		return generalIPv6Prefix(value)
	}
	if !generalDottedQuad(address) {
		return false
	}
	number, err := strconv.Atoi(prefix)
	return err == nil && number >= 0 && number <= 32
}

func generalDottedQuad(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return false
		}
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 || number > 255 {
			return false
		}
	}
	return true
}

func optionSet(values ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}
