// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCertificateBindingIsValidDefersUnknownSet(t *testing.T) {
	t.Parallel()

	unknown := types.SetUnknown(types.StringType)
	if !certificateBindingIsValid(context.Background(), unknown, "default-cert") {
		t.Fatal("unknown certificate set should defer validation")
	}
}

func TestCertificateBindingIsValidDefersUnknownElement(t *testing.T) {
	t.Parallel()

	certificates := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("default-cert"),
		types.StringUnknown(),
	})
	if !certificateBindingIsValid(context.Background(), certificates, "default-cert") {
		t.Fatal("set containing an unknown certificate should defer validation")
	}
}

func TestCertificateBindingIsValidRequiresConcreteDefault(t *testing.T) {
	t.Parallel()

	certificates := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("default-cert")})
	if !certificateBindingIsValid(context.Background(), certificates, "default-cert") {
		t.Fatal("certificate set containing the default should be valid")
	}
	if certificateBindingIsValid(context.Background(), certificates, "missing-cert") {
		t.Fatal("certificate set missing the default should be invalid")
	}
}

func TestCertificateBindingIsValidRejectsNullSet(t *testing.T) {
	t.Parallel()

	if certificateBindingIsValid(context.Background(), types.SetNull(types.StringType), "default-cert") {
		t.Fatal("null certificate set should be invalid")
	}
}
