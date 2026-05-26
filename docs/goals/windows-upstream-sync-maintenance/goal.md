# Windows Upstream Sync Maintenance

## Objective

Verify that the TakeThree Windows compatibility branch can still track
upstream Bumblebee cleanly, and update maintenance receipts without turning the
Windows fork into a divergent scanner.

## Original Request

The user asked for a plan for the next `windows.md` target, then invoked
`$goalbuddy:goal-prep` for that plan.

## Intake Summary

- Input shape: `existing_plan`
- Audience: maintainer of the TakeThree Windows compatibility fork
- Authority: `approved`
- Proof type: `test`
- Completion proof: branch/remotes are verified, upstream sync or no-op sync is
  recorded, any conflicts are classified by maintenance risk, required tests and
  Windows smoke pass, and `windows.md` Goal 13 records only evidence-backed
  maintenance progress.
- Goal oracle: the tranche is complete when `windows/compat-layer` is known to
  be current with `upstream/main` or has been safely rebased onto it, the fork
  boundary is audited for shared-code drift, and any push to
  `origin/windows/compat-layer` happens only after verification.
- Likely misfire: treating an upstream sync as permission to casually resolve
  conflicts in shared parser/schema/output/exposure/scanner semantics, or
  marking upstream PR work complete even though the user is maintaining this
  fork privately.
- Blind spots considered: upstream may have no new commits, the rebase may be a
  no-op, shared-core conflicts may signal a compatibility-layer leak, and a
  force push must use `--force-with-lease` only after tests pass.

## Goal Oracle

The oracle for this goal is:

`Goal 13 upstream-sync maintenance is complete when the branch model is verified,
upstream has been fetched, local main mirrors upstream/main, windows/compat-layer
is either rebased onto main or proven already current, required tests and Windows
smoke pass, any shared-code conflicts are treated as maintenance warnings, and
windows.md records only the evidence-backed result.`

The PM must keep comparing receipts to this oracle. A no-op sync is acceptable
only if ancestry checks prove upstream has not moved. A conflict in shared
schema, parser, output sink, exposure matching, scanner semantics, profile
meaning, or root-kind taxonomy must not be resolved casually.

## Goal Kind

`existing_plan`

## Current Tranche

Complete one maintenance package:

- Verify branch and remote state.
- Fetch upstream and origin.
- Fast-forward local `main` to `upstream/main`.
- Rebase `windows/compat-layer` onto local `main` when needed, or record a
  proven no-op when already current.
- Run verification before any push.
- Push with `--force-with-lease` only if the Windows branch history changed.
- Update `windows.md` Goal 13 with a compact maintenance receipt.

## Non-Negotiable Constraints

- Do not commit Windows compatibility work directly to `main`.
- Do not open upstream PRs to Perplexity as part of this tranche.
- Do not add new Windows inventory coverage.
- Do not change public schemas, parser behavior, output sinks, exposure
  matching, CLI semantics, root kinds, profile meanings, or emitted record
  semantics as part of routine conflict resolution.
- Treat shared-core conflicts as a warning requiring a focused follow-up plan.
- Preserve user and prior-work edits; do not revert unrelated changes.

## Stop Rule

Stop only when a final audit proves the maintenance tranche is complete, or when
a shared-core conflict or unsafe branch state is recorded as a specific blocker
with next steps.

Do not stop after fetching or planning if a safe sync, verification, receipt,
and push/no-push decision remain.

## Canonical Board

Machine truth lives at:

`docs/goals/windows-upstream-sync-maintenance/state.yaml`

If this charter and `state.yaml` disagree, `state.yaml` wins for task status,
active task, receipts, verification freshness, and completion truth.

## Run Command

```text
/goal Follow docs/goals/windows-upstream-sync-maintenance/goal.md.
```

## PM Loop

On every `/goal` continuation:

1. Read this charter.
2. Read `state.yaml`.
3. Re-check the oracle and likely misfire.
4. Work only on the active board task.
5. Write a compact task receipt.
6. Update the board.
7. Continue into the next safe task unless blocked or final audit is due.
8. Finish only with a Judge/PM audit receipt that records
   `full_outcome_complete: true`.
