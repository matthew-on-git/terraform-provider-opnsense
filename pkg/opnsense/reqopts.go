// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import "context"

// ReqOpts configures API endpoints and reconfigure behavior for a resource.
// Each resource provides its own ReqOpts to the generic CRUD functions.
type ReqOpts struct {
	// AddEndpoint is the API path for creating a resource (e.g., "/haproxy/settings/addServer").
	AddEndpoint string
	// GetEndpoint is the API path for reading a resource by UUID.
	GetEndpoint string
	// UpdateEndpoint is the API path for updating a resource by UUID.
	UpdateEndpoint string
	// DeleteEndpoint is the API path for deleting a resource by UUID.
	DeleteEndpoint string
	// SearchEndpoint is the API path for listing/searching resources.
	SearchEndpoint string
	// ReconfigureEndpoint is the standard reconfigure path called via POST after mutations
	// (e.g., "/haproxy/service/reconfigure"). Mutually exclusive with ReconfigureFunc.
	ReconfigureEndpoint string
	// ReconfigureFunc overrides the standard reconfigure endpoint with a custom function.
	// Used by firewall filter resources for the savepoint/apply/cancelRollback flow.
	// Mutually exclusive with ReconfigureEndpoint.
	ReconfigureFunc func(ctx context.Context) error
	// Monad is the JSON wrapper key for request bodies (e.g., "server" wraps as {"server": {...}}).
	Monad string
}
