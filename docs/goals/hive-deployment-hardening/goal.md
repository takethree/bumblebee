# Hive Deployment Hardening

## Objective

Harden the proven Bumblebee + Hive deployment path so the current successful Windows pilot is reproducible, locally uninstallable, and remotely revocable, without starting the second-machine pilot yet.

## Original Request

"All right come up with a plan for everything except number four. We're not ready for that yet" followed by `$goalbuddy:goal-prep`.

## Intake Summary

- Input shape: `existing_plan`
- Audience: the maintainer/operator of the TakeThree Windows compatibility fork and its companion Hive receiver.
- Authority: `approved`
- Proof type: `test`
- Completion proof: docs and code changes are implemented in the correct repos, automated tests pass, local installer uninstall behavior is verified, Hive admin revocation is tested, and a final audit confirms the second-machine pilot remains out of scope.
- Goal oracle: a receipt-backed deployment hardening tranche where the verified installer/Hive path is documented, local uninstall works, Hive can disable a device, token rotation guidance is captured, and no second-machine pilot work is performed.
- Likely misfire: treating the prior successful pilot as enough and only documenting it, or drifting into broader fleet rollout/second-machine validation before the operator is ready.
- Blind spots considered: Cloudflare Access Service Auth gotcha, public Hive repo must stay organization-neutral, local uninstall should not imply remote revocation, shared enrollment token should be treated as a rotated wave token, and cross-repo docs/code must stay aligned.
- Existing plan facts:
  - Update Hive README, Bumblebee Windows deployment docs, and `windows.md` with the verified installer/Hive path and troubleshooting.
  - Add installer `-Uninstall` that removes local state by default.
  - Add Hive admin device disable endpoint plus Wrangler D1 fallback documentation.
  - Keep enrollment as a rotated wave token for now.
  - Exclude the second-machine pilot from this tranche.

## Goal Oracle

The oracle for this goal is:

`The deployment hardening tranche is complete when the local installer lifecycle, Hive device revocation, and deployment runbooks are implemented and verified, while the final audit confirms no second-machine pilot or broader rollout was attempted.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, a passing tiny slice, or a clean-looking board is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete the approved deployment hardening work except the second-machine pilot. The run should validate the plan against the current repos, implement the largest safe coherent slices, verify each slice, deploy only when needed for a live admin-revoke smoke, and finish with an audit that the runbook, installer lifecycle, and Hive revocation behavior are all consistent.

## Non-Negotiable Constraints

- Do not perform the second-machine pilot in this goal.
- Keep Bumblebee scanner behavior generic; Hive-specific behavior belongs in Hive or docs.
- Keep the public Hive repository organization-neutral; do not add TakeThree-specific defaults or secrets.
- Do not print or commit tokens, service secrets, raw inventory, SIDs, usernames, hostnames, or full profile paths.
- Local uninstall removes local state only; remote device disable/revoke is a separate action.
- Preserve the Windows compatibility-layer boundary and maintenance posture.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Judge selection if a safe Worker task can be activated.

Do not stop after a single verified Worker package when the broader owner outcome still has safe local follow-up work. Advance the board to the next highest-leverage safe Worker package and continue unless a phase, risk, rejected-verification, ambiguity, or final-completion review is due.

Do not create one Worker/Judge pair per repeated file, route, or helper. Put repeated same-shape work into one Worker package and review the package as a whole.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

A good task is the largest safe useful slice.

Worker tasks should finish a coherent deployment-hardening slice: docs/runbooks, installer lifecycle, Hive admin revocation, or verification/final alignment.

## Canonical Board

Machine truth lives at:

`docs/goals/hive-deployment-hardening/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/hive-deployment-hardening/goal.md.
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
10. Review at phase, risk, rejected-verification, ambiguity, or final-completion boundaries.
11. Finish only with a Judge/PM audit receipt that maps receipts and verification back to the original user outcome and records `full_outcome_complete: true`.

