# Windows NuGet Global Cache Decision

## Objective

Revisit the deferred NuGet global package-cache baseline-root item and close it
as an evidence-backed Windows compatibility-layer decision without adding cache
inventory accidentally.

## Original Request

The user asked for a plan for the next `windows.md` target after upstream-sync
maintenance, then invoked `$goalbuddy:goal-prep` for that plan.

## Intake Summary

- Input shape: `existing_plan`
- Audience: maintainer of the TakeThree Windows compatibility fork
- Authority: `approved`
- Proof type: `source_backed_answer` plus `test`
- Completion proof: official NuGet docs, aggregate-only local cache volume
  evidence, current parser/dispatch validation, docs updates, tests, and final
  audit prove whether NuGet global cache baseline roots should be added,
  deferred, or planned as a separate parser slice.
- Goal oracle: the tranche is complete when `windows.md` Goal 10 records an
  evidence-backed decision for NuGet global package-cache baseline roots,
  user-facing docs do not overclaim cache inventory, no raw package names from
  local caches are written to tracked docs, and verification passes.
- Likely misfire: adding `%USERPROFILE%\.nuget\packages` as a baseline root even
  though the current NuGet scanner reads only project metadata, or expanding
  Bumblebee into installed/cache-state inventory without a separate design.
- Blind spots considered: NuGet cache path can be overridden, cache contents may
  be stale or manually cleaned, `.nuspec` or `.nupkg` parsing would be a new
  source type, and local cache volume can be high.
- Existing plan facts: the recommended default is to close the current checklist
  item as "revisited and intentionally deferred" after source validation, not to
  implement a root or parser in this tranche.

## Goal Oracle

The oracle for this goal is:

`Goal 10 NuGet global-cache decision is complete when official NuGet cache
behavior is verified, local volume is measured only as aggregate counts, current
Bumblebee NuGet support is confirmed as project/deep metadata only, docs record
the decision without overclaiming global-cache inventory, and tests/diff checks
pass.`

The PM must keep comparing receipts to this oracle. A docs-only decision is
acceptable only if it is source-backed and explicitly explains why adding a
baseline root without a cache parser would be misleading.

## Goal Kind

`existing_plan`

## Current Tranche

Complete one compatibility-layer decision package:

- Source-validate NuGet global packages/cache behavior and override paths.
- Measure local cache volume with aggregate counts only.
- Confirm the current Bumblebee NuGet parser/dispatch only reads
  `packages.config` and `packages.lock.json`.
- Decide whether to add a baseline root now, defer it by design, or spawn a
  separate cache-parser implementation plan.
- Update `windows.md` and, if needed, `docs/inventory-sources.md` so support
  claims match the decision.

## Non-Negotiable Constraints

- Do not add NuGet global cache baseline roots in this tranche unless a Judge
  first proves the current parser can emit meaningful records from them.
- Do not add `.nuspec`, `.nupkg`, `project.assets.json`, registry/config
  discovery, command execution, or Visual Studio API support in this tranche.
- Do not write local NuGet package names, versions, paths, or raw cache listings
  into tracked docs or GoalBuddy receipts.
- Keep the Windows work as a compatibility layer; do not change shared schema,
  sink behavior, exposure matching, or profile semantics.
- Preserve user and prior-work edits; do not revert unrelated changes.

## Stop Rule

Stop only when a final audit proves the current decision tranche is complete, or
when Scout/Judge finds that a safe decision cannot be made without missing
official source evidence.

Do not stop after research if a bounded docs/test Worker can close the current
decision cleanly.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-nuget-global-cache-decision/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-nuget-global-cache-decision/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Re-check the oracle and likely misfire.
4. Work only on the active board task.
5. Write a compact task receipt.
6. Update the board.
7. Continue into the next safe task unless blocked or final audit is due.
8. Finish only with a Judge/PM audit receipt that records
   `full_outcome_complete: true`.
