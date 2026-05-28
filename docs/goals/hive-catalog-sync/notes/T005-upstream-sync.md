# T005 Worker Receipt

## Result

Implemented configurable Hive upstream catalog sync from Bumblebee `threat_intel`.

## Changes

- Added Hive environment config:
  - `CATALOG_UPSTREAM_SYNC_ENABLED`
  - `CATALOG_UPSTREAM_CONTENTS_URL`
  - `CATALOG_UPSTREAM_SOURCE`
  - `CATALOG_UPSTREAM_FILE_LIMIT`
- Added `POST /v1/admin/catalog/sync-upstream` behind existing Access plus admin-token auth.
- Added scheduled upstream sync when `CATALOG_UPSTREAM_SYNC_ENABLED` is true.
- Upstream sync reads a GitHub repository contents style directory listing, fetches JSON files from their `download_url`, validates them through the same exposure catalog validation path, and promotes them through the same current catalog release path.
- Default upstream contents URL points at the public `perplexityai/bumblebee/threat_intel` directory. Operators can override it.
- Updated Hive README with upstream sync endpoint and environment variables.

## Verification

- `npm test` passed in `F:/bumblebee-hive`: 3 test files, 55 tests.
- `npm run build` passed in `F:/bumblebee-hive`.
- `BUMBLEBEE_E2E=1 BUMBLEBEE_REPO=F:/bumblebee GO_EXE=%LOCALAPPDATA%/CodexTools/go1.26.3/go/bin/go.exe npm run test:e2e` passed: 1 test file, 3 tests.
- `go test -count=1 ./...` passed in `F:/bumblebee`.
- `git diff --check` passed in both repositories, with only CRLF working-tree warnings.

## Scope Guard

- No production deploy, remote D1 migration, remote secret mutation, or real credentials were used.
- No Take3-specific behavior was added.
- Scheduled upstream sync is opt-in through environment configuration.
