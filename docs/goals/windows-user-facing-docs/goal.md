# Windows User-Facing Docs

## Objective

Implement Goal 9 from `windows.md`: update public/user-facing documentation so it accurately describes the Windows compatibility layer without overclaiming Windows support.

## Original Request

The user asked for the next plan after Goal 8, accepted Goal 9 as the next part, and invoked `$goalbuddy:goal-prep`.

## Intake Summary

- Input shape: `existing_plan`
- Audience: Bumblebee users, operators, and maintainers reading README and inventory-source docs
- Authority: `approved`
- Proof type: `artifact`
- Completion proof: `README.md`, `docs/inventory-sources.md`, and `windows.md` are updated; Goal 9 checklist items are closed only when covered; doc verification passes.
- Goal oracle: the Goal 9 checklist in `windows.md` plus a final diff/check receipt proving docs say "Windows compatibility layer" and preserve unsupported boundaries.
- Likely misfire: changing public docs to imply full Windows parity or closing unsupported Windows root/ecosystem/WSL gaps.
- Blind spots considered: README support language, Windows quick-start examples, Windows-specific source mapping, unsupported native ecosystems, WSL, remaining browser families, unimplemented package roots, and troubleshooting language.
- Existing plan facts: update `README.md`; update `docs/inventory-sources.md`; add Windows quick-start/root examples; add MCP, browser, and editor-extension Windows notes; clearly label unsupported Windows ecosystems; add Windows troubleshooting; update `windows.md`; do not edit code, schemas, CI, release config, or generated evidence.

## Goal Oracle

The oracle for this goal is:

`Goal 9 in windows.md is accurately closed by user-facing docs that describe the tested Windows compatibility layer, include practical Windows examples, and keep unimplemented Windows gaps explicitly unsupported.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, a passing tiny slice, or a clean-looking board is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete the user-facing documentation slice only. This tranche is docs-only: update README and inventory-source docs, update the Goal 9 tracking section in `windows.md`, verify the docs, and stop after final audit. Code, schemas, CI, release configuration, smoke scripts, generated evidence, and native ecosystem implementation remain out of scope.

## Non-Negotiable Constraints

- Use "Windows compatibility layer" wording instead of broad "full Windows support" unless the sentence is explicitly bounded.
- Do not claim support for unimplemented roots, unsupported browser families, Windows-native ecosystems, WSL, redirected known folders, or untested account shapes.
- Do not close Goals 3, 4, 6, 7, 10, 11, or 13.
- Do not edit code, tests, schemas, CI, release config, generated evidence, or deployment docs unless a final Judge explicitly determines the Goal 9 plan is impossible without a tiny docs-only adjustment.
- Preserve Goal 8 as the deployment guide; Goal 9 should point to it rather than duplicating all operator deployment details.
- Use placeholders only; do not include real hostnames, tokens, SIDs, usernames, tenant IDs, or raw NDJSON evidence.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Worker selection if a safe Worker task can be activated.

Do not mark the goal complete until the user-facing docs cover every Goal 9 checkbox, verification passes, and a final audit confirms the compatibility-layer guardrails were followed.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

This goal should be one coherent Worker slice unless validation reveals a concrete risk: README, inventory-source mapping, and `windows.md` tracking should be updated together so the public wording and checklist stay consistent.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-user-facing-docs/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-user-facing-docs/goal.md.
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
