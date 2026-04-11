# AGENTS

## Startup
Read in this order:
1. `README.md`
2. `docs/architecture.md`
3. `docs/api-spec.md`
4. `docs/constraints.md`
5. `docs/observability.md`

Public API: yes.
This repo owns `Manifest`, `Image`, `Release`, and `Intent`.
If ownership, compatibility, or boundary questions appear, go back to `../devflow-control/docs/system/boundaries.md` and `../devflow-control/docs/policies/api-compatibility.md`.

## Commands
- `bash scripts/regen-swagger.sh`
- `go test ./...`
- `bash scripts/build.sh`

## When to go back to devflow-control
Go back when the task changes release/build ownership, manifest semantics, API compatibility expectations, or control-layer summaries.
