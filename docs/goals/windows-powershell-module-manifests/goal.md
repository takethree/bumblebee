# Windows PowerShell Module Manifests

## Objective

Implement PowerShell module manifest inventory as the next Windows
compatibility-layer slice, preserving Bumblebee's shared scanner model and
read-only file-based behavior.

## Original Request

"$goalbuddy:goal-prep" after the approved plan for the next step: close the
NuGet user-visible help gap, then implement PowerShell `.psd1` module manifest
support in line with `windows.md`.

## Intake Summary

- Input shape: `existing_plan`
- Audience: the TakeThree-maintained Windows compatibility fork maintainer.
- Authority: `approved`
- Proof type: `test`
- Completion proof: PowerShell module manifest records are implemented, docs
  reflect the support boundary, focused and full verification pass, and a final
  audit confirms the work remains a compatibility layer.
- Goal oracle: `windows.md` Goal 10 has the PowerShell design item closed, the
  CLI exposes accurate supported ecosystems, and `bumblebee.exe` can inventory
  `.psd1` module manifests without executing PowerShell or changing shared
  output semantics.
- Likely misfire: adding broad PowerShell/Windows package-manager discovery,
  Gallery/API lookup, `PSModulePath` command execution, or installed-app
  inventory instead of bounded manifest parsing.
- Blind spots considered: ecosystem naming, safe `.psd1` parsing without
  evaluating PowerShell, baseline root volume, redirected Documents/OneDrive
  limitations, and the existing NuGet help-text drift.
- Existing plan facts: fix the NuGet `--ecosystem` help text; emit
  `ecosystem=powershell-module`, `package_manager=powershell`, and
  `source_type=powershell-module-manifest`; derive package name from the
  manifest filename and version from top-level `ModuleVersion`; add curated
  Windows module roots; update docs and smoke tests only for behavior actually
  implemented.

## Goal Oracle

The oracle for this goal is:

`Goal 10 in windows.md records PowerShell module manifest design as complete,
and local verification proves .psd1 manifests emit schema-compatible package
records from file metadata only, with no PowerShell command execution and no
compatibility-layer boundary drift.`

The PM must keep comparing task receipts to this oracle. Planning, discovery,
a passing tiny slice, or a clean-looking board is not enough. The goal finishes
only when a final Judge/PM audit maps receipts and verification back to this
oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Validate and execute the approved PowerShell module manifest slice. The first
safe work is to confirm the exact code/docs touchpoints and then implement one
coherent parser/root/docs/smoke package. Do not split this into repeated tiny
tasks unless verification exposes a specific blocker.

## Non-Negotiable Constraints

- Keep Windows support as a compatibility layer, not a divergent scanner fork.
- Do not execute PowerShell, PowerShellGet, PSResourceGet, package-management
  commands, Gallery/API calls, registry queries, or installed-app inventory.
- Preserve the shared output schema, scanner model, sink behavior, exposure
  matching, and profile meanings.
- Use shared parser and scanner dispatch paths; isolate Windows-specific work
  to root discovery/tests/docs where possible.
- Do not claim redirected Documents, OneDrive known-folder discovery, WSL,
  Chocolatey, Scoop, winget/MSIX/AppX, Visual Studio extensions, or NuGet
  global package-cache roots.
- Keep raw smoke/inventory evidence outside the repo; only redacted aggregate
  receipts belong in tracked docs.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after plan validation if a safe Worker package can be activated.

Do not stop after parser-only implementation if docs, roots, smoke coverage,
or the NuGet help-text cleanup remain required for the oracle.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

This tranche should normally be one coherent Worker package after validation:
model/CLI cleanup, parser, scanner dispatch, Windows root discovery, docs, and
tests. If the `.psd1` parsing design proves riskier than expected, split only
at that risk boundary.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-powershell-module-manifests/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-powershell-module-manifests/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Run the bundled GoalBuddy update checker when available and mention a newer version without blocking.
4. Validate the approved plan against `windows.md`, current code, and official PowerShell docs before editing.
5. Work only on the active board task.
6. Write a compact task receipt.
7. Update the board.
8. If safe local work remains, choose the next largest reversible Worker package and continue unless blocked.
9. Finish only with a Judge/PM audit receipt that maps receipts and verification back to the oracle and records `full_outcome_complete: true`.
