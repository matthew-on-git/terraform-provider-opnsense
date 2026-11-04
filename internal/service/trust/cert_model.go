// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CertResourceModel separates configured operation intent from read observations.
type CertResourceModel struct {
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
	CertType                   types.String `tfsdk:"cert_type"`
	PrivateKeyLocation         types.String `tfsdk:"private_key_location"`
	CSR                        types.String `tfsdk:"csr"`
	AltNamesDNS                types.String `tfsdk:"altnames_dns"`
	AltNamesIP                 types.String `tfsdk:"altnames_ip"`
	AltNamesURI                types.String `tfsdk:"altnames_uri"`
	AltNamesEmail              types.String `tfsdk:"altnames_email"`
	ReturnedCertificate        types.String `tfsdk:"returned_certificate"`
	ReturnedPrivateKey         types.String `tfsdk:"returned_private_key"`
	ValidFrom                  types.String `tfsdk:"valid_from"`
	ValidTo                    types.String `tfsdk:"valid_to"`
	Subject                    types.String `tfsdk:"subject"`
	ReturnedCSR                types.String `tfsdk:"returned_csr"`
	Purpose                    types.String `tfsdk:"purpose"`
	GeneratedPrivateKey        types.String `tfsdk:"generated_private_key"`
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
	ReturnedCertType           types.String `tfsdk:"returned_cert_type"`
	ReturnedAltNamesDNS        types.String `tfsdk:"returned_altnames_dns"`
	ReturnedAltNamesIP         types.String `tfsdk:"returned_altnames_ip"`
	ReturnedAltNamesURI        types.String `tfsdk:"returned_altnames_uri"`
	ReturnedAltNamesEmail      types.String `tfsdk:"returned_altnames_email"`
	InUse                      types.Bool   `tfsdk:"in_use"`
	IsUser                     types.Bool   `tfsdk:"is_user"`
	OperationStatus            types.String `tfsdk:"operation_status"`
}

func (m *CertResourceModel) payload() (map[string]string, error) {
	p := map[string]string{"descr": m.Description.ValueString()}
	fields := map[string]types.String{"action": m.Action, "caref": m.CARef, "commonname": m.CommonName, "country": m.Country, "city": m.City, "state": m.State, "organization": m.Organization, "organizationalunit": m.OrganizationalUnit, "email": m.Email, "key_type": m.KeyType, "digest": m.Digest, "ocsp_uri": m.OCSPURI, "crt_payload": m.Certificate, "prv_payload": m.PrivateKey, "cert_type": m.CertType, "private_key_location": m.PrivateKeyLocation, "csr_payload": m.CSR, "altnames_dns": m.AltNamesDNS, "altnames_ip": m.AltNamesIP, "altnames_uri": m.AltNamesURI, "altnames_email": m.AltNamesEmail}
	for name, value := range fields {
		if value.IsUnknown() {
			if name == "caref" {
				continue
			}
			return nil, errUnknownIntent
		}
		if !value.IsNull() {
			p[name] = value.ValueString()
		}
	}
	if m.Lifetime.IsUnknown() {
		return nil, errUnknownIntent
	}
	if !m.Lifetime.IsNull() && !m.Lifetime.IsUnknown() {
		p["lifetime"] = strconv.FormatInt(m.Lifetime.ValueInt64(), 10)
	}
	return p, nil
}

func (m *CertResourceModel) fromRecord(a map[string]string, id string) {
	m.ID = types.StringValue(id)
	m.RefID = types.StringValue(a["refid"])
	if m.CARef.IsNull() || m.CARef.IsUnknown() {
		m.CARef = types.StringValue(a["caref"])
	}
	m.Description = types.StringValue(a["descr"])
	m.ReturnedCertificate = nullableTrustString(a["crt_payload"])
	m.ReturnedPrivateKey = nullableTrustString(a["prv_payload"])
	m.ValidFrom = nullableTrustString(a["valid_from"])
	m.ValidTo = nullableTrustString(a["valid_to"])
	m.Subject = nullableTrustString(a["name"])
	m.ReturnedCSR = nullableTrustString(a["csr_payload"])
	m.Purpose = nullableTrustString(a["rfc3280_purpose"])
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
	m.ReturnedCertType = nullableTrustString(a["cert_type"])
	m.ReturnedAltNamesDNS = nullableTrustString(a["altnames_dns"])
	m.ReturnedAltNamesIP = nullableTrustString(a["altnames_ip"])
	m.ReturnedAltNamesURI = nullableTrustString(a["altnames_uri"])
	m.ReturnedAltNamesEmail = nullableTrustString(a["altnames_email"])
	m.InUse = types.BoolNull()
	if v, ok := a["in_use"]; ok {
		m.InUse = types.BoolValue(v == "1")
	}
	m.IsUser = types.BoolNull()
	if v, ok := a["is_user"]; ok {
		m.IsUser = types.BoolValue(v == "1")
	}
}

