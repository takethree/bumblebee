# Windows Compatibility Maintenance Cleanup

## Objective

Reduce Windows support maintenance risk by keeping Windows behavior behind narrow compatibility-layer seams and removing avoidable shared core changes that make upstream updates harder.

## Original Request

"I mean code not docs, we need maintenance to be easy, changing base-code complicates life" followed by the approved maintenance cleanup plan.

## Intake Summary

- Input shape: `existing_plan`
- Audience: Bumblebee maintainers and future upstream reviewers
- Authority: `approved`
- Proof type: `test`
- Completion proof: Shared base-code drift is reduced, Windows-specific root/browser/privacy behavior is isolated behind platform hooks or build-tagged files, and verification passes.
- Goal oracle: Full test verification plus a final audit proving no required Windows behavior regressed and no new shared parser/schema/sink/profile/root-kind changes were introduced.
- Likely misfire: Treating "clean up maintenance" as permission to rewrite shared architecture, weaken Windows behavior, or make broad unrelated refactors.
- Blind spots considered: The `StableID()` path-normalization change is a core semantic change and should be removed unless separately approved as an upstream core fix. Windows browser candidate literals still sit in shared root logic and should be moved behind a hook if feasible. Test-file churn itself can create merge pain and should be reduced where it materially helps.
- Existing plan facts: Remove shared `StableID()` path normalization; move Windows browser root literals and classification to platform hooks; preserve existing Windows behavior; keep schema, parsers, sinks, endpoint identity, WSL, native ecosystems, and deployment scope unchanged.

## Goal Oracle

The oracle for this goal is:

`The Windows fork has fewer shared base-code changes while preserving current Windows behavior, proven by tests, smoke verification, and a final maintenance audit.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, or a clean-looking diff is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete one maintenance cleanup tranche: validate the cleanup plan against the current dirty diff, remove avoidable core semantic drift, move remaining Windows root-discovery literals behind platform seams, reduce test merge friction where safe, run verification, and record the result in GoalBuddy and `windows.md`.

## Non-Negotiable Constraints

- Preserve current Windows behavior from the previous verified tranches.
- Do not change output schema fields, parser behavior, sink behavior, endpoint identity, profile meanings, root-kind taxonomy, WSL behavior, deployment policy, or native ecosystem scope.
- Do not broaden Windows account discovery, browser families, package ecosystems, or deep-scan behavior.
- Do not weaken macOS/Linux behavior or tests.
- Do not revert unrelated dirty worktree changes outside this cleanup scope.
- Treat any change to shared model identity, parser, scanner, sink, or public CLI semantics as a stop condition unless it is explicitly part of this maintenance cleanup and approved by Judge.

## Stop Rule

Stop only when a final audit proves the maintenance cleanup is complete and current Windows behavior is still verified.

Do not stop after planning or Judge selection if a safe Worker task can be activated.

Do not stop after implementation if verification, `windows.md`, or the final maintenance audit is still missing.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

This cleanup should be one coherent Worker package unless Judge finds a concrete risk that requires splitting.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-maintenance-cleanup/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-maintenance-cleanup/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Run the bundled GoalBuddy update checker when available and mention a newer version without blocking.
4. Validate the plan against current files before implementation.
5. Work only on the active board task.
6. Write a compact task receipt.
7. Update the board.
8. If safe local work remains, choose the next largest reversible Worker package and continue unless blocked.
9. Finish only with a Judge/PM audit receipt that maps receipts and verification back to the original user outcome and records `full_outcome_complete: true`.
