# Story 5.5: Gateway Resource

Status: done

## Story

As an operator, I want to manage gateways through Terraform.

## Acceptance Criteria

1. CRUD + import + drift detection + acceptance test + documentation
2. Schema: name, interface, gateway address, ip_protocol, weight, priority, monitor settings
3. Standard ReconfigureEndpoint. Uses `disabled` field (inverted logic).

## Tasks / Subtasks

- [x] Task 1: Create model, schema, resource, test in `internal/service/system/`
- [x] Task 2: Register in exports.go
- [x] Task 3: Create docs and examples
- [x] Task 4: Run `make check`

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

gpt-5.5

### Debug Log References

- Verified implementation files under `internal/service/system/`
- Verified provider registration through `system.Resources()`
- Verified docs and examples under `docs/resources/` and `examples/resources/`
- Verified `make check` passes

### Completion Notes List

- Implemented `opnsense_system_gateway` with CRUD, read-back after create/update, import, drift removal on not found, and standard reconfigure endpoint.
- Implemented inverted `disabled` API handling through Terraform `enabled`.
- Added acceptance test coverage for create, import, update, and destroy.
- Added generated resource documentation plus standalone resource and import examples.

### File List

- `internal/service/system/gateway_model.go`
- `internal/service/system/gateway_schema.go`
- `internal/service/system/gateway_resource.go`
- `internal/service/system/gateway_resource_test.go`
- `internal/service/system/exports.go`
- `internal/provider/provider.go`
- `docs/resources/system_gateway.md`
- `examples/resources/opnsense_system_gateway/resource.tf`
- `examples/resources/opnsense_system_gateway/import.sh`
