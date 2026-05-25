# Windows Deployment Documentation

## Objective

Implement Goal 8 from `windows.md`: add Windows deployment documentation for the Bumblebee Windows compatibility layer and update the local Windows checklist with a receipt.

## Original Request

The user approved the plan for Goal 8 and then invoked `$goalbuddy:goal-prep`.

## Intake Summary

- Input shape: `existing_plan`
- Audience: Windows operators maintaining the Bumblebee compatibility-layer fork
- Authority: `approved`
- Proof type: `artifact`
- Completion proof: `docs/deployment-windows.md` exists, covers every Goal 8 deployment topic, `windows.md` marks only the covered Goal 8 boxes complete, and doc-only verification passes.
- Goal oracle: the Goal 8 checklist in `windows.md` plus a final diff/check receipt proving the new doc stays compatibility-layer scoped.
- Likely misfire: writing a broad Windows support announcement or README update instead of bounded operator deployment guidance.
- Blind spots considered: unsupported Windows ecosystems, WSL, redirected folders, untested account shapes, secret handling, current-user versus `--all-users` permissions, and `diagnostics_count` wording.
- Existing plan facts: create `docs/deployment-windows.md`; document Task Scheduler, Intune/RMM/SCCM assumptions, one-shot incident response, recurring baseline runs, file/log-shipper output, HTTPS env-var secrets, permissions, cadence, verification, and `diagnostics_count`; update `windows.md`; do not touch README or code.

## Goal Oracle

The oracle for this goal is:

`Goal 8 in windows.md is accurately closed by docs/deployment-windows.md, with a receipt and doc-only checks proving the guide is complete, bounded, and consistent with the Windows compatibility-layer support boundary.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, a passing tiny slice, or a clean-looking board is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete the Windows deployment documentation slice only. This tranche is docs-only: write the Windows operator guide, update the Goal 8 tracking section in `windows.md`, verify the docs, and stop after final audit. Public docs, README, code, schemas, CI, and release configuration are out of scope.

## Non-Negotiable Constraints

- Treat Windows work as a compatibility layer, not a divergent product fork.
- Do not broaden support claims beyond behavior already implemented and smoke-tested.
- Do not update README or Goal 9 user-facing docs in this tranche.
- Do not edit code, tests, schemas, CI, release config, or generated evidence.
- Do not commit raw NDJSON, SIDs, usernames, hostnames, tokens, local full paths, or HTTP payloads.
- Use placeholder values for URLs, tokens, tenant IDs, and asset identifiers.
- Preserve unrelated open checkboxes in `windows.md`.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Worker selection if a safe Worker task can be activated.

Do not mark the goal complete until `docs/deployment-windows.md` exists, `windows.md` has a Goal 8 receipt, doc-only verification is recorded, and the final audit confirms the compatibility-layer guardrails were followed.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

This goal should be one coherent Worker slice unless verification reveals a concrete gap: create the deployment guide and update the tracking file together, because the checklist cannot be honestly closed without the guide.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-deployment-documentation/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-deployment-documentation/goal.md.
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
10. Review at final completion before marking the goal done.
