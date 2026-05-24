# Windows Record-ID Drift

## Goal

Finish the last unchecked core Windows compatibility risk by proving and, if needed, fixing `record_id` drift caused only by path separator differences. Keep the change as an upstream-friendly compatibility-layer slice: record identity may normalize path separators internally, but emitted JSON paths, schemas, CLI flags, root kinds, parsers, sinks, and profile semantics must remain unchanged.

## Original Request

The user asked what to focus on next after `windows.md`, then approved a plan to handle record-ID drift carefully and right the first time.

## Outcome

`windows.md` can mark the Goal 2 record-ID drift item complete because tests prove separator-equivalent paths hash to the same IDs. Any already-proven Goal 12 smoke-test boxes may be checked, while broader Windows baseline, browser, MCP, HTTP, WSL, multi-user, deployment, and native ecosystem work remains unchecked.

## Oracle

The goal is complete only when:

- model-level tests prove package, finding, diagnostic, and scan_summary IDs are stable across `/` and `\` path spellings;
- tests also prove genuinely different path content still changes IDs;
- emitted JSON paths remain platform-native and unchanged;
- `docs/state-model.md` documents identity-only path normalization;
- `windows.md` is updated only for verified checklist items;
- `go test ./...`, `go test -race ./...`, Windows build/selftest, and `git diff --check` pass.

## Non-Goals

- Do not add Windows baseline roots.
- Do not add browser roots.
- Do not implement all-users, WSL, deployment docs, or Windows-native ecosystems.
- Do not case-fold record identity paths in this tranche.
- Do not change public schema files, emitted field names, parser behavior, root kinds, profile names, or CLI flags.

## Likely Misfire

The dangerous failure mode is “fixing” record IDs by rewriting emitted `source_file`, `project_path`, diagnostic paths, or summary roots into slash-normalized output. The intended behavior is identity-only canonicalization, not output normalization.

## Starter Command

`/goal Follow docs/goals/windows-record-id-drift/goal.md.`
