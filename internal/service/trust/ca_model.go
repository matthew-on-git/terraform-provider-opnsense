// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CAResourceModel separates configured operation intent from read observations.
type CAResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	RefID                      types.String `tfsdk:"refid"`
	Description                types.String `tfsdk:"description"`
	Action                     types.String `tfsdk:"action"`
	OperationToken             types.String `tfsdk:"operation_token"`
	CARef                      types.String `tfsdk:"ca_ref"`
	CommonName                 types.String `tfsdk:"common_name"`
	Country                    types.String `tfsdk:"country"`
	City                       types.String `tfsdk:"city"`
	State                      types.String `tfsdk:"state"`
	Organization               types.String `tfsdk:"organization"`
	OrganizationalUnit         types.String `tfsdk:"organizational_unit"`
	Email                      types.String `tfsdk:"email"`
	KeyType                    types.String `tfsdk:"key_type"`
	Digest                     types.String `tfsdk:"digest"`
	OCSPURI                    types.String `tfsdk:"ocsp_uri"`
	Certificate                types.String `tfsdk:"certificate"`
	PrivateKey                 types.String `tfsdk:"private_key"`
	ReturnedCertificate        types.String `tfsdk:"returned_certificate"`
	ReturnedPrivateKey         types.String `tfsdk:"returned_private_key"`
	ValidFrom                  types.String `tfsdk:"valid_from"`
	ValidTo                    types.String `tfsdk:"valid_to"`
	Subject                    types.String `tfsdk:"subject"`
	Lifetime                   types.Int64  `tfsdk:"lifetime"`
	ReturnedCARef              types.String `tfsdk:"returned_ca_ref"`
	ReturnedCommonName         types.String `tfsdk:"returned_common_name"`
	ReturnedCountry            types.String `tfsdk:"returned_country"`
	ReturnedCity               types.String `tfsdk:"returned_city"`
	ReturnedState              types.String `tfsdk:"returned_state"`
	ReturnedOrganization       types.String `tfsdk:"returned_organization"`
	ReturnedOrganizationalUnit types.String `tfsdk:"returned_organizational_unit"`
	ReturnedEmail              types.String `tfsdk:"returned_email"`
	ReturnedKeyType            types.String `tfsdk:"returned_key_type"`
	ReturnedDigest             types.String `tfsdk:"returned_digest"`
	ReturnedOCSPURI            types.String `tfsdk:"returned_ocsp_uri"`
	Serial                     types.String `tfsdk:"serial"`
	RefCount                   types.String `tfsdk:"refcount"`
	OperationStatus            types.String `tfsdk:"operation_status"`
}

func (m *CAResourceModel) payload() (map[string]string, error) {
	p := map[string]string{"descr": m.Description.ValueString()}
	fields := map[string]types.String{"action": m.Action, "caref": m.CARef, "commonname": m.CommonName, "country": m.Country, "city": m.City, "state": m.State, "organization": m.Organization, "organizationalunit": m.OrganizationalUnit, "email": m.Email, "key_type": m.KeyType, "digest": m.Digest, "ocsp_uri": m.OCSPURI, "crt_payload": m.Certificate, "prv_payload": m.PrivateKey}
	for name, value := range fields {
		if value.IsUnknown() {
			if name == "crt_payload" || name == "commonname" || name == "country" {
				continue
			}
			return nil, errUnknownIntent
		}
		if !value.IsNull() {
			p[name] = value.ValueString()
		}
	}
	if !m.Lifetime.IsNull() && !m.Lifetime.IsUnknown() {
		p["lifetime"] = strconv.FormatInt(m.Lifetime.ValueInt64(), 10)
	}
	return p, nil
}

