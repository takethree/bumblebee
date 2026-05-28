# T003 Worker Receipt

## Result

Implemented the first local/test-backed Hive catalog sync vertical slice across Bumblebee and Hive.

## Bumblebee Changes

- Added `bumblebee hive` command group:
  - `bumblebee hive join`
  - `bumblebee hive catalog sync`
  - `bumblebee hive run`
- Added `internal/hive` for:
  - cross-platform config/cache directory defaults;
  - restricted-permission local config and secret files;
  - Hive enrollment client;
  - current catalog fetch;
  - file-hash verification;
  - exposure catalog validation before cache promotion;
  - last-known-good catalog cache loading.
- `hive run` now:
  - syncs catalog first;
  - falls back to last-known-good cache if sync fails;
  - fails when no valid cache exists;
  - invokes existing scan semantics with `--exposure-catalog` pointed at the cache;
  - uploads through the existing HTTP sink with HMAC and Access headers from environment-backed values;
  - emits optional `scan_summary.catalog` metadata for Hive-managed runs only.
- Normal `bumblebee scan` remains explicit and does not fetch Hive data.

## Hive Changes

- Added migration `0009_catalog_releases.sql`.
- Added `POST /v1/admin/catalog/current` behind existing Access plus admin-token auth.
- Added `GET /v1/admin/catalog/current` for metadata/admin retrieval.
- Added `GET /v1/catalog/current` behind existing Access service auth and active-device validation.
- Hive validates each published file as a Bumblebee exposure catalog before promotion.
- Current catalog responses include a manifest, per-file SHA-256 hashes, entry counts, and inline file content for client cache sync.
- Disabled or unknown devices cannot fetch the current catalog.

## Docs

- Updated Bumblebee README with the Hive command flow and cache/fallback semantics.
- Updated Bumblebee transport docs with optional Hive catalog summary metadata.
- Updated Hive README with catalog publishing and device fetch behavior.

## Verification

- `go test -count=1 ./cmd/bumblebee ./internal/hive ./internal/exposure ./internal/scanner ./internal/output ./internal/model` passed.
- `go test -count=1 ./...` passed.
- `go run ./cmd/bumblebee hive help` passed.
- `npm test` in `F:/bumblebee-hive` passed: 3 test files, 52 tests.
- `npm run build` in `F:/bumblebee-hive` passed.
- `git diff --check` passed in both repos, with only existing CRLF working-tree warnings.

## Remaining Gap

The dedicated cross-repo integration proof is still pending as T004. That proof should publish a Hive catalog, run Bumblebee through `bumblebee hive run`, normalize the uploaded batch, and verify the resulting finding is visible from Hive.
