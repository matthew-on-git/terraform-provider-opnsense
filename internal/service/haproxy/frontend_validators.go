// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigValidators = &frontendResource{}
	_ resource.ResourceWithModifyPlan       = &frontendResource{}
)

func (r *frontendResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{frontendCertificateValidator{}}
}

type frontendCertificateValidator struct{}

func (frontendCertificateValidator) Description(_ context.Context) string {
	return "validates HAProxy frontend certificate relationships"
}

func (v frontendCertificateValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (frontendCertificateValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config FrontendResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.SSLEnabled.IsUnknown() || config.SSLEnabled.IsNull() || !config.SSLEnabled.ValueBool() {
		return
	}
	if config.DefaultCertificate.IsNull() || config.DefaultCertificate.IsUnknown() || config.DefaultCertificate.ValueString() == "" {
		return
	}
	if !certificateBindingIsValid(ctx, config.Certificates, config.DefaultCertificate.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_certificate"),
			"Invalid HAProxy Frontend Certificate Binding",
			"default_certificate must be included in certificates when ssl_enabled is true.",
		)
	}
}

// certificateBindingIsValid defers validation when the certificate set is
// unknown because it may contain a computed cert_ref_id from another resource.
// Once the set is known, the default certificate must be present concretely.
func certificateBindingIsValid(ctx context.Context, certificates types.Set, defaultCertificate string) bool {
	if certificates.IsUnknown() {
		return true
	}
	for _, element := range certificates.Elements() {
		if element.IsUnknown() {
			return true
		}
	}
	return !certificates.IsNull() && stringSetContains(ctx, certificates, defaultCertificate)
}

func (r *frontendResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan FrontendResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || plan.SSLEnabled.IsUnknown() || plan.SSLEnabled.IsNull() || plan.SSLEnabled.ValueBool() {
		return
	}

	certsConfigured := !plan.Certificates.IsNull() && !plan.Certificates.IsUnknown() && len(plan.Certificates.Elements()) > 0
	defaultConfigured := !plan.DefaultCertificate.IsNull() && !plan.DefaultCertificate.IsUnknown() && plan.DefaultCertificate.ValueString() != ""
	if certsConfigured || defaultConfigured {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("certificates"),
			"Ignoring HAProxy Frontend Certificates",
			"certificates and default_certificate are ignored when ssl_enabled is false.",
		)
	}

	plan.Certificates = types.SetValueMust(types.StringType, []attr.Value{})
	plan.DefaultCertificate = types.StringValue("")
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func stringSetContains(ctx context.Context, set types.Set, needle string) bool {
	var values []string
	set.ElementsAs(ctx, &values, false)
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
