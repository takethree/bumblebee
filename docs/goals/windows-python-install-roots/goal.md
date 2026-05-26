# Windows Python Install Roots Compatibility Slice

## Objective

Implement the deferred Goal 3 Windows Python install-root support as a source-backed compatibility-layer baseline expansion, then verify it with unit tests and the Windows smoke suite.

## Original Request

`$goalbuddy:goal-prep` for the planned next step: source-validate and implement Windows Python install roots while staying aligned with `windows.md`.

## Intake Summary

- Input shape: `existing_plan`
- Audience: the Windows Bumblebee compatibility-layer fork maintainer.
- Authority: `approved`
- Proof type: `test`
- Completion proof: Windows baseline roots include only source-backed Python install `Lib\site-packages` roots, docs reflect the support boundary, and the focused Go tests plus full Windows smoke suite pass.
- Goal oracle: a final Judge/PM audit maps implementation receipts, source evidence, test results, and docs updates back to Goal 3 in `windows.md`.
- Likely misfire: adding broad Python environment discovery, custom prefixes, registry probing, or command execution and thereby drifting away from a maintainable Bumblebee compatibility layer.
- Blind spots considered: official-vs-observed install paths, false positives from arbitrary `C:\Python*` prefixes, Store/MSIX and Conda layouts, and smoke tests accidentally writing into real Python installs.
- Existing plan facts:
  - Keep `%APPDATA%\Python\Python*\site-packages` user-site support.
  - Add only existence-filtered CPython install-prefix `Lib\site-packages` roots when source-backed.
  - Include `%LOCALAPPDATA%\Programs\Python\Python*\Lib\site-packages` as a user package root.
  - Include `%ProgramFiles%\Python*\Lib\site-packages` and `%ProgramFiles(x86)%\Python*\Lib\site-packages` as global package roots.
  - Do not include arbitrary `C:\Python*` roots in this slice.
  - Do not execute Python, pip, uv, py launcher, registry APIs, or package-manager commands during scanning.
  - Keep arbitrary virtualenv discovery, Conda, pyenv-win, Store/MSIX internals, custom `PYTHONPATH`, and custom `PYTHONUSERBASE` deferred or explicit-root only.

## Goal Oracle

The oracle for this goal is:

`Goal 3 records the Windows Python install-root slice as implemented only after source evidence, code changes, docs updates, Go verification, and Windows smoke verification prove the supported roots populate correctly without broad Python discovery.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, a passing tiny slice, or a clean-looking board is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete the Windows Python install-root baseline slice end to end: source validation, root expansion, classification consistency, docs updates, strict-parity smoke fixture coverage, verification, and final audit.

## Non-Negotiable Constraints

- Preserve Bumblebee compatibility-layer scope; do not create new Windows-only feature behavior beyond equivalent baseline root discovery.
- Use source-backed static path candidates only.
- Do not add scan-time command execution, registry discovery, package-manager invocation, launcher enumeration, or broad filesystem crawling.
- Do not write smoke fixtures into real Python install directories.
- Keep implementation changes narrow and maintainable against upstream Bumblebee.
- Follow `windows.md` as the source-of-truth checklist and update it when the slice is implemented or a boundary is decided.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Judge selection if a safe Worker task can be activated.

Do not stop after a single verified Worker package when the broader owner outcome still has safe local follow-up work. Advance the board to the next highest-leverage safe Worker package and continue unless a phase, risk, rejected-verification, ambiguity, or final-completion review is due.

Do not create one Worker/Judge pair per repeated file, table, route, or helper. Put repeated same-shape work into one Worker package and review the package as a whole.

Do not stop because a slice needs owner input, credentials, production access, destructive operations, or policy decisions. Mark that exact slice blocked with a receipt, create the smallest safe follow-up or workaround task, and continue all local, non-destructive work that can still move the goal toward the full outcome.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

A good task is the largest safe useful slice.

Small is not the goal. Useful is the goal.

A Worker should finish the whole assigned slice. A Judge should judge the whole assigned slice. A PM should reorient the board when tasks are safe but not moving the outcome.

Tiny tasks are allowed when the failure is isolated, the risk is high, the scope is unknown, or the tiny task unlocks a larger slice. Tiny tasks are bad when they keep happening, do not change behavior, only add wrappers/contracts/proof files, or avoid the real milestone.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-python-install-roots/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-python-install-roots/goal.md.
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
