# T002 Judge Plan

## Decision

Approved with sequence adjustment.

The original board had the installer Worker before the underlying Hive/Bumblebee contracts were updated. That order would force the installer either to guess at future flags or to keep the legacy path. The safe sequence is:

1. Add Hive production/test environment support and default production filtering.
2. Add Bumblebee join idempotency plus environment/new-device controls.
3. Update the installer and verifier to consume those commands.
4. Add guarded device purge endpoint/script.
5. Run full local verification, then live deploy/smoke, then commit/push.

## First Worker

T004 is the first implementation task.

## T004 Allowed Files

- `F:/bumblebee-hive/src/index.ts`
- `F:/bumblebee-hive/public/admin/index.html`
- `F:/bumblebee-hive/public/admin/app.js`
- `F:/bumblebee-hive/migrations`
- `F:/bumblebee-hive/test/ingest.test.ts`
- `F:/bumblebee-hive/test/admin-url-state.test.ts`
- `F:/bumblebee-hive/test/e2e.test.ts`
- `F:/bumblebee-hive/README.md`

## T004 Verification

- `cd F:/bumblebee-hive; npm test`
- `cd F:/bumblebee-hive; npm run build`

## Stop Conditions

- Existing data cannot be migrated to default `production` safely.
- Query filtering requires a schema redesign beyond additive device metadata.
- Test isolation cannot be enforced at API query level.

## Rationale

Device environment is the core safety boundary. Once Hive can represent and filter production/test devices, Bumblebee and installer changes can target a stable API contract. Purge depends on the same environment and device relationships but can follow as a separate Worker because it has higher destructive-operation risk.
