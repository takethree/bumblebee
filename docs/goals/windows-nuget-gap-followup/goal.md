# Windows NuGet Gap Follow-Up

## Objective

Close the useful gaps found during the full NuGet validation pass while keeping
the Windows work inside the compatibility-layer boundary.

## Original Request

"$goalbuddy:goal-prep" after the plan for the NuGet data gaps.

## Intake Summary

- Input shape: `existing_plan`
- Audience: the maintainer of the TakeThree Windows compatibility fork.
- Authority: `approved`
- Proof type: `test`
- Completion proof: NuGet lockfile requested ranges are emitted through the
  existing schema, the Windows smoke script proves NuGet data shape with a
  redacted summary, intentional non-gaps are documented, and the full
  verification set passes.
- Goal oracle: a full local validation run proves the NuGet gap follow-up
  works end to end without package-manager execution, registry/API discovery,
  `project.assets.json`, or NuGet global-cache baseline roots.
- Likely misfire: broadening the scanner into installed-app/cache inventory or
  creating new schema fields when the existing `requested_spec` field is
  enough.
- Blind spots considered: PowerShell 5.1 argument handling, paths with spaces
  in smoke fixtures, duplicate source-accurate NuGet records, and docs that
  still describe `requested_spec` as MCP-only.
- Existing plan facts: add NuGet `requested` range support via
  `requested_spec`, add persisted NuGet smoke coverage, fix smoke argument
  quoting, document intentional non-gaps, and do not expand scope beyond file
  metadata.

## Goal Oracle

The oracle for this goal is:

`NuGet requested ranges appear in requested_spec, scripts/windows-smoke.ps1 proves NuGet package data from a path with spaces, docs describe the intentional boundaries, and vet/tests/race/smoke pass.`

The PM must keep comparing task receipts to this oracle. Planning, discovery,
a passing tiny slice, or a clean-looking board is not enough. The goal finishes
only when a final Judge/PM audit maps receipts and verification back to this
oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Implement the small NuGet follow-up slice identified by full validation: enrich
lockfile output with requested ranges, make the smoke script prove NuGet data
shape, and document which observed gaps are intentional compatibility-layer
boundaries.

## Non-Negotiable Constraints

- Do not execute NuGet, dotnet, PowerShell package-management, Chocolatey,
  Scoop, winget, or Visual Studio tooling.
- Do not add registry/API discovery, `obj/project.assets.json`, WindowsApps, or
  NuGet global package-cache baseline roots.
- Do not add a schema version bump or new public field unless the existing
  `requested_spec` field is proven insufficient.
- Keep raw NDJSON, hostnames, usernames, SIDs, paths, and tokens out of tracked
  receipts; only redacted summaries may be committed.
- Keep the live GoalBuddy board disabled for this goal.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after adding `requested_spec` if the smoke script and docs remain
unverified.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. This should be one
coherent Worker slice unless implementation evidence reveals a risk boundary.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-nuget-gap-followup/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-nuget-gap-followup/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Work only on the active board task.
4. Write a compact task receipt.
5. Update the board.
6. Continue to the next safe task until the oracle is satisfied.
7. Finish only with a Judge/PM audit receipt that records `full_outcome_complete: true`.
