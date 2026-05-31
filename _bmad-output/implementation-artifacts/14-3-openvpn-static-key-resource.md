# Story 14.3: OpenVPN Static Key Resource

Status: done

## Story
As an operator, I want to manage OpenVPN TLS static keys through Terraform, so that I can provision tls-crypt/tls-auth keys referenced by instances as code.

## Acceptance Criteria
1. `opnsense_openvpn_static_key` resource, four-file pattern, full CRUD + import.
2. ReqOpts: `/api/openvpn/instances/{add,get,set,del,search}_static_key`, reconfigure `/api/openvpn/service/reconfigure`, monad `static_key`.
3. Fields: mode (default crypt), key (Required, Sensitive), description. Import ignores `key`.
4. make check green except externally-tracked stdlib security (dev-toolchain#50).

## Dev Notes
Smallest OpenVPN resource. `key` is Sensitive; referenced by `opnsense_openvpn_instance.tls_key`. Template: system/vlan.

## Dev Agent Record
### Completion Notes List
- Implemented `static_key_{model,schema,resource,resource_test}.go`; registered in `openvpn.Resources()`. build/vet/unit green; make check passes lint/format/test/scan/docs. Field/monad accuracy = acceptance-verify (no live box).
### File List
- internal/service/openvpn/static_key_model.go, static_key_schema.go, static_key_resource.go, static_key_resource_test.go
- examples/resources/opnsense_openvpn_static_key/resource.tf
