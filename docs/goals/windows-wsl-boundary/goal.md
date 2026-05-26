# Windows Goal 11 WSL Boundary

## Objective

Define and verify the Windows compatibility-layer policy for WSL behavior.

## Original Request

`$goalbuddy:goal-prep` for the approved next target after choosing Goal 11:
document that WSL package state belongs to the Linux Bumblebee binary inside
each distro, while preserving explicit roots as generic operator-supplied
paths.

## Intake Summary

- Input shape: `existing_plan`
- Audience: maintainer of the TakeThree Windows compatibility fork
- Authority: `approved`
- Proof type: `test`
- Completion proof: docs and tests prove Windows does not auto-discover or
  claim WSL filesystem/package coverage, while existing explicit-root behavior
  remains generic and unchanged.
- Goal oracle: Goal 11 is complete when `windows.md`, public/operator docs,
  and Windows smoke/test coverage all state and prove that WSL is not a
  Windows baseline discovery target; WSL package state should be inventoried by
  running the Linux Bumblebee binary inside the distro.
- Likely misfire: adding WSL discovery, rejecting explicit roots unnecessarily,
  scanning `\\wsl$`/`\\wsl.localhost` by default, or claiming Windows baseline
  roots cover Linux distro package state.
- Blind spots considered: explicit `--root` semantics, WSL UNC paths, distro
  `rootfs` paths under AppData packages, docs already mentioning WSL, and
  preserving the compatibility-layer boundary.
- Existing plan facts: chosen policy is "Generic Only"; do not auto-discover
  WSL; do not add WSL root rejection; explicit WSL-looking `--root` paths remain
  generic explicit roots if Windows can read them, but are not documented as
  supported WSL inventory.

## Goal Oracle

The oracle for this goal is:

`Goal 11 is complete when the Windows compatibility layer has a documented and
tested WSL boundary: Windows default roots and smoke summaries show zero WSL
roots, docs direct WSL users to run the Linux Bumblebee binary inside each
distro, and no WSL discovery, rejection, schema, parser, profile, root-kind, or
CLI semantics were added.`

The PM must keep comparing task receipts to this oracle. A docs edit alone is
not enough if tests or smoke can still imply WSL coverage. A test alone is not
enough if user-facing docs still say WSL behavior is undefined.

## Goal Kind

`existing_plan`

## Current Tranche

Complete one coherent compatibility-layer package:

- Validate current WSL wording in `windows.md`, README, inventory docs, and
  Windows deployment docs.
- Document the final Goal 11 policy and close only the checkboxes that match
  the chosen "Generic Only" behavior.
- Add a regression guard proving Windows default roots do not include WSL UNC
  or distro-rootfs paths.
- Extend Windows smoke receipt coverage with a redacted `wsl_root_count=0`
  check.

## Non-Negotiable Constraints

- Keep Windows as a compatibility layer, not a WSL scanner.
- Do not add `wsl.exe` calls, registry reads, WSL APIs, distro enumeration,
  `\\wsl$` discovery, `\\wsl.localhost` discovery, distro `rootfs` discovery,
  new root kinds, schema changes, parser changes, sink changes, or profile
  semantic changes.
- Do not reject explicit `--root` paths solely because they look WSL-related;
  explicit roots remain generic operator-supplied paths.
- Do not claim WSL inventory from Windows baseline, project, deep, or
  `--all-users` default behavior.
- Preserve user and prior-work edits; do not revert unrelated changes.

## Stop Rule

Stop only when a final audit proves the full Goal 11 outcome is complete.

Do not stop after planning, discovery, or a docs-only update if a safe Worker
task can add the required tests and smoke evidence. If Scout finds existing
docs or tests contradict the selected policy, record the contradiction and let
Judge choose the smallest safe correction.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-wsl-boundary/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-wsl-boundary/goal.md.
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
