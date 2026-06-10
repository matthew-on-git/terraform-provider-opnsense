# Story 5.2: VLAN Resource

Status: done

## Story

As an operator, I want to manage VLAN assignments through Terraform, so that I can define network segmentation as code.

## Acceptance Criteria

1. CRUD + import + drift detection + acceptance test + documentation
2. Schema: parent interface, VLAN tag (1-4094), priority, protocol, description, device name
3. Standard ReconfigureEndpoint

## Tasks / Subtasks

- [x] Task 1: Create model, schema, resource, test in `internal/service/system/`
- [x] Task 2: Create `exports.go`, register in provider
- [x] Task 3: Create docs and examples
- [x] Task 4: Run `make check`

## Dev Notes

### API: `/api/interfaces/vlan_settings/`

| Op | Method | Endpoint |
|----|--------|----------|
| Create | POST | `/api/interfaces/vlan_settings/add_item` |
| Read | GET | `/api/interfaces/vlan_settings/get_item/{uuid}` |
| Update | POST | `/api/interfaces/vlan_settings/set_item/{uuid}` |
| Delete | POST | `/api/interfaces/vlan_settings/del_item/{uuid}` |
| Search | GET/POST | `/api/interfaces/vlan_settings/search_item` |
| Reconfigure | POST | `/api/interfaces/vlan_settings/reconfigure` |

**Monad:** `"vlan"`

### MVP Fields

| API Field | Terraform Attr | Type | Default |
|-----------|---------------|------|---------|
| `if` | `parent_interface` | String (Required) | — |
| `tag` | `tag` | Int64 (Required) | — |
| `pcp` | `priority` | Int64 | `0` |
| `proto` | `proto` | String | `""` |
| `descr` | `description` | String | `""` |
| `vlanif` | `device` | String (Required) | — |

## Dev Agent Record

### Agent Model Used

gpt-5.5

### Debug Log References

- Verified implementation files under `internal/service/system/`
- Verified provider registration through `system.Resources()`
- Verified docs and examples under `docs/resources/` and `examples/resources/`
- Verified `make check` passes

### Completion Notes List

- Implemented `opnsense_system_vlan` with CRUD, read-back after create/update, import, drift removal on not found, and standard reconfigure endpoint.
- Added acceptance test coverage for create, import, update, and destroy.
- Added generated resource documentation plus standalone resource and import examples.

### File List

- `internal/service/system/vlan_model.go`
- `internal/service/system/vlan_schema.go`
- `internal/service/system/vlan_resource.go`
- `internal/service/system/vlan_resource_test.go`
- `internal/service/system/exports.go`
- `internal/provider/provider.go`
- `docs/resources/system_vlan.md`
- `examples/resources/opnsense_system_vlan/resource.tf`
- `examples/resources/opnsense_system_vlan/import.sh`
