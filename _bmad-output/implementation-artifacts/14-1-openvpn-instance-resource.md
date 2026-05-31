# Story 14.1: OpenVPN Instance Resource

Status: done

## Story

As an operator,
I want to manage OPNsense OpenVPN instances (server/client) through Terraform,
so that I can define VPN endpoints as code.

## Acceptance Criteria

1. New package `internal/service/openvpn` with `opnsense_openvpn_instance` resource, full CRUD + ImportState, following the four-file pattern (`instance_resource.go`, `instance_schema.go`, `instance_model.go`, `instance_resource_test.go`) + `exports.go`.
2. ReqOpts: add `/api/openvpn/instances/add`, get `/get`, set `/set`, del `/del`, search `/search`, reconfigure `/api/openvpn/service/reconfigure`, monad `instance`.
3. Curated high-value field set (server + client use): `enabled`, `role` (client/server), `description`, `dev_type` (tun/tap/ovpn), `protocol` (proto), `port`, `local`, `server` (tunnel net), `topology`, `ca`, `cert`, `tls_key`, `data_ciphers` (set), `auth`, `dns_servers` (set), `push_route` (set), `redirect_gateway` (set), `max_clients`, `keepalive_interval`, `keepalive_timeout`, `verb`.
4. State populated from API read-back after create/update; bools convert via `BoolToString`/`StringToBool`; single-option fields use `SelectedMap`; multi-option/CSV fields use `SelectedMapList` ↔ `types.Set` joined by comma (pattern from `haproxy/backend_model.go`).
5. Registered in `openvpn.Resources()` and wired into `internal/provider/provider.go`.
6. Acceptance test covering create→import→update lifecycle (gated by TF_ACC). Example HCL in `examples/resources/opnsense_openvpn_instance/resource.tf`.
7. `make check` green except the externally-tracked stdlib security target (dev-toolchain#50).

## Dev Notes

- Template: `internal/service/haproxy/backend_*.go` (SelectedMap, SelectedMapList↔Set CSV) and `internal/service/system/vlan_*.go` (basic CRUD shape). [Source: internal/service/haproxy/backend_model.go]
- Field inventory from OPNsense `OpenVPN.xml` Instances model. JSON keys with hyphens (`data-ciphers`, `dns_servers`, `push_route`) are valid Go json tags.
- Multi-select OptionFields (`data-ciphers`, `redirect_gateway`) and multi NetworkFields (`dns_servers`, `push_route`) come back as `SelectedMapList`/CSV; send as comma-joined strings.
- `ca`/`cert`/`tls_key` are UUID references (CertificateField/ModelRelationField) — plain string attributes; real values validated against live box (acceptance only).
- Endpoints use `add/get/set/del` (NOT `add_item`). [Source: docs.opnsense.org/development/api/core/openvpn.html]
- Field accuracy is acceptance-verified only (no live box this session); build/lint/unit + make check confirm structure.

### References
- [Source: internal/service/haproxy/backend_resource.go / backend_model.go / backend_schema.go]
- [Source: _bmad-output/planning-artifacts/core-config-gap-analysis.md#VPN]

## Dev Agent Record
### Agent Model Used
### Completion Notes List
### File List
