# T001 Scout Baseline

## Decision

Proceed with the operator-safety plan. The current repos and live Hive state match the risk model in the goal: the deployed state is clean, but durable product safeguards are missing.

## Current Repo State

`F:/bumblebee` is dirty on `windows/compat-layer`:

- Modified: `README.md`, `cmd/bumblebee/main.go`, `docs/schema/v0.1.0/scan-summary.schema.json`, `docs/transport.md`, `internal/model/model.go`
- Untracked: `cmd/bumblebee/hive.go`, `docs/goals/hive-catalog-sync/`, `docs/goals/hive-operator-safety/`, `internal/hive/`

`F:/bumblebee-hive` is dirty on `main`:

- Modified: `README.md`, `src/index.ts`, `test/e2e.test.ts`, `test/ingest.test.ts`
- Untracked: `migrations/0009_catalog_releases.sql`

## Implementation Surfaces

- Installer: `F:/bumblebee-hive/scripts/install-bumblebee.ps1` still writes legacy `secrets.clixml`, legacy config shape, and a scheduled wrapper that calls `scan --output http`.
- Verification: `F:/bumblebee-hive/scripts/verify-bumblebee-pilot.ps1` still expects legacy `hive_base_url` config and `secrets.clixml`.
- Hive enrollment: `F:/bumblebee-hive/src/index.ts` `POST /v1/enroll` accepts only optional `device_id`, creates a new HMAC key, and inserts into `devices` without environment metadata.
- Hive admin views: device, run, finding, package, health, and attention queries do not have an environment dimension.
- Hive cleanup: retention exists for age-based cleanup, and lifecycle disable/enable exists, but there is no operator-safe device purge endpoint.
- Bumblebee Hive CLI: `F:/bumblebee/cmd/bumblebee/hive.go` has `hive join`, `hive catalog sync`, and `hive run`; `hive join` currently enrolls every time and does not explicitly preserve local device identity.

## Live Hive Baseline

Redacted live API checks against `https://hive.take3tech.dev`:

- Current local device prefix: `23ec24e6-639`
- Total Hive devices: `1`
- Active current device: yes
- Total findings: `0`
- `node-ipc` fixture findings: `0`
- Current baseline runs returned: `3`
- Latest current run status: `complete`

## Risks And Corrections

- The plan should avoid using direct D1 deletes except for emergency recovery. Durable cleanup belongs behind a guarded admin endpoint and script.
- Test data isolation must be query-enforced, not only visual. Default admin API and UI views should filter to `production`.
- E2E/live smoke should enroll fixture devices as `test` and prove default production findings remain clean.
- Installer migration should be tackled before live rollout, because the current repo script still installs the old direct-scan path.
