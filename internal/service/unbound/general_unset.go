// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import "github.com/hashicorp/terraform-plugin-framework/types"

// Preserve appliance defaults as null instead of inventing a numeric zero.
func normalizeUnsetGeneralIntegers(state *GeneralResourceModel, api *generalAPIResponse) {
	if api.OutgoingNumTCP == "" {
		state.OutgoingNumTCP = types.Int64Null()
	}
	if api.IncomingNumTCP == "" {
		state.IncomingNumTCP = types.Int64Null()
	}
	if api.NumQueriesPerThread == "" {
		state.NumQueriesPerThread = types.Int64Null()
	}
	if api.OutgoingRange == "" {
		state.OutgoingRange = types.Int64Null()
	}
	if api.JostleTimeout == "" {
		state.JostleTimeout = types.Int64Null()
	}
	if api.DiscardTimeout == "" {
		state.DiscardTimeout = types.Int64Null()
	}
	if api.UnwantedReplyThreshold == "" {
		state.UnwantedReplyThreshold = types.Int64Null()
	}
	if api.MaxTTL == "" {
		state.MaxTTL = types.Int64Null()
	}
	if api.MaxNegativeTTL == "" {
		state.MaxNegativeTTL = types.Int64Null()
	}
	if api.MinTTL == "" {
		state.MinTTL = types.Int64Null()
	}
	if api.HostTTL == "" {
		state.HostTTL = types.Int64Null()
	}
	if api.NumHosts == "" {
		state.NumHosts = types.Int64Null()
	}
	if api.VerbosityLevel == "" {
		state.VerbosityLevel = types.Int64Null()
	}
	if api.ValidationLevel == "" {
		state.ValidationLevel = types.Int64Null()
	}
}
