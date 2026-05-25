# Windows Full Smoke Validation

## Objective

Smoke test the Windows compatibility layer end to end against the behavior that
has already been implemented, then record a redacted receipt in `windows.md`.

The intended outcome is proof that the current Windows fork behaves as a
compatibility layer: build/test/selftest pass, implemented root discovery paths
work, explicit scans emit schema-compatible package and `scan_summary` records,
endpoint identity fields populate as expected, sinks work, and known untested
areas remain explicitly documented rather than silently claimed.

## Oracle

The goal is complete only when:

- The current implemented Windows checklist in `windows.md` is mapped to smoke
  commands or explicitly documented as a non-goal/open gap.
- The local Windows smoke suite runs against the current branch using the
  verified local Go toolchain and any reputable installed prerequisites needed
  for critical paths.
- Redacted evidence proves:
  - `go test`, race tests, build, and `bumblebee.exe selftest` pass,
  - baseline roots preview and real baseline scan complete,
  - explicit project/deep fixture scans emit stable path and record fields,
  - package and `scan_summary` endpoint fields match and include env-provided
    `device_id`,
  - file and HTTP sinks emit package records plus complete summaries,
  - implemented browser/editor/MCP root families are exercised where present or
    covered by controlled fixture roots,
  - all-user expansion is exercised through a controlled temporary users root,
  - raw inventory stays outside the repo.
- `windows.md` gets a concise redacted smoke receipt and keeps current known
  gaps visible.
- No schema, README, scanner behavior, root discovery behavior, release config,
  or user-facing support expansion is changed.
- The branch is committed and pushed to `origin/windows/compat-layer`.

## Constraints

- Keep this as validation of existing behavior, not new feature work.
- Do not commit raw NDJSON, SIDs, hostnames, usernames, tokens, or profile paths.
- Use `%TEMP%` or another non-repo evidence directory for raw smoke artifacts.
- If a critical smoke path is blocked only by a missing reputable prerequisite,
  install or use the reputable prerequisite rather than skipping that path.
- If a smoke fails because behavior is missing, record the failure and stop
  before changing product behavior.
- Do not edit `README.md`, schemas, parser logic, scanner semantics, sinks,
  root-discovery behavior, CI, GoReleaser config, or deployment docs.

## Existing Plan Facts

- The known local Go executable is:
  `C:\Users\bbutner\AppData\Local\Programs\bumblebee-tools\go1.26.3\go\bin\go.exe`.
- Existing `scripts\windows-smoke.ps1` already covers build, selftest, baseline
  roots, baseline file scan, local HTTP sink, redacted summary output, and
  raw evidence outside the repo.
- The existing script does not fully cover endpoint `device_id` propagation,
  explicit `deep` fixture scans, controlled `--all-users` expansion, or every
  completed root family through fixture roots.
- `windows.md` intentionally still leaves some items open, including broader
  username-shape validation, non-implemented root ecosystems, remaining browser
  families, redirected-known-folder behavior, WSL, native Windows ecosystems,
  deployment docs, user-facing docs, and maintenance tasks.
- The smoke receipt should update `windows.md` only after verification passes.

## Likely Misfire

The dangerous failure mode is treating a green partial smoke as proof of all
Windows support. This tranche should prove implemented compatibility-layer
behavior and preserve every unimplemented or untested boundary.
