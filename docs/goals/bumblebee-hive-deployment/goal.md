# Bumblebee Hive Deployment

## Objective

Design and implement the first deployable Bumblebee self-service Windows deployment path plus a companion Cloudflare Worker receiver, tentatively named Bumblebee Hive, so a developer machine can install Bumblebee, enroll securely, run a scheduled baseline scan, and deliver verified inventory data to operator-owned storage.

## Original Request

Come up with a deployment plan, research the best solution, and figure out an endpoint to receive Bumblebee transport data as another open source application alongside Bumblebee.

## Intake Summary

- Input shape: `existing_plan`
- Audience: Windows developers installing Bumblebee and operators receiving inventory data.
- Authority: `approved`
- Proof type: `test`
- Completion proof: A real Windows Bumblebee run uploads a gzip NDJSON batch through Cloudflare Access plus Bumblebee HMAC to a local/dev Hive Worker, and the receiver records immutable raw batch data plus a run index with a complete `scan_summary`.
- Goal oracle: End-to-end deployment smoke: installer/bootstrapper configures Bumblebee without manual flag work, `bumblebee selftest` passes, scheduled-task wrapper can run, Worker verifies Access/HMAC/gzip/raw body semantics, and R2/D1-compatible storage receives a complete run.
- Likely misfire: Building an attractive receiver or installer draft that does not preserve Bumblebee's transport contract, requires developers to hand-configure flags/secrets, or adds Cloudflare-specific logic to Bumblebee instead of a small generic compatibility-layer extension.
- Blind spots considered: Cloudflare Access header support, per-device HMAC key lifecycle, Cloudflare request-size limits, durable accept semantics, Windows secret storage, installer checksum/signature verification, no-dashboard v1 scope, and maintaining upstream-friendly Bumblebee changes.
- Existing plan facts: Use Cloudflare Workers behind Zero Trust, protect ingest with Access plus HMAC, store raw batches in R2, store device/run metadata in D1, use Queues for async normalization, add only generic env-backed HTTP headers to Bumblebee, and make the first receiver scope raw archive plus run index rather than a full dashboard.

## Goal Oracle

The oracle for this goal is:

`A Windows Bumblebee deployment smoke proves a configured install can send a complete gzip NDJSON scan through Cloudflare Access plus Bumblebee HMAC into Bumblebee Hive, and Hive durably records the raw batch and run index before returning 2xx.`

The PM must keep comparing task receipts to this oracle. Planning, discovery, a passing tiny slice, or a clean-looking board is not enough. The goal finishes only when a final Judge/PM audit maps receipts and verification back to this oracle and records `full_outcome_complete: true`.

## Goal Kind

`existing_plan`

## Current Tranche

Complete the first end-to-end deployment tranche: validate the plan against current source docs, implement the smallest upstream-friendly Bumblebee transport enhancement needed for Cloudflare Access service tokens, create the Hive Worker MVP for enrollment and ingest, create the Windows self-service bootstrapper/wrapper path, and verify the full local/dev smoke. Continue through successive safe Worker packages until the oracle is satisfied or a real external dependency blocks only a specific task.

## Non-Negotiable Constraints

- Keep Bumblebee a compatibility-layer fork, not a divergent product fork.
- Do not add Cloudflare-specific behavior to Bumblebee; add only generic env-backed HTTP header support if needed.
- Preserve Bumblebee's existing HTTP transport contract: raw-body HMAC, timestamp prefix, gzip-before-HMAC, and 2xx only after durable full-batch acceptance.
- Developers should not need to understand Bumblebee flags or manually configure secrets for the normal install path.
- Use reputable sources for installs and downloads, verify checksums, and prefer signing/Authenticode validation when available.
- Keep raw inventory, secrets, hostnames, SIDs, usernames, and full profile paths out of commits and receipts.
- Treat dashboard/UI as out of scope for v1 unless a later Judge explicitly finds it necessary for the oracle.

## Stop Rule

Stop only when a final audit proves the full original outcome is complete.

Do not stop after planning, discovery, or Judge selection if the user asked for working software or automation and a safe Worker task can be activated.

Do not stop after a single verified Worker package when the broader owner outcome still has safe local follow-up work. Advance the board to the next highest-leverage safe Worker package and continue unless a phase, risk, rejected-verification, ambiguity, or final-completion review is due.

Do not create one Worker/Judge pair per repeated file, table, route, or helper. Put repeated same-shape work into one Worker package and review the package as a whole.

Do not stop because a slice needs owner input, credentials, production access, destructive operations, or policy decisions. Mark that exact slice blocked with a receipt, create the smallest safe follow-up or workaround task, and continue all local, non-destructive work that can still move the goal toward the full outcome.

## Slice Sizing

Safe means bounded, explicit, verified, and reversible. It does not mean tiny.

A good task is the largest safe useful slice.

Small is not the goal. Useful is the goal.

A Worker should finish the whole assigned slice. A Judge should judge the whole assigned slice. A PM should reorient the board when tasks are safe but not moving the outcome.

Tiny tasks are allowed when the failure is isolated, the risk is high, the scope is unknown, or the tiny task unlocks a larger slice. Tiny tasks are bad when they keep happening, do not change behavior, only add wrappers/contracts/proof files, or avoid the real milestone.

Do not stop because a slice needs owner input, credentials, production access, destructive operations, or policy decisions. Mark that exact slice blocked with a receipt, create the smallest safe follow-up or workaround task, and continue all local, non-destructive work that can still move the goal toward the full outcome.

## Canonical Board

Machine truth lives at:

`docs/goals/bumblebee-hive-deployment/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status, active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/bumblebee-hive-deployment/goal.md.
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
10. If a problem, suggestion, or follow-up should become a repo artifact, create an approved issue/PR or ask the operator whether to create one.
11. Review at phase, risk, rejected-verification, ambiguity, or final-completion boundaries; do not review every small Worker by habit.
12. Finish only with a Judge/PM audit receipt that maps receipts and verification back to the original user outcome and records `full_outcome_complete: true`.

Issue and PR handoffs are supporting artifacts. `state.yaml` remains authoritative, and every external artifact decision must be recorded in a task receipt.