func (m *CertResourceModel) changedOperationInputs(old *CertResourceModel) bool {
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
	if !m.CertType.IsUnknown() && !m.CertType.Equal(old.CertType) {
		return true
	}
	if !m.PrivateKeyLocation.IsUnknown() && !m.PrivateKeyLocation.Equal(old.PrivateKeyLocation) {
		return true
	}
	if !m.CSR.IsUnknown() && !m.CSR.Equal(old.CSR) {
		return true
	}
	if !m.AltNamesDNS.IsUnknown() && !m.AltNamesDNS.Equal(old.AltNamesDNS) {
		return true
	}
	if !m.AltNamesIP.IsUnknown() && !m.AltNamesIP.Equal(old.AltNamesIP) {
		return true
	}
	if !m.AltNamesURI.IsUnknown() && !m.AltNamesURI.Equal(old.AltNamesURI) {
		return true
	}
	if !m.AltNamesEmail.IsUnknown() && !m.AltNamesEmail.Equal(old.AltNamesEmail) {
		return true
	}
	return !m.Lifetime.IsUnknown() && !m.Lifetime.Equal(old.Lifetime)
}

// retainKnownState builds valid provisional state when authoritative readback fails.
func (m *CertResourceModel) retainKnownState(old *CertResourceModel) {
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
	if m.CARef.IsUnknown() {
		m.CARef = types.StringNull()
		if old != nil && !old.CARef.IsUnknown() {
			m.CARef = old.CARef
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
	if m.ReturnedCSR.IsUnknown() {
		m.ReturnedCSR = types.StringNull()
		if old != nil && !old.ReturnedCSR.IsUnknown() {
			m.ReturnedCSR = old.ReturnedCSR
		}
	}
	if m.Purpose.IsUnknown() {
		m.Purpose = types.StringNull()
		if old != nil && !old.Purpose.IsUnknown() {
			m.Purpose = old.Purpose
		}
	}
	if m.GeneratedPrivateKey.IsUnknown() {
		m.GeneratedPrivateKey = types.StringNull()
		if old != nil && !old.GeneratedPrivateKey.IsUnknown() {
			m.GeneratedPrivateKey = old.GeneratedPrivateKey
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
	if m.ReturnedCertType.IsUnknown() {
		m.ReturnedCertType = types.StringNull()
		if old != nil && !old.ReturnedCertType.IsUnknown() {
			m.ReturnedCertType = old.ReturnedCertType
		}
	}
	if m.ReturnedAltNamesDNS.IsUnknown() {
		m.ReturnedAltNamesDNS = types.StringNull()
		if old != nil && !old.ReturnedAltNamesDNS.IsUnknown() {
			m.ReturnedAltNamesDNS = old.ReturnedAltNamesDNS
		}
	}
	if m.ReturnedAltNamesIP.IsUnknown() {
		m.ReturnedAltNamesIP = types.StringNull()
		if old != nil && !old.ReturnedAltNamesIP.IsUnknown() {
			m.ReturnedAltNamesIP = old.ReturnedAltNamesIP
		}
	}
	if m.ReturnedAltNamesURI.IsUnknown() {
		m.ReturnedAltNamesURI = types.StringNull()
		if old != nil && !old.ReturnedAltNamesURI.IsUnknown() {
			m.ReturnedAltNamesURI = old.ReturnedAltNamesURI
		}
	}
	if m.ReturnedAltNamesEmail.IsUnknown() {
		m.ReturnedAltNamesEmail = types.StringNull()
		if old != nil && !old.ReturnedAltNamesEmail.IsUnknown() {
			m.ReturnedAltNamesEmail = old.ReturnedAltNamesEmail
		}
	}
	if m.InUse.IsUnknown() {
		m.InUse = types.BoolNull()
		if old != nil && !old.InUse.IsUnknown() {
			m.InUse = old.InUse
		}
	}
	if m.IsUser.IsUnknown() {
		m.IsUser = types.BoolNull()
		if old != nil && !old.IsUser.IsUnknown() {
			m.IsUser = old.IsUser
		}
	}
	if m.OperationStatus.IsUnknown() {
		m.OperationStatus = types.StringNull()
		if old != nil && !old.OperationStatus.IsUnknown() {
			m.OperationStatus = old.OperationStatus
		}
	}
}
