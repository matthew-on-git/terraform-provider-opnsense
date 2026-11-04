// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *caResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages released OPNsense Trust ca lifecycle operations. Operation intent is not inferred from editor reads. Private material is stored in Terraform state; use encrypted restricted state.", Attributes: map[string]schema.Attribute{
		"id":                           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "UUID of the managed object; use this for Terraform import."},
		"refid":                        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "OPNsense reference identifier used by linked resources, distinct from the UUID."},
		"description":                  schema.StringAttribute{Required: true, MarkdownDescription: "Human-readable object description. Changes alone do not generate, sign or reissue certificates."},
		"action":                       schema.StringAttribute{Optional: true, MarkdownDescription: "Explicit operation intent. Omission preserves legacy create behavior; imported state remains absent. Description-only updates never repeat it."},
		"operation_token":              schema.StringAttribute{Optional: true, MarkdownDescription: "Change this explicit token to repeat the configured action. Never automatically replay a lost operation response."},
		"ca_ref":                       schema.StringAttribute{Optional: true, MarkdownDescription: "Issuing CA reference identifier, not UUID. Empty creates a self-signed root where supported."},
		"common_name":                  schema.StringAttribute{PlanModifiers: []planmodifier.String{legacyTrustReplacement{}}, Optional: true, Computed: true, MarkdownDescription: "Subject common name used by an explicitly requested generation operation."},
		"country":                      schema.StringAttribute{PlanModifiers: []planmodifier.String{legacyTrustReplacement{}}, Optional: true, Computed: true, MarkdownDescription: "Two-letter subject country code used by an explicitly requested generation operation."},
		"city":                         schema.StringAttribute{Optional: true, MarkdownDescription: "Subject locality used by an explicitly requested generation operation."},
		"state":                        schema.StringAttribute{Optional: true, MarkdownDescription: "Subject state or province used by an explicitly requested generation operation."},
		"organization":                 schema.StringAttribute{Optional: true, MarkdownDescription: "Subject organization used by an explicitly requested generation operation."},
		"organizational_unit":          schema.StringAttribute{Optional: true, MarkdownDescription: "Subject organizational unit used by an explicitly requested generation operation."},
		"email":                        schema.StringAttribute{Optional: true, MarkdownDescription: "Subject email address used by an explicitly requested generation operation."},
		"key_type":                     schema.StringAttribute{Optional: true, MarkdownDescription: "Released key type: RSA bits 512, 1024, 2048, 3072, 4096, 7680, 8192 or curve prime256v1, secp384r1, secp521r1."},
		"digest":                       schema.StringAttribute{Optional: true, MarkdownDescription: "Generation digest: sha1, sha224, sha256, sha384 or sha512."},
		"ocsp_uri":                     schema.StringAttribute{Optional: true, MarkdownDescription: "OCSP responder URL used when generating or signing."},
		"certificate":                  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "PEM input preserved verbatim when configured; CA generation also returns its public certificate."},
		"private_key":                  schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "Sensitive PEM input preserved verbatim; not an API-read output."},
		"returned_certificate":         schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API certificate PEM; null if this object contains only a CSR."},
		"returned_private_key":         schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Sensitive API-returned private key, null when absent."},
		"valid_from":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Actual validity start returned by the API as a numeric Unix timestamp or string; not the volatile lifetime default."},
		"valid_to":                     schema.StringAttribute{Computed: true, MarkdownDescription: "Actual validity end returned by the API as a numeric Unix timestamp or string."},
		"subject":                      schema.StringAttribute{Computed: true, MarkdownDescription: "API-reconstructed subject distinguished name."},
		"lifetime":                     schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional generation lifetime in days; configured values are retained. When unconfigured, the legacy integer observation may be only an editor template default, not original validity. Use valid_from and valid_to for actual validity."},
		"returned_ca_ref":              schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API caref observation, null when absent; distinct from configured generation intent."},
		"returned_common_name":         schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API commonname observation, null when absent; distinct from configured generation intent."},
		"returned_country":             schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API country observation, null when absent; distinct from configured generation intent."},
		"returned_city":                schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API city observation, null when absent; distinct from configured generation intent."},
		"returned_state":               schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API state observation, null when absent; distinct from configured generation intent."},
		"returned_organization":        schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API organization observation, null when absent; distinct from configured generation intent."},
		"returned_organizational_unit": schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API organizationalunit observation, null when absent; distinct from configured generation intent."},
		"returned_email":               schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API email observation, null when absent; distinct from configured generation intent."},
		"returned_key_type":            schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API key_type observation, null when absent; distinct from configured generation intent."},
		"returned_digest":              schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API digest observation, null when absent; distinct from configured generation intent."},
		"returned_ocsp_uri":            schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API ocsp_uri observation, null when absent; distinct from configured generation intent."},
		"serial":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API serial observation, null when absent; distinct from configured generation intent."},
		"refcount":                     schema.StringAttribute{Computed: true, MarkdownDescription: "Authoritative API refcount observation, null when absent; distinct from configured generation intent."},
		"operation_status":             schema.StringAttribute{Computed: true, MarkdownDescription: "Last explicit operation outcome: saved or unknown. Uncertain attempts consume their action/token before POST; refresh and metadata edits cannot establish provenance or resolve unknown. Deliberate retry requires a new operation_token."},
	}}
}
