# Story 4.4: HAProxy Health Check Resource

Status: done

## Story

As an operator,
I want to manage HAProxy health checks,
So that I can configure how backends verify server availability.

## Acceptance Criteria

1. **Given** the OPNsense appliance has the `os-haproxy` plugin installed
   **When** the operator defines an `opnsense_haproxy_healthcheck` resource
   **Then** the health check is created with the specified type and configuration

2. **And** the schema includes: name, type (tcp/http/ssl/etc), interval, checkport, http_method, http_uri

3. **And** `terraform plan` with no changes shows "No changes"

4. **And** `terraform import` works by UUID

5. **And** acceptance test covers full lifecycle with CheckDestroy

## Tasks / Subtasks

- [x] Task 1: Create `healthcheck_model.go` (AC: #1, #2)
- [x] Task 2: Create `healthcheck_schema.go` (AC: #2)
- [x] Task 3: Create `healthcheck_resource.go` with CRUD + ImportState (AC: #1, #3, #4)
- [x] Task 4: Register in `exports.go` (AC: all)
- [x] Task 5: Create acceptance test (AC: #5)
- [x] Task 6: Create documentation and examples (AC: all)
- [x] Task 7: Run `make check` (AC: all)

## Dev Notes

### API Endpoints

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Create | POST | `/api/haproxy/settings/addHealthcheck` |
| Read | GET | `/api/haproxy/settings/getHealthcheck/{uuid}` |
| Update | POST | `/api/haproxy/settings/setHealthcheck/{uuid}` |
| Delete | POST | `/api/haproxy/settings/delHealthcheck/{uuid}` |
| Search | GET/POST | `/api/haproxy/settings/searchHealthchecks` |
| Reconfigure | POST | `/api/haproxy/service/reconfigure` |

**Monad key:** `"healthcheck"`

### MVP Schema

| API Field | Terraform Attribute | Terraform Type | Default |
|-----------|---------------------|----------------|---------|
| (UUID) | `id` | `types.String` (Computed) | — |
| `name` | `name` | `types.String` (Required) | — |
| `description` | `description` | `types.String` | `""` |
| `type` | `type` | `types.String` | `"http"` |
| `interval` | `interval` | `types.String` | `"2s"` |
| `checkport` | `check_port` | `types.String` | `""` |
| `http_method` | `http_method` | `types.String` | `"options"` |
| `http_uri` | `http_uri` | `types.String` | `"/"` |
| `http_version` | `http_version` | `types.String` | `"http10"` |
| `force_ssl` | `force_ssl` | `types.Bool` | `false` |

**Type values:** `tcp`, `http`, `agent`, `ldap`, `mysql`, `pgsql`, `redis`, `smtp`, `esmtp`, `ssl`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-4, Story 4.4]
- [Source: _bmad-output/planning-artifacts/prd.md#FR28]

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
