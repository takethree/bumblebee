# T001 Scout Receipt

## Scope

Read-only scout of the current Bumblebee and Hive repositories for the Hive-managed exposure catalog sync goal.

## Bumblebee Current State

- CLI dispatch lives in `cmd/bumblebee/main.go`. Current subcommands are `scan`, `roots`, `selftest`, and `version`; there is no `hive` command group yet.
- `scan` already supports explicit exposure catalogs through `--exposure-catalog`, `--max-catalog-size`, and `--findings-only`.
- Catalog loading and validation live in `internal/exposure/exposure.go`. It supports a file or non-recursive directory of `*.json` catalogs, requires matching `schema_version`, parses before use, and builds exact-match indexes.
- Exposure matching lives in `internal/scanner/scanner.go`. A non-nil catalog emits `record_type=finding` records; package output can be suppressed with `FindingsOnly`.
- HTTP transport lives in `internal/output/httpsink.go` and `cmd/bumblebee/sink.go`. It already supports HMAC signing, gzip, timestamp headers, and env-var-backed custom headers.
- Scan summary shape lives in `internal/model/model.go` and `docs/schema/v0.1.0/scan-summary.schema.json`. It has no catalog provenance fields yet.
- Exposure catalog schema lives in `docs/schema/v0.1.0/exposure-catalog.schema.json`. It already allows extra top-level and entry fields.
- Tests relevant to a Hive slice exist in `cmd/bumblebee/main_test.go`, `internal/exposure/exposure_test.go`, `internal/scanner/findings_test.go`, `internal/output/httpsink_test.go`, and `cmd/bumblebee/selftest_test.go`.
- There is no local protected Hive config/cache package, no `bumblebee hive join`, no `bumblebee hive catalog sync`, and no `bumblebee hive run`.

## Hive Current State

- Worker routing lives in `F:/bumblebee-hive/src/index.ts`.
- Public/device endpoints currently include `POST /v1/enroll` and `POST /v1/ingest`; admin/UI routes cover overview, health, attention, devices, runs, normalization jobs, findings, package views, lifecycle actions, and retention.
- Existing `POST /v1/enroll` issues `device_id`, `hmac_key`, `ingest_path`, and required transport headers.
- Existing `POST /v1/ingest` verifies Cloudflare Access service headers, verifies Bumblebee HMAC over the raw body, stores raw batches in R2, indexes batches/runs in D1, and queues normalization.
- Normalization stores package/finding records in `inventory_records`, current package state in `inventory_current`, and findings in `exposure_findings`.
- D1 schema is migration-based under `F:/bumblebee-hive/migrations`. Catalog publishing/storage tables do not exist yet.
- R2 currently stores raw inventory batches only. There is no catalog bucket binding or current catalog object path.
- Admin metadata and UI routes intentionally avoid raw inventory, `summary_json`, object keys, local paths, usernames, hostnames, SIDs, and secrets.
- Tests use in-memory D1/R2/Queue harnesses in `test/ingest.test.ts` and `test/e2e.test.ts`.
- The installer in `scripts/install-bumblebee.ps1` currently writes per-machine/per-user Bumblebee config and wrapper script for `scan --output=http`; it does not configure catalog sync or Hive subcommands.

## Risks And Decisions

- Adding catalog provenance to `scan_summary` changes the public schema. If done, it should be additive/optional and explicitly approved by the Judge.
- Normal `bumblebee scan` must remain explicit and local-only; only `bumblebee hive run` should sync from Hive.
- Bumblebee needs a cross-platform config/cache abstraction, not a Windows-only wrapper script.
- Hive should publish validated catalog bundles without exposing raw source internals, secrets, or tenant-specific behavior.
- First implementation should avoid production deploy/migration until local tests prove the shape.

## Recommended First Worker Slice

Implement a local/test-backed vertical slice:

1. Add Hive catalog bundle endpoints/storage to Hive:
   - Admin publish endpoint accepts a validated exposure catalog bundle.
   - Device catalog endpoint serves the current manifest plus catalog files.
   - Migration stores current catalog metadata.
2. Add Bumblebee Hive support:
   - `bumblebee hive join` writes local config and protected/permissioned secrets.
   - `bumblebee hive catalog sync` fetches and validates Hive catalog files into a last-known-good cache.
   - `bumblebee hive run` syncs first, falls back to last-known-good on sync failure, fails with no valid cache, runs `scan` semantics with the cached catalog, and uploads to Hive.
3. Add tests proving publish, sync, cached fallback, missing-cache failure, and findings ingestion.

The Judge should bound this further into exact files and verification commands before Worker edits begin.