func (m *CAResourceModel) fromRecord(a map[string]string, id string) {
	if m.Lifetime.IsNull() || m.Lifetime.IsUnknown() {
		m.Lifetime = types.Int64Null()
		if n, err := strconv.ParseInt(a["lifetime"], 10, 64); err == nil {
			m.Lifetime = types.Int64Value(n)
		}
	}
	m.ID = types.StringValue(id)
	m.RefID = types.StringValue(a["refid"])
	m.Description = types.StringValue(a["descr"])
	m.CommonName = nullableTrustString(a["commonname"])
	m.Country = nullableTrustString(a["country"])
	m.Certificate = nullableTrustString(a["crt_payload"])
	m.ReturnedCertificate = nullableTrustString(a["crt_payload"])
	m.ReturnedPrivateKey = nullableTrustString(a["prv_payload"])
	m.ValidFrom = nullableTrustString(a["valid_from"])
	m.ValidTo = nullableTrustString(a["valid_to"])
	m.Subject = nullableTrustString(a["name"])
	m.ReturnedCARef = nullableTrustString(a["caref"])
	m.ReturnedCommonName = nullableTrustString(a["commonname"])
	m.ReturnedCountry = nullableTrustString(a["country"])
	m.ReturnedCity = nullableTrustString(a["city"])
	m.ReturnedState = nullableTrustString(a["state"])
	m.ReturnedOrganization = nullableTrustString(a["organization"])
	m.ReturnedOrganizationalUnit = nullableTrustString(a["organizationalunit"])
	m.ReturnedEmail = nullableTrustString(a["email"])
	m.ReturnedKeyType = nullableTrustString(a["key_type"])
	m.ReturnedDigest = nullableTrustString(a["digest"])
	m.ReturnedOCSPURI = nullableTrustString(a["ocsp_uri"])
	m.Serial = nullableTrustString(a["serial"])
	m.RefCount = nullableTrustString(a["refcount"])
}

func (m *CAResourceModel) changedOperationInputs(old *CAResourceModel) bool {
	if !m.CARef.IsUnknown() && !m.CARef.Equal(old.CARef) {
		return true
	}
	if !m.CommonName.IsUnknown() && !m.CommonName.Equal(old.CommonName) {
		return true
	}
	if !m.Country.IsUnknown() && !m.Country.Equal(old.Country) {
		return true
	}
	if !m.City.IsUnknown() && !m.City.Equal(old.City) {
		return true
	}
	if !m.State.IsUnknown() && !m.State.Equal(old.State) {
		return true
	}
	if !m.Organization.IsUnknown() && !m.Organization.Equal(old.Organization) {
		return true
	}
	if !m.OrganizationalUnit.IsUnknown() && !m.OrganizationalUnit.Equal(old.OrganizationalUnit) {
		return true
	}
	if !m.Email.IsUnknown() && !m.Email.Equal(old.Email) {
		return true
	}
	if !m.KeyType.IsUnknown() && !m.KeyType.Equal(old.KeyType) {
		return true
	}
	if !m.Digest.IsUnknown() && !m.Digest.Equal(old.Digest) {
		return true
	}
	if !m.OCSPURI.IsUnknown() && !m.OCSPURI.Equal(old.OCSPURI) {
		return true
	}
	if !m.Certificate.IsUnknown() && !m.Certificate.Equal(old.Certificate) {
		return true
	}
	if !m.PrivateKey.IsUnknown() && !m.PrivateKey.Equal(old.PrivateKey) {
		return true
	}
	return !m.Lifetime.IsUnknown() && !m.Lifetime.Equal(old.Lifetime)
}

