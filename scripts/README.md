# Scripts

## verify.sh

Preferred repo-local verification entrypoint.
Runs `go test ./...` and then `bash scripts/build.sh` so Swagger regeneration and the Linux build happen in one repeatable handoff path.

## build.sh

Regenerates Swagger and builds the Linux amd64 service binary at `bin/devflow-release-service`.
Use this when you need the packaged service artifact without running the full repo verify flow.

## regen-swagger.sh

Regenerates `docs/generated/swagger/*` from handler annotations under `cmd/` and `pkg/api/`.
Run this whenever the public API contract changes.

## create_github_repo.sh

Creates a GitHub repository through the GitHub REST API.
Requires `GITHUB_TOKEN` and is only for workspace bootstrap or repo-splitting operations.

## create_service_repos.sh

Calls `create_github_repo.sh` for the standard DevFlow service-repo set.
Use it only when bootstrapping the historical split-repo layout.

## export_service_repo.sh

Thin wrapper around `../devflow-control/scripts/export_service_repo.sh`.
Use it to export one service repo from the workspace with the control-repo logic.

## export_service_repos.sh

Exports the standard DevFlow service repos into a target directory, defaulting to `/tmp/devflow-split`.
This is a coordination helper, not part of normal service verification.
