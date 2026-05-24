# Windows Walker Privacy Hardening

## Original Request

Prepare GoalBuddy for the next Windows compatibility-layer slice: Goal 6A,
privacy-first Windows walker hardening from `windows.md`.

## Interpreted Outcome

The next `/goal` run should implement and verify a bounded Windows walker
hardening tranche that keeps broad Windows scans away from sensitive and
high-cost profile data while preserving Bumblebee's original read-only,
metadata-only scanner design.

## Goal Oracle

The tranche is complete when tests and the Windows smoke harness prove:

- Windows-sensitive and high-cost directory excludes prevent traversal of
  browser profile data, credential stores, cloud-sync folders, and known cache
  trees during broad scans.
- Required metadata paths still work: baseline browser extension roots,
  editor extension roots, MCP config roots, and normal project package metadata
  remain scannable.
- Sensitive browser profile files are not considered or opened by scanner
  dispatch during explicit broad-root scans.
- `windows.md` marks only the Goal 6 items proven by this tranche and leaves
  junction/reparse-point, `dirKey`, ACL-denied-path, all-users, WSL, and
  Windows-native ecosystem work unchecked.
- Verification passes with focused tests, full Go tests, race tests, the
  redacted Windows smoke harness, and `git diff --check`.

## Non-Negotiable Constraints

- Keep Windows support as an upstream-friendly compatibility layer, not a
  divergent scanner fork.
- Do not change public schema, record fields, parser output, profile names,
  root kinds, ecosystem names, sink behavior, exposure matching, or CLI flags.
- Do not add new browser roots, Windows-native ecosystems, all-users behavior,
  WSL behavior, endpoint identity changes, deployment docs, or package-manager
  execution in this tranche.
- Do not implement junction/reparse-point identity, `dirKey` redesign, or
  Windows ACL-denied test behavior in this tranche.
- Do not check raw NDJSON inventory, browser inventory details, or sensitive
  local path inventories into the repo.

## Existing Plan Facts

- This is Goal 6A, not all of Goal 6.
- The plan is privacy-first: harden excludes and scanner-side sensitive-file
  skipping before adding more browser families or deeper Windows behavior.
- Add Windows-specific excludes for sensitive/high-cost browser, credential,
  cache, and cloud-sync paths.
- Add narrow OneDrive handling for `OneDrive` and `OneDrive - <tenant>` style
  directories without broad globbing.
- Add scanner-side sensitive browser filename skips so explicit broad roots do
  not count or dispatch cookies, history, login, storage, or session files.
- Preserve current Firefox extension discovery through profile
  `extensions.json` and Chromium extension discovery through per-profile
  `Extensions/<id>/<version>/manifest.json`.
- Use the verified Go toolchain from `%TEMP%` if Go is still not on PATH, and
  use `F:\msys64\ucrt64\bin` for race-test GCC if needed.

## Likely Misfire

The dangerous failure mode is turning privacy hardening into a scanner fork or
declaring all of Goal 6 complete. This tranche should not redesign Windows
junction behavior, expand inventory scope, or weaken deep scans by blocking the
known metadata files Bumblebee is designed to read.

## Enough For This Tranche

Windows broad-scan privacy hardening is implemented, tested, smoke-verified,
and documented in `windows.md`. The remaining Goal 6 items stay open unless
the implementation directly proves them.
