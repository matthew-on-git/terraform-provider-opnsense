# Story 5.2: VLAN Resource

Status: ready-for-dev

## Story

As an operator, I want to manage VLAN assignments through Terraform, so that I can define network segmentation as code.

## Acceptance Criteria

1. CRUD + import + drift detection + acceptance test + documentation
2. Schema: parent interface, VLAN tag (1-4094), priority, protocol, description, device name
3. Standard ReconfigureEndpoint

## Tasks / Subtasks

- [ ] Task 1: Create model, schema, resource, test in `internal/service/system/`
- [ ] Task 2: Create `exports.go`, register in provider
- [ ] Task 3: Create docs and examples
- [ ] Task 4: Run `make check`

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
### Debug Log References
### Completion Notes List
### File List
