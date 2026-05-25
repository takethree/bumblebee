# Windows Goal 6A Redirected Documents Compatibility

## Objective

Implement the next Windows compatibility-layer tranche for redirected known
folders by supporting the current user's real Windows `Documents` location for
PowerShell module baseline roots.

## Original Request

`$goalbuddy:goal-prep` for the approved next target after choosing Goal 6A:
redirected `Documents` compatibility.

## Intake Summary

- Input shape: `existing_plan`
- Audience: maintainer of the Windows compatibility fork
- Authority: `approved`
- Proof type: `test`
- Completion proof: source validation, focused tests, full Go tests, Windows
  smoke, docs updates, and final audit prove current-user redirected
  `Documents` works without expanding Windows into registry/SID/domain/Azure AD
  discovery.
- Goal oracle: Windows baseline includes the current user's real `Documents`
  known-folder path for PowerShell module roots while preserving the existing
  compatibility-layer boundaries.
- Likely misfire: adding broad OneDrive crawling, registry profile discovery,
  all-users redirected-folder inference, or a new Windows product model instead
  of a narrow compatibility-layer fix.
- Blind spots considered: all-users redirected folders, registry/hive access,
  domain/Azure AD profiles, OneDrive privacy, known-folder API availability,
  duplicate roots, and hosts where `Documents` is not redirected.
- Existing plan facts: use the Windows known-folder API for the current user's
  `Documents`; keep `%USERPROFILE%\Documents` fallback; de-duplicate roots;
  keep all-users conservative; update docs and smoke evidence.

## Goal Oracle

The oracle for this goal is:

`Goal 6A is complete when Windows baseline roots include the current user's resolved Documents known-folder path for PowerShell module roots, tests prove redirected/fallback/de-dupe/all-users behavior, smoke records redacted current-host evidence, docs state the exact support boundary, and final audit confirms no registry/SID/domain/Azure AD/OneDrive crawling was introduced.`

The PM must keep comparing task receipts to this oracle. Planning, source
research, or a passing small test is not enough. The goal finishes only when a
final Judge/PM audit maps receipts and verification back to this oracle and
records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete one coherent compatibility-layer package:

- Source-validate the Windows known-folder API and PowerShell redirected
  `Documents` behavior.
- Implement current-user resolved `Documents` roots for PowerShell modules.
- Preserve fallback and all-users conservative behavior.
- Add Windows tests and smoke receipt coverage.
- Update README, deployment docs, inventory docs, and `windows.md`.

## Non-Negotiable Constraints

- Keep Windows as a compatibility layer, not a new product model.
- Do not add registry, SID, domain, Azure AD, token impersonation, WMI, browser
  API, package-manager command, or broad OneDrive crawling behavior.
- Do not change parser behavior, schema, emitted ecosystems, root-kind names,
  output shape, or profile semantics.
- Do not claim all-users redirected known-folder support.
- Preserve user and prior-work edits; do not revert unrelated changes.

## Stop Rule

Stop only when a final audit proves the full Goal 6A outcome is complete.

Do not stop after planning, discovery, or source validation if a safe Worker
task can be activated. If source validation rejects the current plan, record the
rejection and choose the next safe board task rather than implementing around
the missing proof.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-redirected-documents/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-redirected-documents/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Run the bundled GoalBuddy update checker when available and mention a newer
   version without blocking.
4. Re-check the oracle and likely misfire.
5. Work only on the active board task.
6. Write a compact task receipt.
7. Update the board.
8. Continue into the next safe task unless blocked or final audit is due.
9. Finish only with a Judge/PM audit receipt that records
   `full_outcome_complete: true`.
