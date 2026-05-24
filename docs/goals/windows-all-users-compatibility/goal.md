# Windows All-Users Compatibility Layer

## Objective

Implement Windows `--all-users` support as a compatibility-layer equivalent of Bumblebee's existing macOS multi-user behavior, without expanding Bumblebee's product model or scanner scope.

## Original Request

"Okay let's do it. We're trying to keep feature parity. This is essentially a Windows compatibility layer ... It shouldn't do things that Bumblebee doesn't do. Look at the Mac implementation and give me a plan for Windows."

## Intake Summary

- Input shape: `existing_plan`
- Audience: Windows fork maintainers and future upstream reviewers
- Authority: `approved`
- Proof type: `test`
- Completion proof: Windows `baseline --all-users` and `project --all-users` expand curated per-user roots across local Windows profile homes, preserve macOS guardrails, update `windows.md`, and pass targeted Go tests plus the Windows smoke script.
- Goal oracle: `go test ./cmd/bumblebee ./internal/...` and `powershell -ExecutionPolicy Bypass -File scripts/windows-smoke.ps1`, with a final audit proving no schema, parser, sink, deep-profile, explicit-root, or bare-home behavior drift.
- Likely misfire: Implementing broad Windows user discovery, registry/domain/Azure AD profile enumeration, deep home crawling, or new attribution fields instead of matching the bounded macOS `--all-users` model.
- Blind spots considered: Windows `%APPDATA%` / `%LOCALAPPDATA%` currently describe only the process user; cross-user expansion must derive per-user AppData paths from each enumerated home. Admin rights may be needed operationally but should not become a new CLI permission model.
- Existing plan facts: Keep `--all-users` invalid with `deep`; keep `--all-users` invalid with explicit `--root`; never add bare user homes for `baseline` or `project`; update CLI help and diagnostics; update `windows.md` Goal 5.

## Goal Oracle

The oracle for this goal is:

`Windows implements the same bounded all-users expansion semantics that macOS has today, verified by targeted tests, smoke output, and a final audit against windows.md Goal 5.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, a passing tiny slice, or a clean-looking board is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete the Windows `--all-users` compatibility layer as one bounded implementation tranche: validate the macOS semantics, implement the Windows equivalent behind platform helpers, add targeted tests, update operator-facing docs, run verification, and audit for maintenance risk and feature drift.

## Non-Negotiable Constraints

- Preserve Bumblebee semantics; this is a compatibility layer, not a forked scanner model.
- Do not change output schema, parser behavior, sink behavior, record attribution, ecosystem semantics, or root-kind taxonomy.
- Do not implement registry, SID, domain, Azure AD, OneDrive, or redirected-profile discovery in this tranche.
- Do not make `deep --all-users` valid.
- Do not allow `--all-users` with explicit `--root`.
- Do not add bare Windows user home directories as `baseline` or `project` roots.
- Keep Linux and unsupported OS behavior honest; do not silently claim unsupported platforms are expanded.
- Work with existing dirty-tree changes; do not revert user or prior goal work.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Judge selection if a safe Worker task can be activated.

Do not stop after implementation if verification, `windows.md`, or the maintenance-risk audit is still missing.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

This goal should avoid one helper per task. The main Worker package should implement the coherent Windows all-users slice, including tests and docs, unless Judge finds a concrete risk that requires splitting.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-all-users-compatibility/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-all-users-compatibility/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Run the bundled GoalBuddy update checker when available and mention a newer version without blocking.
4. Validate the existing plan against current repo files before implementation.
5. Work only on the active board task.
6. Write a compact task receipt.
7. Update the board.
8. If safe local work remains, choose the next largest reversible Worker package and continue unless blocked.
9. Finish only with a Judge/PM audit receipt that maps receipts and verification back to the original user outcome and records `full_outcome_complete: true`.
