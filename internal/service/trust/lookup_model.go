// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// trustReadModel contains observations only, never resource operation provenance.
type trustReadModel struct {
	ID                 types.String `tfsdk:"id"`
	RefID              types.String `tfsdk:"refid"`
	Description        types.String `tfsdk:"description"`
	CARef              types.String `tfsdk:"ca_ref"`
	CommonName         types.String `tfsdk:"common_name"`
	Country            types.String `tfsdk:"country"`
	City               types.String `tfsdk:"city"`
	State              types.String `tfsdk:"state"`
	Organization       types.String `tfsdk:"organization"`
	OrganizationalUnit types.String `tfsdk:"organizational_unit"`
	Email              types.String `tfsdk:"email"`
	Certificate        types.String `tfsdk:"certificate"`
	PrivateKey         types.String `tfsdk:"private_key"`
	CSR                types.String `tfsdk:"csr"`
	ValidFrom          types.String `tfsdk:"valid_from"`
	ValidTo            types.String `tfsdk:"valid_to"`
	Subject            types.String `tfsdk:"subject"`
	Purpose            types.String `tfsdk:"purpose"`
	KeyType            types.String `tfsdk:"key_type"`
	Digest             types.String `tfsdk:"digest"`
	CertType           types.String `tfsdk:"cert_type"`
	AltNamesDNS        types.String `tfsdk:"altnames_dns"`
	AltNamesIP         types.String `tfsdk:"altnames_ip"`
	AltNamesURI        types.String `tfsdk:"altnames_uri"`
	AltNamesEmail      types.String `tfsdk:"altnames_email"`
	Lifetime           types.Int64  `tfsdk:"lifetime"`
	Serial             types.String `tfsdk:"serial"`
	RefCount           types.String `tfsdk:"refcount"`
	InUse              types.Bool   `tfsdk:"in_use"`
	IsUser             types.Bool   `tfsdk:"is_user"`
	OCSPURI            types.String `tfsdk:"ocsp_uri"`
}

func (m *trustReadModel) fromRecord(record map[string]string, id string) {
	m.ID = types.StringValue(id)
	m.RefID = nullableTrustString(record["refid"])
	m.Description = nullableTrustString(record["descr"])
	m.CARef = nullableTrustString(record["caref"])
	m.CommonName = nullableTrustString(record["commonname"])
	m.Country = nullableTrustString(record["country"])
	m.City = nullableTrustString(record["city"])
	m.State = nullableTrustString(record["state"])
	m.Organization = nullableTrustString(record["organization"])
	m.OrganizationalUnit = nullableTrustString(record["organizationalunit"])
	m.Email = nullableTrustString(record["email"])
	m.Certificate = nullableTrustString(record["crt_payload"])
	m.PrivateKey = nullableTrustString(record["prv_payload"])
	m.CSR = nullableTrustString(record["csr_payload"])
	m.ValidFrom = nullableTrustString(record["valid_from"])
	m.ValidTo = nullableTrustString(record["valid_to"])
	m.Subject = nullableTrustString(record["name"])
	m.Purpose = nullableTrustString(record["rfc3280_purpose"])
	m.KeyType = nullableTrustString(record["key_type"])
	m.Digest = nullableTrustString(record["digest"])
	m.CertType = nullableTrustString(record["cert_type"])
	m.AltNamesDNS = nullableTrustString(record["altnames_dns"])
	m.AltNamesIP = nullableTrustString(record["altnames_ip"])
	m.AltNamesURI = nullableTrustString(record["altnames_uri"])
	m.AltNamesEmail = nullableTrustString(record["altnames_email"])
	m.Lifetime = types.Int64Null()
	if n, err := strconv.ParseInt(record["lifetime"], 10, 64); err == nil {
		m.Lifetime = types.Int64Value(n)
	}
	m.Serial = nullableTrustString(record["serial"])
	m.RefCount = nullableTrustString(record["refcount"])
	m.OCSPURI = nullableTrustString(record["ocsp_uri"])
	m.InUse = types.BoolNull()
	if v, ok := record["in_use"]; ok {
		m.InUse = types.BoolValue(v == "1")
	}
	m.IsUser = types.BoolNull()
	if v, ok := record["is_user"]; ok {
		m.IsUser = types.BoolValue(v == "1")
	}
}
