# Windows Native Ecosystem Scope

## Objective

Finalize the Goal 10 Windows-native ecosystem support decisions in `windows.md`,
then implement the first approved compatibility-layer slice: NuGet
project/deep inventory from file metadata.

## Original Request

"$goalbuddy:goal-prep" after the plan to update `windows.md` with what is
supported and what to start with.

## Intake Summary

- Input shape: `existing_plan`
- Audience: the maintainer of the TakeThree Windows compatibility fork.
- Authority: `approved`
- Proof type: `test`
- Completion proof: `windows.md` records the Goal 10 support decisions, NuGet
  project/deep parsing is implemented, focused tests and `go test` pass, and a
  final audit confirms the work remains a compatibility layer.
- Goal oracle: Goal 10 in `windows.md` accurately says what is supported,
  deferred, and next; `bumblebee.exe` can inventory NuGet project metadata
  without package-manager execution or schema divergence.
- Likely misfire: treating Windows support as broad installed-app inventory
  or adding native Windows command/API discovery instead of file-based
  metadata parsing.
- Blind spots considered: NuGet global cache output volume, PowerShell `.psd1`
  parsing risk, ecosystem naming for PowerShell modules, and maintenance risk
  from shared schema/scanner drift.
- Existing plan facts: NuGet is first; PowerShell modules are second; Chocolatey,
  Scoop, winget/MSIX/AppX, and Visual Studio extensions are deferred; Cargo,
  Maven, and Gradle are cross-platform follow-ups.

## Goal Oracle

The oracle for this goal is:

`Goal 10 in windows.md is decision-complete, and NuGet project/deep scans emit schema-compatible package records from packages.config and packages.lock.json with passing tests and no package-manager execution.`

The PM must keep comparing task receipts to this oracle. Planning, discovery,
a passing tiny slice, or a clean-looking board is not enough. The goal finishes
only when a final Judge/PM audit maps receipts and verification back to this
oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Apply the approved support-scope decisions to `windows.md`, implement NuGet as
the first native Windows ecosystem parser, verify it with unit and scanner
tests, and audit that the resulting code still follows the compatibility-layer
principle.

## Non-Negotiable Constraints

- Keep this as a Windows compatibility layer, not a divergent scanner fork.
- Do not execute `nuget`, `dotnet`, PowerShell package commands, winget, Scoop,
  Chocolatey, or Visual Studio tooling.
- Do not add registry, WindowsApps, installed-app, or broad global-cache
  inventory in this tranche.
- Preserve the shared output schema, scanner model, sink behavior, and exposure
  matching.
- Prefer shared parsers and narrow scanner dispatch over Windows-only semantic
  branches.
- Keep raw smoke/inventory evidence outside the repo; only redacted receipts
  belong in tracked docs.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Judge selection if a safe Worker task
can be activated.

Do not stop after the `windows.md` update if the NuGet implementation remains
queued and locally safe to do.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.
The NuGet implementation should be one coherent parser slice rather than a
series of tiny helper-only tasks.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-native-ecosystem-scope/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-native-ecosystem-scope/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Run the bundled GoalBuddy update checker when available and mention a newer version without blocking.
4. Work only on the active board task.
5. Write a compact task receipt.
6. Update the board.
7. If safe local work remains, choose the next largest reversible Worker package and continue unless blocked.
8. Finish only with a Judge/PM audit receipt that maps receipts and verification back to the oracle and records `full_outcome_complete: true`.
