# Windows Fork Hygiene Boundary Pass

## Goal

Before adding Windows baseline roots, perform a narrow compatibility-layer hygiene pass. The current Windows work should be easy to review, easy to rebase, and ready for Goal 3 without changing scan semantics or adding new inventory coverage.

## Original Request

The user asked what should come next while carefully following `windows.md`, then approved the plan to do Goal 13 fork hygiene before starting Goal 3 baseline roots.

## Outcome

The current Windows-specific root/home handling is audited and, if justified by the audit, refactored behind small OS-specific helpers. `windows.md` Goal 13 is updated only for evidence-backed items. No Windows baseline roots, browser roots, all-users support, WSL behavior, deployment docs, endpoint identity policy, native ecosystem support, schema changes, parser changes, sink changes, or CLI semantic changes are added.

## Oracle

The goal is complete only when:

- a read-only audit maps the current Windows diff to shared vs platform-specific surfaces;
- a Judge confirms the exact allowed refactor, or confirms no refactor is needed;
- any Worker changes are behavior-neutral and limited to root/home helper isolation plus tests/checklist updates;
- `windows.md` Goal 13 is checked only where current evidence supports it;
- `go test ./...`, `go test -race ./...`, Windows build/selftest, and `git diff --check` pass;
- final audit confirms no schema, parser, sink, exposure, CLI, root-kind, profile, baseline-root, browser-root, all-users, WSL, deployment, endpoint-identity, or native ecosystem scope leaked into this tranche.

## Non-Goals

- Do not add Windows baseline roots.
- Do not add browser roots.
- Do not implement Windows `--all-users`.
- Do not define WSL behavior.
- Do not add Windows deployment docs.
- Do not decide Windows endpoint identity policy.
- Do not add native Windows ecosystem inventory.
- Do not change public schemas, parser behavior, sink behavior, exposure matching, CLI flags, root kinds, or profile meanings.

## Likely Misfire

The dangerous failure mode is treating “fork hygiene” as permission to begin Goal 3 discovery work. This tranche is about keeping the compatibility boundary clean; it must not expand inventory coverage.

## Starter Command

`/goal Follow docs/goals/windows-fork-hygiene/goal.md.`
