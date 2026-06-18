# Multi-Domain Edge Composition

This composition models a brownfield OPNsense HTTPS edge migrating from Ansible to Terraform. It wires HAProxy servers, backends, ACLs, a domain map, actions, HTTP and HTTPS frontends, and an ACME-issued certificate whose `cert_ref_id` is bound to HAProxy.

The example uses documentation-only hostnames and IP addresses. Replace them with appliance-specific values before applying to a live firewall.

## Migration Crosswalk

| Ansible role/task | Terraform resource |
|---|---|
| `roles/haproxy` servers and backend pools | `opnsense_haproxy_server`, `opnsense_haproxy_backend` |
| `roles/haproxy` domain map | `opnsense_haproxy_mapfile.domain_map` |
| `roles/haproxy` map routing action | `opnsense_haproxy_action.route_domain_map` with `type = "map_use_backend"` |
| `roles/haproxy` explicit host routing | `opnsense_haproxy_acl.is_tipsyhive_host` plus `opnsense_haproxy_action.route_tipsyhive_host` |
| `roles/haproxy` internal-only deny for protected apps | `opnsense_haproxy_acl.is_protected_host`, `opnsense_haproxy_acl.is_external_source`, and `opnsense_haproxy_action.deny_external_protected` |
| `roles/haproxy` HTTP to HTTPS redirect | `opnsense_haproxy_action.redirect_https` linked to `opnsense_haproxy_frontend.http_in` |
| `roles/haproxy` forwarded-proto header | `opnsense_haproxy_action.set_forwarded_proto` |
| `roles/haproxy` HTTPS listener and certificate binding | `opnsense_haproxy_frontend.https_in` with `ssl_enabled`, `certificates`, and `default_certificate` |
| `roles/acme` account, challenge, certificate signing, and polling | `opnsense_acme_account`, `opnsense_acme_challenge`, and `opnsense_acme_certificate` |
| `roles/dyndns` daemon settings and accounts | `opnsense_ddclient_settings` and `opnsense_ddclient_account` in separate Dynamic DNS configuration |
| `roles/dhcp` PXE/TFTP options | Still outside this provider until Stories 11.3 / 21.4 resolve Kea endpoint support |

## ACME Refid Binding

`opnsense_acme_certificate.edge` signs the certificate during create/update and waits for OPNsense to report a successful issuance status. The computed `cert_ref_id` is the HAProxy certificate identifier, not the ACME certificate API UUID.

Use `cert_ref_id` for HAProxy TLS binding:

```hcl
certificates        = [opnsense_acme_certificate.edge.cert_ref_id]
default_certificate = opnsense_acme_certificate.edge.cert_ref_id
```

If DNS propagation or CA response times are slow, tune `issuance_timeout` and `issuance_poll_interval` instead of adding manual sleeps outside Terraform.

## Validation

Before applying on a live appliance, validate the example against the built provider schema:

```shell
terraform fmt -check examples/compositions/multi-domain-edge
terraform -chdir=examples/compositions/multi-domain-edge init
terraform -chdir=examples/compositions/multi-domain-edge validate
```

After the first successful apply, run a second plan immediately:

```shell
terraform -chdir=examples/compositions/multi-domain-edge plan
```

The migration-readiness signal is a no-op plan: Terraform should report `No changes` for the full edge stack.

## Remaining Boundary

Epic 29 makes the HAProxy edge and ACME certificate path expressible, but DHCP PXE/TFTP options remain tracked outside Terraform. Keep those settings in Ansible or OPNsense until the Kea DHCP option endpoint work in Stories 11.3 / 21.4 is resolved.
