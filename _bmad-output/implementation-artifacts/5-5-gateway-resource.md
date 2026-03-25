# Story 5.5: Gateway Resource

Status: ready-for-dev

## Story

As an operator, I want to manage gateways through Terraform.

## Acceptance Criteria

1. CRUD + import + drift detection + acceptance test + documentation
2. Schema: name, interface, gateway address, ip_protocol, weight, priority, monitor settings
3. Standard ReconfigureEndpoint. Uses `disabled` field (inverted logic).

## Tasks / Subtasks

- [ ] Task 1: Create model, schema, resource, test in `internal/service/system/`
- [ ] Task 2: Register in exports.go
- [ ] Task 3: Create docs and examples
- [ ] Task 4: Run `make check`

## Dev Notes

### API: `/api/routing/settings/`

| Op | Method | Endpoint |
|----|--------|----------|
| Create | POST | `/api/routing/settings/add_gateway` |
| Read | GET | `/api/routing/settings/get_gateway/{uuid}` |
| Update | POST | `/api/routing/settings/set_gateway/{uuid}` |
| Delete | POST | `/api/routing/settings/del_gateway/{uuid}` |
| Search | GET | `/api/routing/settings/search_gateway` |
| Reconfigure | POST | `/api/routing/settings/reconfigure` |

**Monad:** `"gateway"`

### MVP Fields

| API Field | Terraform Attr | Type | Default |
|-----------|---------------|------|---------|
| `disabled` | `enabled` | Bool (inverted) | `true` |
| `name` | `name` | String (Required) | — |
| `descr` | `description` | String | `""` |
| `interface` | `interface` | String (Required) | — |
| `ipprotocol` | `ip_protocol` | String | `"inet"` |
| `gateway` | `gateway` | String (Required) | — |
| `defaultgw` | `default_gateway` | Bool | `false` |
| `monitor_disable` | `monitor_disable` | Bool | `true` |
| `weight` | `weight` | Int64 | `1` |
| `priority` | `priority` | Int64 | `255` |

## Dev Agent Record

### Agent Model Used
### Debug Log References
### Completion Notes List
### File List
