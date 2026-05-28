# T002 Judge Receipt

## Decision

Approve a bounded first Worker slice.

## Rationale

The largest safe first slice is a local/test-backed vertical path that proves Hive can publish a current exposure catalog bundle and Bumblebee can sync and run against it. This should not deploy, migrate remote D1, change normal `bumblebee scan` behavior, or require production credentials.

The slice should use Hive-specific commands only. Normal `scan` remains explicit and local-only.

## Exact First Worker Objective

Implement a local/test-backed Hive catalog sync vertical slice:

- Hive:
  - Add D1 migrations for catalog releases and catalog files.
  - Add `POST /v1/admin/catalog/current` behind existing Access plus admin-token auth.
  - Add `GET /v1/catalog/current` behind existing Access service auth and active-device validation.
  - Validate each published file as a Bumblebee exposure catalog before making it current.
  - Return a manifest with file hashes and inline file content for sync.
- Bumblebee:
  - Add `bumblebee hive join`, `bumblebee hive catalog sync`, and `bumblebee hive run`.
  - Store Hive config and secrets in local config/cache files with restricted file permissions where the platform supports them.
  - `hive catalog sync` fetches the current catalog bundle, verifies file hashes, parses the downloaded catalogs with existing exposure validation, and atomically promotes it as last-known-good.
  - `hive run` syncs first, falls back to last-known-good cache on sync failure, fails when no valid cache exists, then runs existing scan semantics with `--exposure-catalog` pointing at the cache and uploads to Hive with existing HTTP sink auth.
  - Add additive optional catalog metadata on `scan_summary` only for Hive runs so operators can see which catalog was used.

## Allowed Files

Bumblebee:

- `cmd/bumblebee/main.go`
- `cmd/bumblebee/main_test.go`
- `cmd/bumblebee/hive.go`
- `cmd/bumblebee/hive_test.go`
- `internal/hive/**`
- `internal/model/model.go`
- `internal/model/model_test.go`
- `docs/schema/v0.1.0/scan-summary.schema.json`
- `README.md`
- `docs/transport.md`
- `docs/goals/hive-catalog-sync/**`

Hive:

- `F:/bumblebee-hive/src/index.ts`
- `F:/bumblebee-hive/test/ingest.test.ts`
- `F:/bumblebee-hive/test/e2e.test.ts`
- `F:/bumblebee-hive/migrations/0009_catalog_releases.sql`
- `F:/bumblebee-hive/README.md`
- `F:/bumblebee-hive/docs/goals/**` only if receipts are useful

## Verification Commands

Bumblebee:

- `gofmt -w cmd/bumblebee/main.go cmd/bumblebee/hive.go cmd/bumblebee/*_test.go internal/hive/*.go internal/hive/*_test.go internal/model/model.go internal/model/model_test.go`
- `go test -count=1 ./cmd/bumblebee ./internal/hive ./internal/exposure ./internal/scanner ./internal/output ./internal/model`
- `go test -count=1 ./...`

Hive:

- `npm test`
- `npm run build`

Cross-repo:

- Local integration test proving a published Hive catalog syncs into Bumblebee cache and `bumblebee hive run` emits/uploads a finding.

## Stop If

- The implementation would print or commit raw Hive Access secrets, HMAC keys, enrollment tokens, device IDs, local usernames, hostnames, SIDs, or full local profile paths.
- The implementation requires production Cloudflare deploy, remote D1 migration, remote secret mutation, or GitHub release mutation.
- Normal `bumblebee scan` starts syncing or fetching from Hive.
- Cross-platform config/cache support would require a platform-specific privileged API beyond normal user file permissions.
- Verification fails twice for the same underlying cause.
