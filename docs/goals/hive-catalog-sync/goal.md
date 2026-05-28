# Hive Catalog Sync

## Objective

Build first-class, cross-platform Bumblebee Hive support so enrolled Bumblebee
instances can join a Hive, sync the latest Hive-managed exposure catalog, and run
with last-known-good catalog fallback.

## Original Request

`$goalbuddy:goal-prep` after approving the cross-repo plan for Hive-managed
`threat_intel` catalog distribution and Bumblebee Hive sync/run commands.

## Intake Summary

- Input shape: `existing_plan`
- Audience: Bumblebee operators, developers running scheduled endpoint scans,
  and maintainers of the Bumblebee Windows compatibility fork.
- Authority: `approved`
- Proof type: `test`
- Completion proof: Bumblebee and Hive tests prove join, catalog sync, cached
  fallback, Hive run, catalog metadata, server-side catalog publishing, and
  findings ingestion without weakening existing scanner semantics.
- Goal oracle: A local or test-backed Hive publishes a validated catalog bundle;
  Bumblebee joins that Hive, syncs the bundle, runs against the cached catalog,
  emits catalog metadata in the summary, uploads to Hive, and Hive shows the
  resulting finding.
- Likely misfire: Making this Windows-only, silently auto-enabling bundled
  `threat_intel`, making normal scans depend on the network, or turning the
  Windows compatibility layer into a divergent threat-intel platform.
- Existing plan facts:
  - Hive should auto-sync upstream Bumblebee `threat_intel`.
  - Bumblebee support must be cross-platform, not only a Windows wrapper.
  - Add `bumblebee hive join`, `bumblebee hive catalog sync`, and
    `bumblebee hive run`.
  - Use last-known-good catalog cache when Hive sync fails.
  - If no valid cache exists, Hive run should fail rather than silently scanning
    without findings.
  - Store local Hive config and protected secrets on Windows, macOS, and Linux.
  - Keep normal `bumblebee scan` behavior explicit and unchanged.

## Goal Oracle

The oracle for this goal is:

`A verified cross-repo integration proves Hive can publish a validated current
catalog and Bumblebee can join, sync/cache it, run with it, upload results, and
show findings in Hive while preserving normal scan semantics.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, a
passing helper package, or a Hive-only/Bumblebee-only partial slice is not enough.
The goal finishes only when a final audit maps receipts and verification back to
this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Validate and implement the first useful cross-repo vertical slice for
Hive-managed catalog sync. The first slice should be large enough to prove the
real integration path, but it may stage deploy/push as explicit follow-up tasks
if local implementation is verified first.

## Non-Negotiable Constraints

- Keep Bumblebee's normal `scan` behavior unchanged unless a later Judge task
  approves a narrow schema addition.
- Do not silently use embedded or repo-shipped `threat_intel` by default.
- Keep Hive generic and open source; do not add Take3-specific behavior.
- Do not print or commit Hive Access secrets, enrollment tokens, HMAC keys,
  device IDs, local usernames, hostnames, SIDs, or full local profile paths.
- Use source-backed catalog validation and preserve last-known-good behavior.
- Keep Windows compatibility-layer changes upstream-friendly and avoid
  platform-specific behavior where shared Bumblebee code is appropriate.

## Stop Rule

Stop only when a final audit proves the full cross-repo outcome is complete or
when a specific task is blocked by missing credentials, production approval, or a
decision outside the board's authority. Blocked subtasks should not stop safe
local work.

## Canonical Board

Machine truth lives at:

`docs/goals/hive-catalog-sync/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/hive-catalog-sync/goal.md.
```
