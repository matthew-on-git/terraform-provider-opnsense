# Story 14.2: OpenVPN Client Overwrite Resource

Status: done

## Story
As an operator, I want to manage OpenVPN client-specific overrides (CCD) through Terraform, so that I can inject per-client routes, tunnel networks, and DNS as code.

## Acceptance Criteria
1. `opnsense_openvpn_client_overwrite` resource, four-file pattern, full CRUD + import.
2. ReqOpts: `/api/openvpn/client_overwrites/{add,get,set,del,search}`, reconfigure `/api/openvpn/service/reconfigure`, monad `overwrite`.
3. Fields: enabled, common_name (required), description, servers (set), block, push_reset, tunnel_network, local_networks (set), remote_networks (set), dns_servers (set).
4. make check green except externally-tracked stdlib security (dev-toolchain#50).

## Dev Notes
Curated subset of the OpenVPN Overwrites model. Set fields use CSV (NetworkField) or SelectedMapList (servers) per [[14-1-openvpn-instance-resource]] helpers. Template: haproxy/backend.

## Dev Agent Record
### Completion Notes List
- Implemented `client_overwrite_{model,schema,resource,resource_test}.go`; registered in `openvpn.Resources()`. build/vet/unit green; make check passes lint/format/test/scan/docs. Field/monad accuracy = acceptance-verify (no live box).
### File List
- internal/service/openvpn/client_overwrite_model.go, client_overwrite_schema.go, client_overwrite_resource.go, client_overwrite_resource_test.go
- examples/resources/opnsense_openvpn_client_overwrite/resource.tf
