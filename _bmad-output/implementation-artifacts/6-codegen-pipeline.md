# Epic 6: Code Generation Pipeline

Status: done (6-1, 6-2, 6-3)

## Outcome
A `text/template`-based generator (`internal/generate`) produces the repetitive
Terraform resource four-file boilerplate from declarative YAML schemas, wired via
`//go:generate go run ./internal/generate` (root main.go) and `go generate ./...`.

## Stories
- **6-1 YAML schema format** — documented in `internal/generate/README.md`; schemas in `internal/generate/schemas/*.yaml`. Fields support bool/int/string/selectmap/selectmaplist/csvset; item and singleton kinds.
- **6-2 Generation pipeline** — `internal/generate/{main,templates}.go`: parses YAML, computes exact grouped imports per file, renders model/schema/resource, runs `go/format`. Output is `*_model.gen.go` / `*_schema.gen.go` / `*_resource.gen.go` with a `DO NOT EDIT` header (auto-excluded by golangci). **Idempotent** (re-run yields byte-identical files → CI `go generate && git diff --exit-code` safe).
- **6-3 Generate resources** — generated 11 OSPF/OSPFv3 item resources (completing Epic 19's 19-4/19-5). The pipeline is reusable to accelerate Wave B.

## Notes
- Decision history: originally bypassed, then **"keep & build"** chosen; sequenced after Wave A so the hand-written quagga/openvpn resources established the stable pattern the templates encode.
- Generated resources reuse the package helpers `setToCSV`/`sliceToSet`/`intOrEmpty`/`intOrZero` and the Epic 13 singleton client.
- Acceptance tests are not generated — validate field/monad/endpoint accuracy against a live box ([[acceptance-testing-hardware-box]]).
- `make check` fully green with generator linted and generated files excluded.

## File List
- internal/generate/main.go, templates.go, README.md
- internal/generate/schemas/quagga_ospf.yaml, quagga_ospf6.yaml
- internal/service/quagga/ospf_*.gen.go, ospf6_*.gen.go (33 files)
- main.go (go:generate directive), go.mod/go.sum (yaml.v3 direct)
