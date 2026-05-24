# Windows Current-User Baseline Roots

## Goal

Add the first Windows `baseline` default roots as an upstream-friendly compatibility-layer slice. This tranche adds safe current-user Windows root discovery only; it must not fork scanner semantics, parsers, schemas, output, sinks, exposure matching, root kinds, profile meanings, or CLI flags.

## Original Request

The user asked for the next careful `windows.md` plan while preserving the compatibility-layer approach. The approved plan is Goal 3A: current-user Windows baseline roots, excluding browsers.

## Outcome

`baseline` on Windows can discover safe current-user roots for Go, editor extensions, and MCP config locations when those directories exist. Absent candidate roots remain non-fatal. `windows.md` Goal 3 is updated only for verified items, while browser coverage, multi-user scanning, walker hardening, deployment docs, endpoint identity, WSL, and native Windows ecosystems remain out of scope.

## Oracle

The goal is complete only when:

- a Scout/Judge gate confirms the exact Windows candidate roots and allowed files;
- Windows candidate roots are implemented behind the existing compatibility-layer root-discovery boundary;
- `bumblebee roots --profile baseline` tests prove Windows current-user candidates are included only when present and tagged with correct `root_kind`;
- absent Windows candidates are skipped without failing if at least one baseline root exists;
- browser roots are not added in this tranche;
- `windows.md` Goal 3 marks only verified items complete;
- `go test ./...`, `go test -race ./...`, Windows build/selftest, and `git diff --check` pass.

## In-Scope Candidate Roots

- `%USERPROFILE%\go`
- `%USERPROFILE%\.vscode\extensions`
- `%USERPROFILE%\.vscode-insiders\extensions`
- `%USERPROFILE%\.cursor\extensions`
- `%USERPROFILE%\.cursor-server\extensions`
- `%USERPROFILE%\.windsurf\extensions`
- `%USERPROFILE%\.windsurf-server\extensions`
- `%USERPROFILE%\.vscodium\extensions`
- `%APPDATA%\Claude`

## Non-Goals

- Do not add browser roots. Goal 4 owns browser extension coverage.
- Do not implement Windows `--all-users`.
- Do not add Windows Python, npm/global, pipx, Ruby/Bundler, Composer, or native package-manager roots in this first slice unless the Judge narrows/approves a specific source-backed addition.
- Do not add WSL behavior.
- Do not add deployment docs.
- Do not change endpoint identity.
- Do not add NuGet, PowerShell, Chocolatey, Scoop, winget/MSIX/AppX, or Visual Studio extension inventory.
- Do not change parser, schema, sink, exposure, CLI, root-kind, profile, or emitted-record semantics.

## Likely Misfire

The dangerous failure mode is treating baseline roots as permission to scan broad Windows user data or browser profile trees. This slice should only add bounded, existing-directory roots that map to already-supported parsers and existing `RootKind` values.

## Starter Command

`/goal Follow docs/goals/windows-baseline-roots-current-user/goal.md.`
