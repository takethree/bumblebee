# Windows Redirected Known-Folder Boundary Closure

## Objective

Close the remaining Goal 6 redirected-known-folder gap by making the Windows compatibility-layer boundary source-backed, documented, and regression-tested without adding broad OneDrive crawling, registry discovery, token impersonation, or command execution.

## Original Request

`$goalbuddy:goal-prep` for the planned next step: address the `windows.md` Goal 6 gap for broader OneDrive and all-users redirected known folders.

## Intake Summary

- Input shape: `existing_plan`
- Audience: the Windows Bumblebee compatibility-layer fork maintainer.
- Authority: `approved`
- Proof type: `test`
- Completion proof: the Goal 6 checkbox is closed as an explicit supported/deferred boundary, docs match behavior, regression tests protect the boundary, and verification passes.
- Goal oracle: a final Judge/PM audit maps source-backed policy, current behavior, docs, tests, and verification back to the redirected-known-folder gap in `windows.md`.
- Likely misfire: implementing broad OneDrive crawling, all-users redirected `Documents` guessing, registry hive loading, token impersonation, or `PSModulePath` expansion and making the compatibility layer harder to maintain.
- Blind spots considered: current-user known-folder support already exists, all-users expansion currently varies only the home prefix, Microsoft APIs need a user token for another user's known folder, and representative redirected-Documents smoke still requires a real redirected host.
- Existing plan facts:
  - Keep current-user `Documents` known-folder support for curated PowerShell module roots.
  - Keep `--all-users` expansion based on real local profile homes and literal per-home candidates.
  - Do not infer other users' redirected `Documents` paths from OneDrive directory names.
  - Do not use registry hive loading, impersonation, per-user token acquisition, broad OneDrive crawling, or package-manager/PowerShell command execution.
  - Keep `scripts\windows-smoke.ps1 -RequireRedirectedDocuments` as the representative-host proof gate, not a requirement for this non-redirected host.

## Goal Oracle

The oracle for this goal is:

`Goal 6 records a source-backed redirected-known-folder boundary decision, regression tests prove current-user support and all-users non-discovery behavior, docs explain operator options, and verification passes without broad OneDrive or cross-user known-folder discovery.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, a passing tiny slice, or a clean-looking board is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete the redirected-known-folder boundary closure end to end: source validation, docs update, targeted regression tests, verification, and final audit.

## Non-Negotiable Constraints

- Preserve Bumblebee compatibility-layer scope.
- Do not add broad OneDrive crawling or inferred redirected-known-folder discovery.
- Do not add registry hive loading, `PSModulePath` expansion, token impersonation, per-user token acquisition, or command execution.
- Keep production changes minimal; prefer tests and docs unless source validation exposes a real code gap.
- Follow `windows.md` as the source-of-truth checklist and update it when the boundary is closed.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Judge selection if a safe Worker task can be activated.

Do not stop after a single verified Worker package when the broader owner outcome still has safe local follow-up work. Advance the board to the next highest-leverage safe Worker package and continue unless a phase, risk, rejected-verification, ambiguity, or final-completion review is due.

Do not create one Worker/Judge pair per repeated file, table, route, or helper. Put repeated same-shape work into one Worker package and review the package as a whole.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

A good task is the largest safe useful slice.

Small is not the goal. Useful is the goal.

A Worker should finish the whole assigned slice. A Judge should judge the whole assigned slice. A PM should reorient the board when tasks are safe but not moving the outcome.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-redirected-known-folders/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-redirected-known-folders/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Run the bundled GoalBuddy update checker when available and mention a newer version without blocking.
4. Re-check the intake: original request, input shape, authority, proof, blind spots, existing plan facts, and likely misfire.
5. Work only on the active board task.
6. Assign Scout, Judge, Worker, or PM according to the task.
7. Write a compact task receipt.
8. Update the board.
9. If safe local work remains, choose the next largest reversible Worker package and continue unless blocked.
10. Review at phase, risk, rejected-verification, ambiguity, or final-completion boundaries; do not review every small Worker by habit.
11. Finish only with a Judge/PM audit receipt that maps receipts and verification back to the original user outcome and records `full_outcome_complete: true`.
