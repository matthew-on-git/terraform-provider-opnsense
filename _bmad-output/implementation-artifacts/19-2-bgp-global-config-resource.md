# Story 19.2: BGP Global Configuration Resource (singleton)

Status: done

(Also satisfies the long-deferred Epic 7 story 7-2.)

## Story
As an operator, I want to manage BGP global configuration through Terraform, so that I can define AS number, router ID, and advertised networks as code.

## Acceptance Criteria
1. `opnsense_quagga_bgp_global` singleton resource using `GetSingleton`/`UpdateSingleton`.
2. ReqOpts: get `/api/quagga/bgp/get`, set `/api/quagga/bgp/set`, reconfigure `/api/quagga/service/reconfigure`, monad `bgp`. Synthetic id `bgp`.
3. Fields: enabled, as_number (required), router_id, distance, graceful_restart, network_import_check, enforce_first_as, log_neighbor_changes, networks (set), maximum_paths, maximum_paths_ibgp.
4. make check fully green.

## Dev Notes
Singleton pattern per [[19-1-frr-general-settings-resource]]. `networks` is a CSVListField (CSV string ↔ types.Set). Fields from os-frr `Quagga/BGP.xml`. Requires `os-frr` + an enabled FRR general.

## Dev Agent Record
### Completion Notes List
- Implemented `bgp_global_{model,schema,resource,resource_test}.go` + local set/int helpers; registered in `quagga.Resources()`. build/vet/unit green; **make check fully green**. Field accuracy = acceptance-verify.
### File List
- internal/service/quagga/bgp_global_model.go, bgp_global_schema.go, bgp_global_resource.go, bgp_global_resource_test.go
- examples/resources/opnsense_quagga_bgp_global/resource.tf
