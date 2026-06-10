---
title: Registry Documentation Audit
date: 2026-06-02
author: BMad PM
status: current
inputs:
  - docs/index.md
  - templates/index.md.tmpl
  - docs/resources/*.md
  - docs/data-sources/*.md
  - templates/data-sources/*.md.tmpl
  - examples/data-sources/*/data-source.tf
  - support-matrix.md
---

# Registry Documentation Audit

## Summary

The local Registry documentation source is internally consistent after Epic 25B:

| Check | Result |
|---|---:|
| Resource docs | 90 |
| Data-source docs | 76 |
| Data-source templates | 43 |
| Data-source examples | 43 |
| Remaining resource-matching data-source gaps | 15 |

The Terraform Registry web application is JavaScript-rendered and was not
machine-readable through the available fetch path, so this audit uses local
Registry source files under `docs/`, `templates/`, and `examples/` as the source
of truth. No browser-rendered Registry review was recorded in this story; that
remains a post-release human verification item after the next published update.

The 15 remaining resource-matching data-source gaps are provider capabilities
that do not yet have safe read-only lookup counterparts. They are separate from
the 33 generated-only data-source docs that exist but lack custom template polish.

## Findings and Fixes

| Finding | Severity | Action |
|---|---|---|
| Provider index counts are current at 90 resources, 76 data sources, 90 resource docs, and 76 data-source docs. | Info | No fix required. |
| Import guidance now correctly distinguishes UUID-backed resources from singleton/settings-style resources. | Info | No fix required. |
| Story 26.2 still referenced the original v0.1.0 baseline of 34 data sources. | Medium | Updated Story 26.2 to require 76 supported data sources after Epic 25B completion. |
| Generated-only docs are schema-accurate but sparse. 33 data-source docs lack custom templates/examples. | Medium | Recorded as follow-up for Registry docs hardening. |
| 59 resource docs lack custom import guidance/templates. | Medium | Recorded as follow-up for migration/import docs work. |

## Representative Spot Checks

| Area | Document | Result |
|---|---|---|
| Provider index | `docs/index.md` | Covers authentication, environment variables, minimum OPNsense version, permissions, support counts, and migration/import sequencing. |
| Firewall resource | `docs/resources/firewall_filter_rule.md` | Has examples, savepoint warning, generated schema, and UUID import guidance. |
| HAProxy data source | `docs/data-sources/haproxy_backend.md` | Has UUID lookup example and read-only attributes. |
| WireGuard data source | `docs/data-sources/wireguard_server.md` | Correctly documents omitted `private_key` write-only material. |
| Routing resource | `docs/resources/quagga_bgp_neighbor.md` | Has example, schema, plugin requirement, and UUID import command. |
| Traffic shaper data source | `docs/data-sources/trafficshaper_pipe.md` | Generated schema is accurate, but page lacks custom example/subcategory polish. |
| Trust certificate resource | `docs/resources/trust_cert.md` | Correctly documents write-only certificate/private-key material, but lacks custom import guidance. |

## Residual Risks

- The Registry-rendered page still needs a human/browser spot check after the next release because automated fetch only returns the JavaScript shell; this is outside the automated acceptance scope for this story.
- Generated docs without templates have weak user guidance even when the schema is technically correct.
- Import guidance is uneven across resources; Story 26.3 has since added broad migration/import guidance, while per-resource template polish remains a follow-up.
- Sparse generated data-source pages remain a docs-hardening backlog item; Story 26.2 scoped the public support matrix instead of broad per-data-source templates.

## Recommended Follow-Up

1. Keep the public support matrix current as Stories 27.x and 28.x adjust Coming / Needs research / Upstream-blocked classifications.
2. Add custom templates/examples for the highest-value generated-only data-source pages if Registry usability becomes the next docs-hardening priority.
3. Perform a human/browser Registry spot check after the next published release to confirm the rendered page matches local generated docs.
