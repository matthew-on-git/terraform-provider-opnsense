# Story 19.6: RIP General Resource (singleton)

Status: done

## Story
As an operator, I want to manage RIP routing configuration through Terraform.

## Acceptance Criteria
1. `opnsense_quagga_rip` singleton (GetSingleton/UpdateSingleton), get/set `/api/quagga/rip`, monad `rip`, id `rip`.
2. Fields: enabled, version (1-2), networks (set), redistribute (set), default_metric. make check fully green.

## Dev Notes
Singleton pattern per [[19-1-frr-general-settings-resource]]. Fields from os-frr `Quagga/RIP.xml`.

## Dev Agent Record
### Completion Notes List
- `rip_{model,schema,resource,resource_test}.go`; registered; make check fully green.
### File List
- internal/service/quagga/rip_model.go, rip_schema.go, rip_resource.go, rip_resource_test.go
