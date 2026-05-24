# Windows Core Compatibility

## Objective

Implement the minimum Windows support bar as an upstream-friendly compatibility layer: `bumblebee.exe` builds on Windows, passes `selftest`, scans explicit Windows roots, and emits schema-compatible NDJSON with a complete `scan_summary`.

## Original Request

User asked to prepare a GoalBuddy board for the core implementation work for Windows support, based on the existing compatibility-layer plan.

## Intake Summary

- Input shape: `existing_plan`
- Audience: Bumblebee maintainers and operators who need a Windows-compatible scanner without diverging from upstream.
- Authority: `requested`
- Proof type: `test`
- Completion proof: Windows CI/build/selftest and explicit-root scan tests pass, and `windows.md` reflects the completed core checkboxes only.
- Goal oracle: A Windows-compatible `bumblebee.exe` can scan explicit `C:\...` roots read-only and emit package/finding/summary NDJSON using the existing schema.
- Likely misfire: Adding broad Windows baseline/all-users/ecosystem work instead of keeping this tranche to the core compatibility layer.
- Blind spots considered: CI script portability, GoReleaser Windows archive shape, Windows home-root classification, path separator assertions, diagnostic behavior, and avoiding schema/parser divergence.
- Existing plan facts: Preserve the proposed plan: Windows CI/release, explicit-root behavior/tests, narrow broad-home detection, schema compatibility, and `windows.md` checkbox updates. Exclude baseline defaults, all-users, WSL, deployment docs, and Windows-native ecosystems.

## Goal Oracle

The oracle for this goal is:

`go test ./...` and `go test -race ./...` pass, Windows CI includes build/selftest coverage, `bumblebee.exe` can scan explicit Windows roots and emit schema-compatible NDJSON ending in `scan_summary.status=complete`, and `windows.md` marks only the verified core items complete.

The PM must keep comparing task receipts to this oracle. Planning, discovery, a passing tiny slice, or a clean-looking board is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete the core Windows compatibility layer only. The largest safe useful implementation package is expected to cover CI/release edits, explicit Windows root tests, narrow path/home-root fixes if required, and `windows.md` tracking updates.

## Non-Negotiable Constraints

- Keep Windows support as a compatibility layer, not a divergent scanner fork.
- Do not add Windows baseline defaults in this tranche.
- Do not add Windows `--all-users` behavior in this tranche.
- Do not add WSL behavior in this tranche.
- Do not add Windows-native ecosystems such as NuGet or PowerShell modules in this tranche.
- Do not change public schema fields, ecosystem names, profile names, root kinds, or CLI flags.
- Keep shared parsers, emitters, sinks, exposure matching, and scanner behavior shared unless a Windows build/test failure proves a narrow fix is required.
- Update `windows.md` only for verified core progress.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Judge selection if a safe Worker task can be activated.

Do not mark the tranche complete until implementation evidence proves the core oracle.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

For this tranche, prefer one coherent Worker package over many tiny file-by-file tasks if Judge confirms the allowed files and verification commands are sufficient.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-core-compatibility/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-core-compatibility/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Validate the existing plan before edits.
4. Work only on the active board task.
5. Keep implementation bounded to the compatibility-layer core.
6. Write compact task receipts with commands and changed files.
7. Continue until final audit proves the oracle.
