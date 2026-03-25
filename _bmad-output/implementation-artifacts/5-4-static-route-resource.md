# Story 5.4: Static Route Resource

Status: ready-for-dev

## Story

As an operator, I want to manage static routes through Terraform.

## Acceptance Criteria

1. CRUD + import + drift detection + acceptance test + documentation
2. Schema: network (CIDR), gateway, description, disabled
3. Standard ReconfigureEndpoint

## Tasks / Subtasks

- [ ] Task 1: Create model, schema, resource, test in `internal/service/system/`
- [ ] Task 2: Register in exports.go
- [ ] Task 3: Create docs and examples
- [ ] Task 4: Run `make check`

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
### Debug Log References
### Completion Notes List
### File List
