# Story 5.3: Virtual IP Resource

Status: done

## Story

As an operator, I want to manage virtual IPs (CARP, IP Alias) through Terraform.

## Acceptance Criteria

1. CRUD + import + drift detection + acceptance test + documentation
2. Schema: interface, mode (ipalias/carp/proxyarp), subnet, subnet_bits, vhid, description
3. Standard ReconfigureEndpoint

## Tasks / Subtasks

- [x] Task 1: Create model, schema, resource, test in `internal/service/system/`
- [x] Task 2: Register in exports.go
- [x] Task 3: Create docs and examples
- [x] Task 4: Run `make check`

## Dev Notes

### API: `/api/interfaces/vip_settings/`

| Op | Method | Endpoint |
|----|--------|----------|
| Create | POST | `/api/interfaces/vip_settings/add_item` |
| Read | GET | `/api/interfaces/vip_settings/get_item/{uuid}` |
| Update | POST | `/api/interfaces/vip_settings/set_item/{uuid}` |
| Delete | POST | `/api/interfaces/vip_settings/del_item/{uuid}` |
| Search | GET/POST | `/api/interfaces/vip_settings/search_item` |
| Reconfigure | POST | `/api/interfaces/vip_settings/reconfigure` |

**Monad:** `"vip"`

### MVP Fields

| API Field | Terraform Attr | Type | Default |
|-----------|---------------|------|---------|
| `interface` | `interface` | String (Required) | — |
| `mode` | `mode` | String | `"ipalias"` |
| `subnet` | `address` | String (Required) | — |
| `subnet_bits` | `subnet_bits` | Int64 (Required) | — |
| `descr` | `description` | String | `""` |
| `vhid` | `vhid` | Int64 | — |
| `password` | `password` | String (Sensitive) | `""` |
| `advbase` | `adv_base` | Int64 | `1` |
| `advskew` | `adv_skew` | Int64 | `0` |

## Dev Agent Record

### Agent Model Used

gpt-5.5

### Debug Log References

- Verified implementation files under `internal/service/system/`
- Verified provider registration through `system.Resources()`
- Verified docs and examples under `docs/resources/` and `examples/resources/`
- Verified `make check` passes

### Completion Notes List

- Implemented `opnsense_system_vip` with CRUD, read-back after create/update, import, drift removal on not found, and standard reconfigure endpoint.
- Added acceptance test coverage for create, import, update, and destroy.
- Added generated resource documentation plus standalone resource and import examples.

### File List

- `internal/service/system/vip_model.go`
- `internal/service/system/vip_schema.go`
- `internal/service/system/vip_resource.go`
- `internal/service/system/vip_resource_test.go`
- `internal/service/system/exports.go`
- `internal/provider/provider.go`
- `docs/resources/system_vip.md`
- `examples/resources/opnsense_system_vip/resource.tf`
- `examples/resources/opnsense_system_vip/import.sh`
