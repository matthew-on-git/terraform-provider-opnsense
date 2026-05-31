# Story 19.3: BGP Sub-Resources (aspath, communitylist, peergroup, redistribution)

Status: done

## Story
As an operator, I want to manage BGP AS-path lists, community lists, peer groups, and redistributions through Terraform.

## Acceptance Criteria
1. Four item resources on the `/api/quagga/bgp/{add,get,set,del,search}_<kind>` endpoints, standard CRUD + import, four-file pattern each.
   - `opnsense_quagga_bgp_aspath` (monad `aspath`)
   - `opnsense_quagga_bgp_communitylist` (monad `communitylist`)
   - `opnsense_quagga_bgp_peergroup` (monad `peergroup`)
   - `opnsense_quagga_bgp_redistribution` (monad `redistribution`)
2. Registered in `quagga.Resources()`; make check fully green (all 6 targets).

## Dev Notes
Modeled on existing `bgp_neighbor`. Fields from os-frr `Quagga/BGP.xml` sub-models. Field/monad accuracy = acceptance-verify ([[acceptance-testing-hardware-box]]).

## Dev Agent Record
### Completion Notes List
- 16 files across the 4 resources; build/vet/unit green; **make check fully green**.
### File List
- internal/service/quagga/bgp_aspath_*.go, bgp_communitylist_*.go, bgp_peergroup_*.go, bgp_redistribution_*.go
