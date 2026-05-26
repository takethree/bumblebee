# Windows Goal 7 Endpoint Username Validation

## Objective

Validate and document Windows `endpoint.username` behavior without changing the
endpoint schema or inventing account-provider coverage.

## Original Request

`$goalbuddy:goal-prep` for the next Windows compatibility-layer target after
Goal 11: plan Goal 7 endpoint username validation using the non-invasive
approach.

## Intake Summary

- Input shape: `existing_plan`
- Audience: maintainer of the TakeThree Windows compatibility fork
- Authority: `approved`
- Proof type: `test`
- Completion proof: source validation, redacted smoke evidence, docs updates,
  focused tests, full tests, Windows smoke, and final audit prove the supported
  username behavior without overclaiming local/domain/Azure AD live coverage.
- Goal oracle: Goal 7 username validation is complete when Windows smoke records
  redacted evidence that `endpoint.username` is present, shape-classified, and
  consistent across emitted records, docs say it is scanner-process identity
  with account-provider-dependent shape, and any untested local/Azure AD/domain
  provider gaps remain explicit.
- Likely misfire: creating accounts, reading registry/Entra/domain state,
  normalizing username shapes, treating username as stable endpoint identity,
  writing raw usernames/SIDs/hostnames into tracked docs, or closing live
  provider validation without evidence.
- Blind spots considered: domain-style output on the current host,
  local-account and Azure AD host availability, service-account/SYSTEM runs,
  `scan_summary` vs package endpoint consistency, and redaction requirements.
- Existing plan facts: chosen validation depth is `Non-invasive`; do not create
  temporary Windows users; preserve endpoint JSON shape; keep `device_id` as the
  preferred stable endpoint key; split or leave explicit provider-specific
  follow-up evidence if this host cannot prove all account-provider shapes.

## Goal Oracle

The oracle for this goal is:

`Goal 7 username validation is complete when endpoint username behavior is
validated as scanner-process identity, Windows smoke emits only redacted
username-shape evidence and proves consistency across records, docs preserve the
device_id/username boundary, and windows.md accurately distinguishes verified
behavior from untested account-provider host evidence.`

The PM must keep comparing receipts to this oracle. A raw username in docs is a
failure. Closing all local/domain/Azure AD live validation without actual host
evidence is a failure. Adding Windows-only identity discovery is a failure.

## Goal Kind

`existing_plan`

## Current Tranche

Complete one coherent compatibility-layer package:

- Validate the current endpoint identity implementation and docs.
- Add redacted username-shape and endpoint-consistency smoke receipt coverage.
- Update Windows/docs wording so `endpoint.username` is scanner-process
  identity and not a stable endpoint key.
- Update `windows.md` Goal 7 to close only the evidence-backed behavior, and
  keep unavailable account-provider host checks explicit.

## Non-Negotiable Constraints

- Do not change the endpoint JSON schema.
- Do not normalize, parse, transform, or derive `endpoint.username`.
- Do not create temporary Windows users or require admin account creation.
- Do not read registry, Entra, Intune, domain APIs, SMBIOS, MachineGuid, or
  hardware identifiers.
- Do not write raw usernames, SIDs, hostnames, tokens, HTTP payloads, or full
  profile paths into tracked docs.
- Preserve user and prior-work edits; do not revert unrelated changes.

## Stop Rule

Stop only when a final audit proves the full current tranche is complete.

Do not stop after planning, discovery, or a docs-only update if a safe Worker
task can add the required smoke and test evidence. If Scout finds a mismatch
between current docs and implementation, record it and let Judge choose the
bounded correction.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-endpoint-username-validation/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-endpoint-username-validation/goal.md.
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
