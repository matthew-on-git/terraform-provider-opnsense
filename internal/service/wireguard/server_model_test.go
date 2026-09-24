// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestServerResourceModelFromAPI_preservesDescriptionWhenAPIDropsIt(t *testing.T) {
	t.Parallel()

	model := ServerResourceModel{
		Description: types.StringValue("cross-site underlay"),
	}

	model.fromAPI(context.Background(), &wireguardServerAPIResponse{
		Enabled:       "1",
		Name:          "wg0",
		Port:          "51820",
		TunnelAddress: opnsense.SelectedMap("10.10.0.1/24"),
		Description:   "",
	}, "server-uuid")

	if got, want := model.Description.ValueString(), "cross-site underlay"; got != want {
		t.Fatalf("description = %q, want %q", got, want)
	}
}

func TestServerResourceModelFromAPIUsesReturnedDescription(t *testing.T) {
	t.Parallel()

	model := ServerResourceModel{
		Description: types.StringValue("planned description"),
	}

	model.fromAPI(context.Background(), &wireguardServerAPIResponse{
		Enabled:       "1",
		Name:          "wg0",
		Port:          "51820",
		TunnelAddress: opnsense.SelectedMap("10.10.0.1/24"),
		Description:   "api description",
	}, "server-uuid")

	if got, want := model.Description.ValueString(), "api description"; got != want {
		t.Fatalf("description = %q, want %q", got, want)
	}
}

func TestServerResourceModelFromAPIInitializesMissingDescription(t *testing.T) {
	t.Parallel()

	var model ServerResourceModel

	model.fromAPI(context.Background(), &wireguardServerAPIResponse{
		Enabled:       "1",
		Name:          "wg0",
		Port:          "51820",
		TunnelAddress: opnsense.SelectedMap("10.10.0.1/24"),
		Description:   "",
	}, "server-uuid")

	if model.Description.IsNull() || model.Description.IsUnknown() {
		t.Fatalf("description was not initialized: %#v", model.Description)
	}
	if got, want := model.Description.ValueString(), ""; got != want {
		t.Fatalf("description = %q, want %q", got, want)
	}
}
