---
title: System Tunables and Sysctl Research
date: 2026-06-09
author: BMad Dev
status: current
inputs:
  - OPNsense core API docs fetched 2026-06-09
  - TunablesController.php from opnsense/core master
  - Tunables.xml from opnsense/core master
  - resource-gap-verification.md
  - core-config-gap-analysis.md
  - support-matrix.md
---

# System Tunables and Sysctl Research

Story 28.3 reviewed the OPNsense `core/tunables` API surface to classify tunables/sysctl support. No Terraform resource or data source was implemented in this story.

## Source Evidence

Published OPNsense core API docs list `TunablesController.php` under the Core API. Current source extends `ApiMutableModelControllerBase`, uses internal model name `sysctl`, model class `OPNsense\Core\Tunables`, and exposes item CRUD/search plus `reset` and `reconfigure` operations.

`Tunables.xml` mounts at `//sysctl` and defines `item` entries with `TunableField`. Persistent fields are `tunable`, `value`, and `descr`. `default_value` and `type` are volatile non-persistent model fields and should not be managed as Terraform configuration.

## Endpoint Table

| Endpoint | Method | Wrapper / model path | Lifecycle | Terraform classification |
|---|---|---|---|---|
| `/api/core/tunables/add_item` | `POST` | wrapper `sysctl`, model path `item` | Adds persistent tunable config. | UUID item resource candidate. |
| `/api/core/tunables/get_item/{uuid?}` | `GET` | wrapper `sysctl`, model path `item` | Reads a persistent tunable item. | Resource read/import and data-source candidate. |
| `/api/core/tunables/set_item/{uuid}` | `POST` | wrapper `sysctl`, model path `item` | Updates persistent tunable config; if `{uuid}` is any non-UUID-like key, controller generates a new UUID. | UUID item resource candidate; name/key-as-create behavior needs live validation before relying on it. |
| `/api/core/tunables/del_item/{uuid}` | `POST` | model path `item` | Deletes persistent tunable config. | UUID item resource candidate. |
| `/api/core/tunables/search_item` | `GET,POST` | model path `item`, root `sysctl` | Searches persistent tunable config. | Data-source/listing support candidate. |
| `/api/core/tunables/reconfigure` | `POST` | n/a | Restarts `login` and `sysctl`; returns `ok` only when both restart calls succeed. | Required apply/reconfigure step after mutations. |
| `/api/core/tunables/reset` | `POST` | n/a | Replaces tunables from factory defaults. | Operational action; not part of resource CRUD. |
| `/api/core/tunables/get` / `/api/core/tunables/set` | `GET` / `POST` | root model | Generic mutable-model endpoints listed in docs. | Not needed for item resource; avoid unless a singleton use case emerges. |

## Lifecycle Classification

Tunables are **durable configuration CRUD with operational apply side effects**:

- Persistent config is stored under `//sysctl/item`.
- Runtime effect is applied by POST `/api/core/tunables/reconfigure`, which restarts `login` and `sysctl`.
- Some tunables may still require reboot or subsystem restart beyond `sysctl` reload depending on kernel behavior; the provider must document this risk.
- `default_value` and `type` are volatile API/model fields and should be omitted from managed resource configuration, or exposed only after live read semantics are verified.

## Safety Risks

- Incorrect tunables can break networking, firewall behavior, kernel behavior, or remote management access.
- Validation is largely delegated to OPNsense `TunableField`; provider-side allow-lists would be incomplete without target-version data.
- Runtime state may differ from persisted config until reconfigure succeeds, until a subsystem reloads, or after reboot-only tunables take effect.
- `reset` is broad and destructive relative to individual Terraform resources; it should not be used for normal resource deletion.

## Target-Version Availability

Current OPNsense published docs and source expose `core/tunables`, aligning with the provider's current target series. This is enough to move tunables/sysctl out of upstream-blocked for current upstream source/docs, but target-version availability across the provider's supported minimum range remains unverified. Live validation is still required before implementation to confirm behavior on the provider's supported target appliance and minimum supported version range.

## Recommendation

- Classify system tunables/sysctl as **Coming with safety/target-version live-validation gate**.
- Implement a future UUID item resource only after live validation confirms create/read/update/delete/reconfigure behavior and safe import semantics.
- Manage only persistent fields (`tunable`, `value`, `descr`) unless volatile fields are exposed as computed-only attributes.
- Do not model runtime-only sysctl state as desired state.
- Do not use the `reset` endpoint in normal resource lifecycle.
- Add follow-up work for a tunables/sysctl resource implementation with explicit safety documentation and live validation requirements.
