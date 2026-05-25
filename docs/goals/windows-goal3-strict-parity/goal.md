# Windows Goal 3A Strict-Parity Baseline Roots

## Objective

Implement the strict-parity Windows Goal 3A tranche: add Windows baseline root
defaults only where they are reliable Windows equivalents of existing
macOS/Linux baseline behavior and already-supported file parsers.

## Original Request

`$goalbuddy:goal-prep` for the approved plan to close the next Windows
compatibility-layer parity gap after comparing Windows against macOS/Linux.

## Intake Summary

- Input shape: `existing_plan`
- Audience: maintainer of the Windows compatibility fork
- Authority: `approved`
- Proof type: `test`
- Completion proof: focused tests, full Go tests, Windows smoke, docs updates,
  and final audit prove npm/Python/pipx Windows baseline roots work without
  schema/parser divergence.
- Goal oracle: Windows baseline roots gain strict-parity npm/Python/pipx
  coverage while Ruby/Bundler and Composer global roots remain explicitly
  deferred or cross-platform follow-ups.
- Likely misfire: closing every Goal 3 checkbox by adding Windows-only roots
  that macOS/Linux do not claim, weakening the compatibility-layer model.
- Blind spots considered: custom prefixes, registry/env discovery, redirected
  folders, WSL, virtualenv discovery, global caches, and package-manager
  command execution.
- Existing plan facts: add Windows npm global, Python user site-packages, and
  pipx venv roots; keep root discovery literal and existence-filtered; update
  tests, smoke, README, inventory docs, and windows.md; do not add Ruby or
  Composer Windows-only baseline roots in this tranche.

## Goal Oracle

The oracle for this goal is:

`Goal 3A is complete when Windows baseline resolves source-backed npm/Python/pipx root candidates, tests and smoke prove records can emit from those roots, docs define the support boundary, and final audit confirms no Windows-only support divergence was introduced.`

The PM must keep comparing task receipts to this oracle. Planning, discovery,
a passing tiny slice, or a clean-looking board is not enough. The goal finishes
only when a final Judge/PM audit maps receipts and verification back to this
oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete one coherent strict-parity implementation package:

- Source-validate exact Windows npm/Python/pipx paths.
- Implement only literal baseline candidates in the Windows platform hook.
- Add current-user and `--all-users` root tests.
- Extend controlled Windows smoke for the new roots.
- Update docs and `windows.md` to show implemented and deferred Goal 3 items.

## Non-Negotiable Constraints

- Keep Windows as a compatibility layer, not a new product model.
- Do not add registry, command-output, package-manager execution, browser/API,
  WSL, redirected-folder, custom-prefix, or global-cache discovery.
- Do not change parser behavior, schema, emitted ecosystems, root-kind names,
  or output shape unless a Judge first rejects the plan.
- Defer Ruby/Bundler and Composer global roots unless the board explicitly
  converts them into a cross-platform expansion decision.
- Preserve user and prior-work edits; do not revert unrelated changes.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Judge selection if a safe Worker task
can be activated. Do not stop after a single verified work package when the
broader owner outcome still has safe local follow-up work. Advance the board to
the next highest-leverage safe Worker package and continue unless a phase, risk,
rejected-verification, ambiguity, or final-completion review is due.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

The preferred Worker package is the full Goal 3A implementation and verification
slice. Split only if source evidence invalidates one of npm/Python/pipx, if
verification fails twice, or if a candidate requires behavior outside the
non-negotiable constraints.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-goal3-strict-parity/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-goal3-strict-parity/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Run the bundled GoalBuddy update checker when available and mention a newer
   version without blocking.
4. Re-check the strict-parity oracle and likely misfire.
5. Work only on the active board task.
6. Assign Scout, Judge, Worker, or PM according to the task.
7. Write a compact task receipt.
8. Update the board.
9. If safe local work remains, choose the next largest reversible Worker
   package and continue unless blocked.
10. Finish only with a Judge/PM audit receipt that maps receipts and
    verification back to the original user outcome and records
    `full_outcome_complete: true`.
