# Windows Walker Safety Hardening

## Objective

Implement and verify the next Goal 6 Windows compatibility-layer slice: harden
walker behavior around Windows junctions, reparse points, and ACL-denied paths
before adding more Windows roots.

The intended outcome is a bounded compatibility-layer change that proves broad
or deep Windows scans do not loop through junctions/reparse points and that
unreadable paths are surfaced through diagnostics without failing otherwise
healthy scans.

## Oracle

The goal is complete only when:

- A reputable Go toolchain is available and `go version` is recorded before
  implementation verification.
- Focused tests prove Windows junction/reparse directories are not descended
  into, including a loop-style case.
- Focused tests prove normal directories and intended metadata remain
  scannable.
- Focused scanner tests prove ACL-denied paths produce structured diagnostics
  without failing a healthy scan, or the board records a precise environment
  blocker if Windows cannot enforce the denial in this workspace.
- `windows.md` checks only the Goal 6 items directly proven by the tranche and
  leaves unrelated Goal 6, root, browser, WSL, endpoint, deployment, and native
  ecosystem items open.
- Verification passes:
  - `go test ./internal/walk ./internal/scanner`
  - `go test ./cmd/bumblebee ./internal/...`
  - `powershell -ExecutionPolicy Bypass -File scripts\windows-smoke.ps1`
  - `git diff --check -- internal/walk internal/scanner windows.md`

## Constraints

- Keep Windows support as an upstream-friendly compatibility layer, not a
  divergent scanner fork.
- Do not change public schema, emitted record fields, parser output, profile
  names, root kinds, ecosystem names, sink behavior, exposure matching, or CLI
  flags.
- Do not add new package roots, browser roots, Windows-native ecosystems,
  all-users behavior, WSL behavior, endpoint identity changes, deployment docs,
  or package-manager execution in this tranche.
- Do not check raw NDJSON inventory, sensitive local path inventories, or smoke
  raw output into the repo.
- Install or locate prerequisites from reputable sources rather than skipping
  critical verification because the local PATH is missing `go`.

## Existing Plan Facts

- Current shell did not find `go`; implementation should first locate an
  existing Go toolchain or install one from the official Go distribution source.
- Add or verify Windows-specific traversal behavior in `internal/walk` so
  directory reparse points, including junctions, are skipped before descent.
- Treat the reparse guard as the supplement to the non-Unix `dirKey` path
  fallback; do not redesign record identity or root behavior.
- Add Windows tests for junction/reparse loop safety and normal metadata
  traversal.
- Add Windows scanner coverage for ACL-denied paths using stable local
  permissions setup, such as `icacls`, when the environment allows it.
- Update `windows.md` with receipts and check only proven Goal 6 items.
- Commit and push after verification with a message like:
  `fix: harden Windows walker reparse handling`.

## Likely Misfire

The dangerous failure mode is broadening this into root expansion, deployment
documentation, endpoint identity, or native ecosystem support. This tranche is
walker safety only: traversal hardening, diagnostics proof, tests, and a narrow
`windows.md` receipt.
