# Story 19.7: FRR Static Routing (general + routes)

Status: done

## Story
As an operator, I want to enable FRR static routing and manage static routes through Terraform.

## Acceptance Criteria
1. `opnsense_quagga_static` singleton (enabled) — get/set `/api/quagga/static`, monad `static`, id `static`.
2. `opnsense_quagga_static_route` item — `/api/quagga/static/{add,get,set,del,search}_route`, monad `route`. Fields: enabled, network (req), gateway, interface, bfd, description.
3. Registered; make check fully green.

## Dev Notes
Singleton + item pattern. Fields from os-frr `Quagga/STATICd.xml`.

## Dev Agent Record
### Completion Notes List
- `static_general_*.go` + `static_route_*.go`; registered; make check fully green.
### File List
- internal/service/quagga/static_general_*.go, static_route_*.go