// retainKnownState builds valid provisional state when authoritative readback fails.
func (m *CAResourceModel) retainKnownState(old *CAResourceModel) {
	if m.ID.IsUnknown() {
		m.ID = types.StringNull()
		if old != nil && !old.ID.IsUnknown() {
			m.ID = old.ID
		}
	}
	if m.RefID.IsUnknown() {
		m.RefID = types.StringNull()
		if old != nil && !old.RefID.IsUnknown() {
			m.RefID = old.RefID
		}
	}
	if m.CommonName.IsUnknown() {
		m.CommonName = types.StringNull()
		if old != nil && !old.CommonName.IsUnknown() {
			m.CommonName = old.CommonName
		}
	}
	if m.Country.IsUnknown() {
		m.Country = types.StringNull()
		if old != nil && !old.Country.IsUnknown() {
			m.Country = old.Country
		}
	}
	if m.Certificate.IsUnknown() {
		m.Certificate = types.StringNull()
		if old != nil && !old.Certificate.IsUnknown() {
			m.Certificate = old.Certificate
		}
	}
	if m.ReturnedCertificate.IsUnknown() {
		m.ReturnedCertificate = types.StringNull()
		if old != nil && !old.ReturnedCertificate.IsUnknown() {
			m.ReturnedCertificate = old.ReturnedCertificate
		}
	}
	if m.ReturnedPrivateKey.IsUnknown() {
		m.ReturnedPrivateKey = types.StringNull()
		if old != nil && !old.ReturnedPrivateKey.IsUnknown() {
			m.ReturnedPrivateKey = old.ReturnedPrivateKey
		}
	}
	if m.ValidFrom.IsUnknown() {
		m.ValidFrom = types.StringNull()
		if old != nil && !old.ValidFrom.IsUnknown() {
			m.ValidFrom = old.ValidFrom
		}
	}
	if m.ValidTo.IsUnknown() {
		m.ValidTo = types.StringNull()
		if old != nil && !old.ValidTo.IsUnknown() {
			m.ValidTo = old.ValidTo
		}
	}
	if m.Subject.IsUnknown() {
		m.Subject = types.StringNull()
		if old != nil && !old.Subject.IsUnknown() {
			m.Subject = old.Subject
		}
	}
	if m.Lifetime.IsUnknown() {
		m.Lifetime = types.Int64Null()
		if old != nil && !old.Lifetime.IsUnknown() {
			m.Lifetime = old.Lifetime
		}
	}
	if m.ReturnedCARef.IsUnknown() {
		m.ReturnedCARef = types.StringNull()
		if old != nil && !old.ReturnedCARef.IsUnknown() {
			m.ReturnedCARef = old.ReturnedCARef
		}
	}
	if m.ReturnedCommonName.IsUnknown() {
		m.ReturnedCommonName = types.StringNull()
		if old != nil && !old.ReturnedCommonName.IsUnknown() {
			m.ReturnedCommonName = old.ReturnedCommonName
		}
	}
	if m.ReturnedCountry.IsUnknown() {
		m.ReturnedCountry = types.StringNull()
		if old != nil && !old.ReturnedCountry.IsUnknown() {
			m.ReturnedCountry = old.ReturnedCountry
		}
	}
	if m.ReturnedCity.IsUnknown() {
		m.ReturnedCity = types.StringNull()
		if old != nil && !old.ReturnedCity.IsUnknown() {
			m.ReturnedCity = old.ReturnedCity
		}
	}
	if m.ReturnedState.IsUnknown() {
		m.ReturnedState = types.StringNull()
		if old != nil && !old.ReturnedState.IsUnknown() {
			m.ReturnedState = old.ReturnedState
		}
	}
	if m.ReturnedOrganization.IsUnknown() {
		m.ReturnedOrganization = types.StringNull()
		if old != nil && !old.ReturnedOrganization.IsUnknown() {
			m.ReturnedOrganization = old.ReturnedOrganization
		}
	}
	if m.ReturnedOrganizationalUnit.IsUnknown() {
		m.ReturnedOrganizationalUnit = types.StringNull()
		if old != nil && !old.ReturnedOrganizationalUnit.IsUnknown() {
			m.ReturnedOrganizationalUnit = old.ReturnedOrganizationalUnit
		}
	}
	if m.ReturnedEmail.IsUnknown() {
		m.ReturnedEmail = types.StringNull()
		if old != nil && !old.ReturnedEmail.IsUnknown() {
			m.ReturnedEmail = old.ReturnedEmail
		}
	}
	if m.ReturnedKeyType.IsUnknown() {
		m.ReturnedKeyType = types.StringNull()
		if old != nil && !old.ReturnedKeyType.IsUnknown() {
			m.ReturnedKeyType = old.ReturnedKeyType
		}
	}
	if m.ReturnedDigest.IsUnknown() {
		m.ReturnedDigest = types.StringNull()
		if old != nil && !old.ReturnedDigest.IsUnknown() {
			m.ReturnedDigest = old.ReturnedDigest
		}
	}
	if m.ReturnedOCSPURI.IsUnknown() {
		m.ReturnedOCSPURI = types.StringNull()
		if old != nil && !old.ReturnedOCSPURI.IsUnknown() {
			m.ReturnedOCSPURI = old.ReturnedOCSPURI
		}
	}
	if m.Serial.IsUnknown() {
		m.Serial = types.StringNull()
		if old != nil && !old.Serial.IsUnknown() {
			m.Serial = old.Serial
		}
	}
	if m.RefCount.IsUnknown() {
		m.RefCount = types.StringNull()
		if old != nil && !old.RefCount.IsUnknown() {
			m.RefCount = old.RefCount
		}
	}
	if m.OperationStatus.IsUnknown() {
		m.OperationStatus = types.StringNull()
		if old != nil && !old.OperationStatus.IsUnknown() {
			m.OperationStatus = old.OperationStatus
		}
	}
}
