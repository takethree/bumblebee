# Hive Operator Safety

## Original Request

Prepare a GoalBuddy execution board for the Hive operator-safety plan: make Hive safe for real operator use after the live smoke-test fixture produced a scary, non-production finding in the admin UI.

## Interpreted Outcome

The Bumblebee and Bumblebee Hive repositories should support a safer production workflow where:

- the Windows installer uses `bumblebee hive join` and scheduled `bumblebee hive run`;
- test smoke data is explicitly isolated from production operator views;
- operators have a safe, audited purge path for test or stale devices;
- re-running enrollment on the same workstation does not create surprise duplicate production devices;
- verification proves default Hive admin views cannot show test findings as real workstation findings.

## Input Shape

Existing plan plus recovery/safety hardening.

## Existing Plan Facts

- Update the Windows installer path in `F:\bumblebee-hive` to use the new Hive compatibility-layer commands instead of legacy direct HTTP scan config.
- Add device environment support with default `production` and explicit `test`.
- Make `bumblebee hive join` idempotent by reusing the existing local device ID unless a new device is explicitly requested.
- Default admin and UI collection views to production data only.
- Add a safe admin purge endpoint and operator script with dry-run and confirmation.
- Update tests and E2E coverage so smoke-test data is enrolled as test data and is hidden from production views.
- Commit/push/deploy/migrate only after local verification is clean and the live safety smoke proves production views stay clean.

## Constraints

- Do not print or commit Hive Access credentials, enrollment tokens, HMAC keys, raw object keys, hostnames, usernames, SIDs, or full local profile paths.
- Do not use direct D1 deletes as the long-term operator workflow; implement a guarded product path.
- Keep Hive organization-neutral and do not bake Take3-specific behavior into the open source app.
- Preserve Bumblebee's compatibility-layer boundary: normal `bumblebee scan` must remain explicit and local-only; Hive catalog sync belongs to `bumblebee hive run`.
- Avoid overclaiming production readiness until installer, test isolation, purge, and live verification all pass.

## Likely Misfire

The goal could appear successful by merely hiding test rows in the UI while stale test data still exists, or by cleaning the current database manually without adding durable product safeguards. The completion proof must show both durable code paths and verified operator behavior.

## Goal Oracle

Local and live verification show:

- Hive default admin views contain only production data.
- Test-device fixture findings appear only when explicitly filtered to test data.
- The purge endpoint/script dry-run and real purge remove a test device and all related D1/R2 data.
- The Windows installer installs and schedules `bumblebee hive run`.
- Re-running join on the same workstation preserves the device identity by default.
- Both repositories' tests/builds pass.

## Completion Criteria

The tranche is complete when a final Judge/PM audit maps the implemented changes to the oracle above, confirms all required tests pass, confirms live Hive deploy/migration safety checks pass, and records that no required Worker task remains queued or active.
