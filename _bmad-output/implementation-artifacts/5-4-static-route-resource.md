# Story 5.4: Static Route Resource

Status: done

## Story

As an operator, I want to manage static routes through Terraform.

## Acceptance Criteria

1. CRUD + import + drift detection + acceptance test + documentation
2. Schema: network (CIDR), gateway, description, disabled
3. Standard ReconfigureEndpoint

## Tasks / Subtasks

- [x] Task 1: Create model, schema, resource, test in `internal/service/system/`
- [x] Task 2: Register in exports.go
- [x] Task 3: Create docs and examples
- [x] Task 4: Run `make check`

## Dev Notes

### API: `/api/routes/routes/`

| Op | Method | Endpoint |
|----|--------|----------|
| Create | POST | `/api/routes/routes/addroute` |
| Read | GET | `/api/routes/routes/getroute/{uuid}` |
| Update | POST | `/api/routes/routes/setroute/{uuid}` |
| Delete | POST | `/api/routes/routes/delroute/{uuid}` |
| Search | GET/POST | `/api/routes/routes/searchroute` |
| Reconfigure | POST | `/api/routes/routes/reconfigure` |

**Monad:** `"route"`

### MVP Fields

| API Field | Terraform Attr | Type | Default |
|-----------|---------------|------|---------|
| `network` | `network` | String (Required) | — |
| `gateway` | `gateway` | String (Required) | — |
| `descr` | `description` | String | `""` |
| `disabled` | `enabled` | Bool (inverted) | `true` |

**Note:** `disabled` field uses inverted logic (same as DNat port forward).

## Dev Agent Record

### Agent Model Used

gpt-5.5

### Debug Log References

- Verified implementation files under `internal/service/system/`
- Verified provider registration through `system.Resources()`
- Verified docs and examples under `docs/resources/` and `examples/resources/`
- Verified `make check` passes

### Completion Notes List

- Implemented `opnsense_system_route` with CRUD, read-back after create/update, import, drift removal on not found, and standard reconfigure endpoint.
- Implemented inverted `disabled` API handling through Terraform `enabled`.
- Added acceptance test coverage for create, import, update, and destroy.
- Added generated resource documentation plus standalone resource and import examples.

### File List

- `internal/service/system/route_model.go`
- `internal/service/system/route_schema.go`
- `internal/service/system/route_resource.go`
- `internal/service/system/route_resource_test.go`
- `internal/service/system/exports.go`
- `internal/provider/provider.go`
- `docs/resources/system_route.md`
- `examples/resources/opnsense_system_route/resource.tf`
- `examples/resources/opnsense_system_route/import.sh`
