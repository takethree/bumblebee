# Windows Minimum Support Documentation

## Objective

Document the minimum Windows support policy in `windows.md` as the next Goal 1
compatibility-layer slice.

The intended outcome is a docs-only change that separates the Go runtime floor,
the CI validation target, release artifact architecture scope, and the operator
support claim without broadening README or public support claims.

## Oracle

The goal is complete only when:

- Current authoritative sources are checked for Go Windows minimum
  requirements, GitHub Actions Windows runner image policy, and Microsoft
  Windows lifecycle status.
- `windows.md` checks Goal 1's "Document the minimum supported Windows
  versions" item only after the support policy exists.
- `windows.md` records:
  - runtime floor,
  - validated CI target,
  - operator support claim,
  - Windows 10 support caveat,
  - `amd64` / `arm64` artifact and validation scope.
- The known-limitations/current-support-boundary section no longer says the
  platform boundary is unknown.
- Public user-facing docs such as `README.md` remain untouched.
- Verification passes:
  - `git diff --check -- windows.md`
  - targeted searches for the support-policy terms in `windows.md`

## Constraints

- Docs-only implementation.
- Do not edit code, CI, GoReleaser config, README, deployment docs, or scanner
  behavior in this tranche.
- Do not imply full Windows support beyond the compatibility layer that has
  been tested.
- Prefer source-backed wording over memory; verify lifecycle/toolchain facts
  during the `/goal` run before editing.
- Preserve the TakeThree self-maintained fork model; upstream PRs are not part
  of this tranche.

## Existing Plan Facts

- `go.mod` uses Go 1.25 and CI setup-go requests Go 1.25.
- CI validates `windows-latest`.
- GoReleaser builds Windows artifacts for `amd64` and `arm64`.
- The proposed support stance is:
  - Go-compatible Windows runtime floor: Windows 10 / Windows Server 2016 or
    newer, if confirmed by current Go minimum requirements.
  - CI validation target: GitHub Actions `windows-latest`.
  - Operator claim: actively Microsoft-serviced Windows versions only.
  - Windows 10 should be caveated because general support ended on
    2025-10-14, except still-serviced LTSC/ESU paths.
  - Windows `arm64` artifacts are built but not runtime-smoke-validated yet.
- Update only `windows.md` during implementation.
- Commit and push after verification with:
  `docs: document Windows support floor`.

## Likely Misfire

The dangerous failure mode is turning a narrow support-floor note into broad
marketing or README claims. This tranche should make the platform boundary more
precise while keeping unsupported or unverified Windows cases visible.
