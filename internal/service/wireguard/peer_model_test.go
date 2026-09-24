// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestPeerModelMapsServers(t *testing.T) {
	model := PeerResourceModel{
		Enabled:       types.BoolValue(true),
		Name:          types.StringValue("cross-site-peer"),
		PublicKey:     types.StringValue("PUBLIC_KEY"),
		TunnelAddress: types.StringValue("169.254.155.1/32"),
		ServerAddress: types.StringValue("edge-01.example.invalid"),
		ServerPort:    types.StringValue("51822"),
		Keepalive:     types.Int64Value(25),
		Servers:       types.StringValue("server-uuid"),
	}

	api := model.toAPI(context.Background())
	if api.Servers != "server-uuid" {
		t.Fatalf("toAPI servers = %q, want %q", api.Servers, "server-uuid")
	}

	var read PeerResourceModel
	read.fromAPI(context.Background(), &wireguardPeerAPIResponse{
		Enabled:       "1",
		Name:          "cross-site-peer",
		PublicKey:     "PUBLIC_KEY",
		TunnelAddress: opnsense.SelectedMap("169.254.155.1/32"),
		ServerAddress: "edge-01.example.invalid",
		ServerPort:    "51822",
		Keepalive:     "25",
		Servers:       "server-uuid",
	}, "peer-uuid")

	if got := read.Servers.ValueString(); got != "server-uuid" {
		t.Fatalf("fromAPI servers = %q, want %q", got, "server-uuid")
	}
}
