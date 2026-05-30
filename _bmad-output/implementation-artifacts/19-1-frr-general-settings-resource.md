# Story 19.1: FRR General Settings Resource (singleton)

Status: done

(Also satisfies the long-deferred Epic 7 story 7-1.)

## Story
As an operator, I want to manage FRR (Quagga) general service settings through Terraform, so that I can enable/disable the routing service and set the profile as code.

## Acceptance Criteria
1. `opnsense_quagga_general` singleton resource (Read/Create=apply/Update/Delete=no-op/Import), using `opnsense.GetSingleton`/`UpdateSingleton` (Epic 13).
2. ReqOpts: get `/api/quagga/general/get`, set `/api/quagga/general/set`, reconfigure `/api/quagga/service/reconfigure`, monad `general`. Synthetic id `general`.
3. Fields: enabled, profile (traditional/datacenter), enable_carp, enable_syslog, enable_snmp, syslog_level, firewall_rules.
4. make check fully green.

## Dev Notes
First singleton resource — establishes the pattern: Create = UpdateSingleton + read-back; Delete = no-op (appliance config persists); Import sets fixed id. Fields from os-frr `Quagga/General.xml`. Requires `os-frr` plugin (acceptance only).

## Dev Agent Record
### Completion Notes List
- Implemented `general_{model,schema,resource,resource_test}.go`; registered in `quagga.Resources()`. Exercises Epic 13 singleton client. build/vet/unit green; **make check fully green (all 6)**. Field accuracy = acceptance-verify ([[acceptance-testing-hardware-box]]).
### File List
- internal/service/quagga/general_model.go, general_schema.go, general_resource.go, general_resource_test.go
- examples/resources/opnsense_quagga_general/resource.tf
