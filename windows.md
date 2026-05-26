# Windows Support Plan

This file tracks the Windows fork/port work for Bumblebee. Work through one
checkbox at a time, keeping each change small enough to test and review.

## Compatibility-Layer Principle

- [x] Keep Windows support as an upstream-friendly compatibility layer, not a divergent scanner fork.

Windows work should preserve the shared scanner model, output schema, parsers,
emitter, exposure matching, and safety rules wherever possible. Prefer
OS-specific root discovery, deployment documentation, CI/release additions, and
small `*_windows.go` adapters over changes that fork core scanner behavior.

## Core Requirement

- [x] A Windows `bumblebee.exe` reliably performs read-only inventory scans over explicit Windows roots and emits schema-compatible NDJSON with a complete `scan_summary`.

This is the first support bar. The scanner must build on Windows, pass
`selftest`, scan explicit `C:\...` roots without executing package managers,
preserve `source_file`, `project_path`, `root_kind`, dedupe, and record ID
behavior, and degrade access-denied paths into diagnostics rather than failed
healthy scans.

## Goal 1: Establish The Windows Build Baseline

- [x] Add Windows to CI with `windows-latest`.
- [x] Add a Windows `go test ./...` job.
- [x] Add a Windows `go build ./cmd/bumblebee` job.
- [x] Add a Windows self-test job once the binary builds.
- [x] Add `windows` to `.goreleaser.yaml` `goos`.
- [x] Emit `.exe` binaries for Windows releases.
- [x] Use `.zip` archives for Windows release artifacts.
- [x] Document the minimum supported Windows versions.
- [x] Confirm whether `os/user`, signal handling, path handling, and race tests pass on Windows.

Why: before changing behavior, the fork needs proof that the current CLI,
parsers, output sinks, and self-test can compile and run on Windows.

Minimum Windows support policy:

- Runtime floor: this fork currently targets Go 1.25, and Go 1.21 and later
  require Windows 10 or higher, or Windows Server 2016 or higher. That Go
  requirement is the compatibility layer's minimum runtime floor, not a broad
  product support promise for every Windows install at or above that version.
- Operator support target: supported Windows operator paths should be limited
  to actively Microsoft-serviced Windows client and Windows Server releases
  that also meet the Go runtime floor. Windows 11 serviced releases and
  Windows Server 2016, 2019, 2022, and 2025 LTSC/LTSB releases are the source
  backed families at the time of this check.
- Windows 10 caveat: Windows 10 version 22H2 and listed Windows 10 Enterprise
  LTSB 2015 editions reached end of support on 2025-10-14 and no longer
  receive security updates after that date. Treat Windows 10 as
  runtime-compatible with the Go floor, but do not generally claim normal
  Windows 10 support unless the operator is on a valid serviced LTSC or ESU
  path.
- CI validation target: CI validates GitHub Actions `windows-latest`, which is
  Windows Server 2025 x64 at the time of verification. GitHub's `-latest`
  label intentionally follows the newest stable OS image, so this validation
  target may move over time.
- Artifact and architecture scope: GoReleaser builds Windows artifacts for
  `amd64` and `arm64`. Current runtime smoke validation is Windows `amd64`;
  Windows `arm64` is release-built but not yet runtime-smoke-validated.

Sources verified 2026-05-24:

- https://go.dev/wiki/MinimumRequirements
- https://github.com/actions/runner-images
- https://learn.microsoft.com/en-us/windows/release-health/windows11-release-information
- https://learn.microsoft.com/en-gb/lifecycle/announcements/windows-10-end-of-support
- https://learn.microsoft.com/en-us/windows/release-health/windows-server-release-info

## Goal 2: Support Explicit Windows Roots

- [x] Verify `--profile deep --root C:\...` works against a small fixture tree.
- [x] Verify `--profile project --root C:\...` works against a project tree.
- [x] Add tests using Windows-shaped paths where possible.
- [x] Confirm `source_file`, `project_path`, and `root_kind` are stable and readable with Windows paths.
- [x] Check record IDs for undesirable drift caused only by path separator differences.
- [x] Validate explicit roots on paths containing spaces, such as `Application Data` or `Program Files`.
- [x] Document that explicit-root scanning is the first supported Windows mode.

Why: explicit roots should be the lowest-risk Windows milestone because most
ecosystem parsers only read metadata files and do not depend on OS-specific
package-manager execution.

## Goal 3: Add Windows Baseline Root Defaults

- [x] Add `%USERPROFILE%\go`.
- [x] Add source-backed user npm/global package roots under `%APPDATA%\npm\node_modules`.
- [x] Add Windows Python user site roots under `%APPDATA%\Python\Python*\site-packages`.
- [ ] Defer `%LOCALAPPDATA%` and common Python install roots until they are source-validated as a separate compatibility slice.
- [x] Add source-backed pipx venv roots under `%USERPROFILE%\pipx\venvs`, `%LOCALAPPDATA%\pipx\venvs`, and `%USERPROFILE%\.local\pipx\venvs`.
- [ ] Defer arbitrary virtualenv discovery; project/deep scans already find virtualenv metadata under supplied roots.
- [ ] Defer Ruby/Bundler user package roots until there is a cross-platform baseline decision.
- [ ] Defer Composer user/global package roots until there is a cross-platform baseline decision.
- [x] Add Windows VS Code extension roots.
- [x] Add Windows Cursor extension roots.
- [x] Add Windows Windsurf extension roots.
- [x] Add Windows VSCodium extension roots.
- [x] Add Windows MCP config roots, including `%APPDATA%\Claude` and `%LOCALAPPDATA%\Packages\Claude_pzs8sxrjxfjjc\LocalCache\Roaming\Claude`.
- [x] Keep absent root candidates non-fatal, matching macOS/Linux behavior.
- [x] Add `bumblebee roots --profile baseline` tests for Windows default candidates.

Why: the current default roots are intentionally macOS/Linux-shaped. Windows
baseline support requires native AppData, user profile, and tool-specific paths.

## Goal 4: Add Windows Browser Extension Coverage

- [x] Add Chrome profile extension roots under `%LOCALAPPDATA%\Google\Chrome\User Data`.
- [x] Add Edge profile extension roots under `%LOCALAPPDATA%\Microsoft\Edge\User Data`.
- [x] Add Brave profile extension roots under `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data`.
- [x] Add Chromium profile extension roots under `%LOCALAPPDATA%\Chromium\User Data`.
- [x] Add Vivaldi profile extension roots under `%LOCALAPPDATA%\Vivaldi\User Data`.
- [x] Add Firefox profile roots under `%APPDATA%\Mozilla\Firefox\Profiles`.
- [x] Add LibreWolf and Waterfox Windows profile roots if their layouts are reliable.
- [x] Preserve the current narrow profile strategy: `Default` and `Profile 1` through `Profile 9`.
- [x] Avoid scanning cookies, login databases, local storage, cache, and history.
- [x] Add Windows browser root tests.

Why: the browser scanners are mostly portable, but the curated baseline roots
currently know only macOS and Linux profile locations.

Receipt from 2026-05-25 remaining Windows browser root validation:

- Added Windows baseline roots for Brave, Chromium, and Vivaldi using the same
  narrow Chromium-family strategy as Chrome and Edge: only `Default` and
  `Profile 1` through `Profile 9` `Extensions` directories are candidates.
- Added Firefox-family profile-parent roots for LibreWolf and Waterfox. Waterfox
  includes both `%APPDATA%\Waterfox\Waterfox\Profiles`, which is documented by
  Waterfox support, and `%APPDATA%\Waterfox\Profiles`, which was created by the
  signed Waterfox 6.6.13 installer during local validation.
- Validated real root creation from official sources: Brave official release
  zip with matching release SHA256, Chromium official Win_x64 snapshot
  `LAST_CHANGE=1635789`, signed Vivaldi Technologies AS installer, LibreWolf
  official portable build signed by OSSign as documented by LibreWolf, and
  signed BrowserWorks Waterfox installer with matching official SHA512.
- Extended Windows sensitive browser-profile guards so LibreWolf and both
  Waterfox profile layouts keep cookies, history, credential stores, storage,
  cache, and per-extension payload directories out of deep scans while leaving
  `extensions.json` scannable.
- Added controlled smoke coverage for Brave, Chromium, Vivaldi, LibreWolf, and
  both Waterfox profile layouts. The smoke fixture proves package-record
  emission without reading or writing real personal browser profile contents.
- Corrective live-profile validation replaced an invalid empty-directory check:
  `go run ./cmd/bumblebee scan --profile baseline --ecosystem browser-extension`
  emitted real package records from browser-created data for Brave, Chromium,
  Vivaldi, LibreWolf, and the observed `%APPDATA%\Waterfox\Profiles` layout.
  The scan output is outside the repo at
  `%TEMP%\bumblebee-live-browser-test-20260525-154256\browser-extension-scan.ndjson`.
  The alternate `%APPDATA%\Waterfox\Waterfox\Profiles` layout remains
  source-documented but not observed on this host.

## Goal 5: Design Windows Multi-User Scanning

- [x] Decide whether `--all-users` should support Windows or remain unsupported initially.
- [x] If supported, enumerate real user homes under the local Windows profile parent.
- [x] Filter service/system profiles such as `Public`, `Default`, `Default User`, `All Users`, and `desktop.ini`.
- [x] Handle profile directories on non-`C:` drives if needed.
- [x] Decide how domain, Azure AD, and redirected profile paths should be discovered.
- [x] Decide whether admin rights are required for multi-user baseline scans.
- [x] Add tests equivalent to the current macOS `/Users/<name>` expansion tests.
- [x] Ensure `--all-users` never adds bare home directories for `baseline` or `project`.

Why: macOS `--all-users` is built around `/Users/<name>`. Windows fleet support
needs explicit profile enumeration rules and permission expectations.

Receipt: Windows `--all-users` is implemented as a compatibility-layer
equivalent of the macOS behavior. `baseline` and `project` now expand curated
per-user roots across local Windows profile directories, while `deep` and
explicit `--root` remain invalid with `--all-users`. Windows v1 intentionally
uses local profile-directory enumeration only: no registry, SID, domain, Azure
AD, OneDrive, or redirected-profile discovery was added. The default profile
parent comes from `BUMBLEBEE_USERS_DIR` in tests, otherwise `%SystemDrive%\Users`,
then the drive from `USERPROFILE`, then `C:\Users`. Admin rights are not enforced
by the CLI; elevated deployment may be needed operationally to read other users'
profiles, and unreadable paths should surface through normal scanner diagnostics.

## Goal 6: Harden The Windows Walker

- [x] Review Windows symlink, junction, and reparse point behavior.
- [x] Replace or supplement the non-Unix `dirKey` path fallback if needed.
- [x] Ensure deep scans do not loop through junctions.
- [x] Add Windows-specific excludes for high-cost or sensitive paths.
- [x] Exclude Windows browser profile data that is not needed for extension inventory.
- [x] Exclude credential and cloud-sync sensitive directories.
- [x] Account for the current user's redirected `Documents` known-folder path for PowerShell module roots.
- [ ] Account for broader OneDrive and all-users redirected known folders.
- [x] Ensure ACL-denied paths produce diagnostics without failing healthy scans.
- [x] Add tests for inaccessible paths where Windows permits stable test setup.

Why: the current walker has strong Unix/macOS assumptions around inode identity,
TCC-style access errors, and home-directory noise.

Receipt from 2026-05-24 Windows walker safety hardening:

- Added a Windows reparse-point directory guard behind platform-specific walker
  code so junctions and other directory links are skipped before descent, while
  shared traversal, parser, schema, sink, profile, and root-kind behavior stays
  unchanged.
- Supplemented the non-Unix `dirKey` path fallback with the Windows
  reparse-point guard; no record identity or output path semantics changed.
- Added Windows junction tests proving a junction to an outside tree is not
  descended into and a junction loop back to an ancestor does not hang the
  walker.
- Added a Windows `icacls`-backed ACL-denial scanner test proving a healthy
  project scan still emits package records and reports the denied directory as
  a structured `debug` diagnostic.
- Current-user redirected `Documents` support is implemented only for curated
  PowerShell module roots. Broader OneDrive crawling and all-users redirected
  known-folder discovery remain unchecked.

## Goal 7: Normalize Windows Endpoint Identity

- [x] Decide whether `endpoint.uid` should contain a Windows SID.
- [x] Update docs so downstream consumers do not assume numeric Unix UIDs.
- [x] Verify `endpoint.username` is emitted, redacted shape-classified, and
  consistent across package and `scan_summary` records on the current Windows
  smoke host.
- [ ] Validate live local-account and Azure AD account-provider username shapes
  on representative hosts when available.
- [x] Keep `endpoint.device_id` as the preferred stable identity.
- [x] Document Windows device ID provisioning through environment variables.
- [x] Add tests around endpoint fields that are stable on Windows.

Why: Windows identity is not POSIX UID-based. The schema can remain compatible,
but the meaning needs to be clear.

Receipt: Windows endpoint identity keeps the shared endpoint schema unchanged.
`endpoint.uid` is the Windows user SID returned by Go `os/user` for the
scanner process when user lookup succeeds; the Windows fallback intentionally
leaves `uid` empty rather than emitting Go's `os.Getuid()` value of `-1`.
`endpoint.username` remains the scanner-process account name. The Windows smoke
script now validates the evidence-backed part of that contract without writing
raw identity values: username is present, its shape is redacted to a class, and
the same value is used across package and `scan_summary` records. The exact
shape is still account-provider dependent; live local-account and Azure AD
representative-host validation remains open until those hosts are available.
`endpoint.device_id` remains the preferred stable machine correlation key and
is populated only from the environment variable named by `--device-id-env`.
Windows operators should provision that value from an existing fleet identity
such as MDM, RMM, EDR, Microsoft Entra, Intune, or a provisioning script.
Bumblebee must not automatically derive `device_id` from `MachineGuid`, SMBIOS
UUID, hostname, registry state, Entra state, Intune state, or hardware
identifiers.

## Goal 8: Add Windows Deployment Documentation

- [x] Create Windows deployment guidance separate from `docs/deployment-macos.md`.
- [x] Document Task Scheduler deployment.
- [x] Document Intune/RMM/SCCM-style deployment assumptions.
- [x] Document one-shot incident response runs with explicit `--root`.
- [x] Document recurring baseline runs.
- [x] Document output to file plus log shipper.
- [x] Document HTTPS output with token/HMAC secrets supplied by environment variables.
- [x] Document required permissions for current-user and all-user scans.
- [x] Document recommended cadence by profile.
- [x] Document Windows verification steps.

Why: launchd, TCC, and `/Users` guidance does not apply to Windows operators.

Receipt: `docs/deployment-windows.md` now documents Windows deployment for the
compatibility layer only. It covers Task Scheduler, Intune/RMM/SCCM deployment
assumptions, recurring baseline runs, explicit-root incident-response runs,
file output for log shippers, HTTPS bearer/HMAC output with environment-backed
secrets, current-user versus elevated `--all-users` permissions, profile
cadence, stable `--device-id-env` identity, verification steps, and the
operator meaning of `diagnostics_count`. README and broader user-facing docs
are now covered by Goal 9.

## Goal 9: Update User-Facing Docs

- [x] Update `README.md` scope from macOS/Linux only once Windows support is real.
- [x] Update `docs/inventory-sources.md` with Windows profile-to-source mapping.
- [x] Add Windows root examples to quick start.
- [x] Add Windows notes for MCP config locations.
- [x] Add Windows notes for browser extension profile locations.
- [x] Add Windows notes for editor extension roots.
- [x] Clearly label unsupported Windows ecosystems.
- [x] Add a Windows troubleshooting section.

Why: docs should not claim Windows support until the behavior is implemented
and tested, but the fork needs a checklist for every doc touchpoint.

Receipt: `README.md` now describes Windows as a compatibility layer rather
than broad full support, adds PowerShell quick-start examples, links deployment
details to `docs/deployment-windows.md`, lists the tested Windows support
boundary, and adds Windows troubleshooting for empty roots, `device_id`,
`diagnostics_count`, ACL diagnostics, `--all-users`, WSL, and raw evidence
hygiene. `docs/inventory-sources.md` now includes Windows profile mapping,
Windows Claude Desktop MCP locations, Windows Chrome/Edge/Firefox notes,
Windows VS Code-family editor roots, and explicit unsupported Windows
ecosystems. Goals 3, 4, 6, 7, 10, 11, and 13 remain open where behavior is not
implemented or tested.

## Goal 10: Decide Windows-Native Ecosystem Scope

- [x] Decide whether NuGet is in scope.
- [x] Decide whether PowerShell modules are in scope.
- [x] Decide whether Chocolatey packages are in scope.
- [x] Decide whether Scoop packages are in scope.
- [x] Decide whether winget/MSIX/AppX inventory is in scope.
- [x] Decide whether Visual Studio extensions are in scope.
- [x] Decide whether Cargo/Maven/Gradle should be handled as cross-platform follow-ups rather than Windows-specific work.
- [x] Implement NuGet project/deep parser support for `packages.config` and `packages.lock.json`.
- [ ] Revisit NuGet global package-cache baseline roots only after project/deep metadata support is implemented and output volume is understood.
- [x] Design PowerShell module manifest support, including `.psd1` parsing and the emitted ecosystem name, before implementation.
- [x] Keep unsupported and deferred ecosystems explicitly documented.

Why: Windows support can ship without Windows-native ecosystems, but full
developer endpoint coverage probably needs at least NuGet and PowerShell
module inventory.

Decision from 2026-05-25:

- NuGet is in scope and should be the first Windows-native ecosystem slice.
  The first implementation should scan project/deep metadata files only:
  `packages.config` and `packages.lock.json`. Emit a shared `nuget`
  ecosystem and keep `package_manager=nuget`. Do not add `%USERPROFILE%\.nuget`
  or other NuGet global package-cache baseline roots in the first slice,
  because cache inventory has different volume and installed-state semantics
  than project metadata.
- PowerShell modules are in scope as manifest inventory. The design uses
  file-based `.psd1` parsing and the emitted ecosystem
  `powershell-module`. Do not execute PowerShell package-management commands
  to discover modules.
- Chocolatey and Scoop are deferred for this compatibility-layer phase. They
  are useful endpoint tooling inventory, but they are closer to installed app
  package-manager state than developer project dependency metadata.
- winget, MSIX, and AppX inventory are out of scope for this phase. They would
  push the scanner toward installed-application inventory, Windows APIs,
  registry state, permissions-sensitive package locations, or command output.
- Visual Studio extensions are deferred. VSIX manifests are parseable metadata,
  but reliable installed-root discovery across Visual Studio instances is a
  separate path-discovery problem and should not be mixed into the NuGet slice.
- Cargo, Maven, and Gradle are not Windows-native support work. They should be
  handled as shared cross-platform parser follow-ups so Windows does not create
  separate behavior for ecosystems that also matter on macOS and Linux.

Implementation receipt from 2026-05-25:

- Added shared NuGet parser support for project/deep scans over
  `packages.config` and `packages.lock.json`. Records emit `ecosystem=nuget`,
  `package_manager=nuget`, and source types `nuget-packages-config` or
  `nuget-lockfile`.
- `packages.config` emits high-confidence records for package entries with
  both `id` and `version`. `packages.lock.json` emits high-confidence records
  for resolved non-project dependencies and marks direct/transitive dependency
  state when available.
- Scanner dispatch, `--ecosystem nuget` filtering, README coverage, and
  inventory-source documentation were updated through shared parser/model paths.
  No NuGet package-manager execution, registry discovery, Visual Studio API,
  `obj/project.assets.json`, or global package-cache baseline root was added.
- Verified with the local Go toolchain: `go test ./internal/ecosystem/nuget
  ./internal/scanner`, `go test ./cmd/bumblebee ./internal/model`, and
  `go test ./cmd/bumblebee ./internal/...`.

Smoke receipt from 2026-05-25:

- `scripts\windows-smoke.ps1` passed after prepending the existing local Go
  toolchain to `PATH` for that process. The redacted summary reported 9
  baseline roots, 3 browser extension roots, 1 editor extension root, 5 MCP
  config roots, 1,042 baseline package records, 0 findings, 5 duplicates, 4
  diagnostics, no timeout, no summary error, and a complete HTTP sink project
  scan with bearer auth over loopback.
- A focused temporary NuGet project smoke used raw evidence outside the repo
  and emitted 3 package records from `packages.config` and
  `packages.lock.json`. All package records had `ecosystem=nuget`,
  `root_kind=project_root`, and source types `nuget-packages-config` or
  `nuget-lockfile`; the local project reference in the lockfile was not
  emitted, and the scan summary was `status=complete`.

Gap follow-up receipt from 2026-05-25 full validation:

- NuGet lockfile `requested` ranges are now preserved in the existing
  `requested_spec` field, while `version` remains the resolved version.
- `scripts\windows-smoke.ps1` now creates a NuGet fixture under a path with
  spaces and includes a redacted `nuget_project_scan` summary. The smoke
  validates 5 NuGet package records, 2 `nuget-packages-config` records, 3
  `nuget-lockfile` records, 3 `requested_spec` values, complete summary
  status, project-root stamping, required field coverage, direct/transitive
  counts, and expected skips for project references, missing versions, and
  missing resolved versions.
- The smoke helper now quotes captured command arguments before
  `Start-Process` so Windows PowerShell 5.1 preserves roots containing spaces.
- These remain intentional boundaries rather than bugs: `packages.config`
  records leave `direct_dependency` empty, duplicate package/version records
  from `packages.config` and `packages.lock.json` are source-accurate, and
  NuGet global package-cache baseline roots remain deferred.

PowerShell module manifest design from 2026-05-25:

- PowerShell modules use the project-local emitted ecosystem
  `powershell-module`, with `package_manager=powershell` and
  `source_type=powershell-module-manifest`. This avoids claiming an OSV or
  Gallery ecosystem mapping that Bumblebee does not implement.
- Package identity comes from the `.psd1` manifest filename without extension;
  `version` comes from the top-level `ModuleVersion` manifest key. Manifests
  without a constant scalar `ModuleVersion` are skipped instead of emitting
  weak or invented package records.
- Parsing must remain text/file based. Do not execute PowerShell, import
  modules, run `Test-ModuleManifest`, call PowerShellGet/PSResourceGet, query
  PowerShell Gallery, read the registry, or evaluate PowerShell expressions.
- Windows baseline roots may include current-user and all-users module roots
  when present:
  `%USERPROFILE%\Documents\PowerShell\Modules`,
  `%USERPROFILE%\Documents\WindowsPowerShell\Modules`,
  the current user's resolved Windows `Documents` known-folder path with
  `PowerShell\Modules` and `WindowsPowerShell\Modules`,
  `%ProgramFiles%\PowerShell\Modules`, and
  `%ProgramFiles%\WindowsPowerShell\Modules`.
- All-users redirected Documents, broader OneDrive discovery, custom
  `PSModulePath` registry entries, Gallery/API discovery, and command-based
  installed-module inventory remain out of scope for this compatibility-layer
  slice.

PowerShell module manifest implementation receipt from 2026-05-25:

- Added the shared `powershell-module` ecosystem and scanner dispatch for
  `.psd1` manifests without changing the output schema or creating a
  Windows-only scanner fork.
- Added a conservative file parser for top-level constant scalar
  `ModuleVersion` values. The implementation derives package identity from
  the manifest filename, emits `package_manager=powershell` and
  `source_type=powershell-module-manifest`, skips manifests without a usable
  version, and does not evaluate PowerShell expressions.
- Added Windows current-user and all-users PowerShell module roots for
  baseline scans when those directories exist. Current-user roots include the
  resolved Windows `Documents` known-folder path; all-users redirected
  Documents, broader OneDrive discovery, custom `PSModulePath`, registry
  lookup, Gallery/API discovery, and command-based installed-module inventory
  remain out of scope.
- Added parser tests, scanner integration coverage, Windows root tests,
  model-derived CLI help coverage, and Windows smoke validation for a
  PowerShell module fixture.
- Verified with focused Go tests, full `go test ./...`, `go vet ./...`,
  `CGO_ENABLED=1 go test -race ./...`, `scripts\windows-smoke.ps1`, and
  `git diff --check`. The smoke emitted one `powershell-module` package from a
  Pester fixture, produced a complete project `scan_summary`, and skipped
  missing-version and expression-version manifests.

Goal 6A smoke gap-fill receipt from 2026-05-25:

- [x] Updated `scripts\windows-smoke.ps1` to create a run-specific temporary
  PowerShell module under the actual current user's Windows `Documents`
  known-folder path.
- [x] The smoke now fails if the real known-folder `PowerShell\Modules` root is
  not listed by `bumblebee roots`, or if the temporary module is not emitted as
  a `powershell-module-manifest` package in the baseline scan.
- [x] The smoke removes the temporary module after the run and uses a trap to
  clean up on terminating errors.
- [x] Added `-RequireRedirectedDocuments` to make redirected-Documents
  validation an explicit manual/lab gate. When the flag is set, the smoke fails
  before build/scan unless Windows reports that the current user's `Documents`
  known-folder path differs from `%USERPROFILE%\Documents`.
- [x] Verified on this host with a redacted summary under the outside-repo
  smoke evidence directory
  plus a post-run cleanup check:
  `known_documents_smoke_fixture_created=true`,
  `known_documents_powershell_modules_exists=true`,
  `known_documents_powershell_modules_listed=true`,
  `known_documents_smoke_package_emitted=true`, and
  `leftover_smoke_module_count=0`.
- [x] Verified `-RequireRedirectedDocuments` fails clearly on this
  non-redirected host instead of silently passing an inapplicable validation.
- [ ] A real redirected-Documents host is still needed to prove an actual
  production redirection policy end to end by running
  `powershell -ExecutionPolicy Bypass -File scripts\windows-smoke.ps1 -RequireRedirectedDocuments`;
  this machine's known `Documents` path still resolves to the standard
  `%USERPROFILE%\Documents` location.

## Goal 11: Define WSL Behavior

- [x] Decide whether the Windows binary should inspect WSL filesystems: it should not auto-discover or claim WSL package state.
- [x] If not, document that WSL should run the Linux Bumblebee binary inside each distro.
- [x] If yes, define how distro paths are discovered; not applicable because no Windows-side WSL discovery is implemented.
- [x] Avoid claiming WSL coverage from Windows user-profile scanning alone.
- [x] Add tests or fixtures only after the behavior is explicitly chosen.

Why: WSL is Linux userland with Linux package/tool layouts. Treating it as
ordinary Windows filesystem coverage would be misleading.

Receipt from 2026-05-25 Windows WSL boundary:

- Goal 11 uses the Generic Only policy: Windows baseline and `--all-users`
  defaults do not discover WSL UNC paths, `\\wsl.localhost`, or distro
  `rootfs` package state.
- Operator-supplied explicit `--root` paths remain generic explicit roots when
  Windows can read them. They are not rejected merely for looking WSL-related,
  but they are not documented as supported WSL inventory.
- WSL package state should be inventoried by running the Linux Bumblebee binary
  inside each distro.
- Added Windows root tests and smoke validation for `wsl_root_count=0` so the
  compatibility layer does not silently start claiming WSL default roots.
- Verified with a redacted summary under the outside-repo smoke evidence
  directory:
  `wsl_root_count=0` and `failures=0`.

## Goal 12: Validate End To End On Windows

- [x] Build `bumblebee.exe` locally on Windows.
- [x] Run `bumblebee.exe selftest`.
- [x] Run `bumblebee.exe roots --profile baseline`.
- [x] Run an explicit-root project scan against a fixture repo.
- [x] Run a baseline scan on a real Windows developer profile.
- [x] Run a browser-extension baseline scan on a profile with Chrome or Edge.
- [x] Run a Windows Claude Desktop MCP config root scan.
- [x] Run an HTTP sink smoke test against a local endpoint.
- [x] Run a file output smoke test with append mode.
- [x] Confirm `scan_summary.status=complete` for healthy runs.
- [x] Capture known limitations before declaring Windows support complete.

Why: implementation is not complete until the Windows binary proves the same
operator workflow as macOS/Linux: roots preview, scan, output, summary, and
self-test.

Smoke receipt from 2026-05-24 real current-user Windows run:

- `roots --profile baseline` resolved 5 roots: 1 editor extension root and 4 MCP config roots.
- No browser roots, bare `%USERPROFILE%` root, all-users roots, or WSL roots appeared in the baseline preview.
- `scan --profile baseline --output file` completed successfully with 10,343 files considered, 1,013 package records, 0 findings, 5 duplicates, 3 informational diagnostics, no timeout, and no summary error.
- `%APPDATA%\Claude` was absent on the test machine, so the first smoke run did not prove Claude Desktop config coverage.
- Raw NDJSON inventory should stay outside the repo; only redacted summary counts belong in this plan file.

Smoke receipt from 2026-05-24 Claude Desktop MSIX fix:

- `roots --profile baseline` resolved 6 roots: 1 editor extension root and 5 MCP config roots.
- Claude Desktop MSIX config root was present and listed; `%APPDATA%\Claude` remained absent.
- `scan --profile baseline --output file` completed successfully with 14,605 files considered, 1,013 package records, 0 findings, 5 duplicates, 4 informational diagnostics, no timeout, and no summary error.
- No browser roots, bare `%USERPROFILE%` root, all-users roots, or WSL roots appeared in the baseline preview.

Smoke receipt from 2026-05-24 Chrome/Edge browser root fix:

- `roots --profile baseline` resolved 8 roots: 2 browser extension roots, 1 editor extension root, and 5 MCP config roots.
- Chrome and Edge `Default\Extensions` roots were present and listed; no browser profile parent directory was emitted as a root.
- `scan --profile baseline --output file` completed successfully with 19,278 files considered, 1,034 package records, 0 findings, 5 duplicates, 4 informational diagnostics, no timeout, and no summary error.
- No bare `%USERPROFILE%` root, all-users roots, or WSL roots appeared in the baseline preview.

Smoke receipt from 2026-05-24 Firefox browser root fix:

- `roots --profile baseline` resolved 9 roots: 3 browser extension roots, 1 editor extension root, and 5 MCP config roots.
- Firefox `%APPDATA%\Mozilla\Firefox\Profiles` was present and listed; 2 Firefox profiles were found, and 1 had `extensions.json`.
- Chrome and Edge `Default\Extensions` roots remained present and listed; LibreWolf and Waterfox remain unchecked.
- `scan --profile baseline --output file` completed successfully with 19,482 files considered, 1,042 package records, 0 findings, 5 duplicates, 4 informational diagnostics, no timeout, and no summary error.
- No raw Firefox add-on or profile details were written to this plan; only redacted aggregate counts belong here.

Smoke receipt from 2026-05-24 Windows walker privacy hardening:

- Added default excludes for Windows browser profile parents, credential/security stores, cache trees, `Packages`, and cloud-sync folders such as OneDrive, Google Drive, and Dropbox.
- Added path-scoped scanner skips for sensitive browser profile files such as browser cookies, login data, history, Firefox SQLite profile databases, and sessionstore files before they count as considered files.
- Regression tests prove Windows sensitive directories are pruned while normal project metadata and Firefox `extensions.json` remain scannable; curated Chrome/Edge `Extensions` roots are not pruned when used as scan roots.
- `scan --profile baseline --output file` completed successfully with 19,438 files considered, 1,042 package records, 0 findings, 5 duplicates, 4 informational diagnostics, no timeout, and no summary error.
- OneDrive basename and `OneDrive - <tenant>` folder pruning is implemented.
  Current-user redirected `Documents` support is limited to curated PowerShell
  module roots; broader redirected-known-folder policy remains unchecked.

Maintenance receipt from 2026-05-24 Windows privacy platform-hook refactor:

- Moved Windows-only privacy policy out of shared walker/scanner bodies and behind `*_windows.go` / `*_nonwindows.go` compatibility hooks.
- Kept exported `walk.DefaultExcludes` as a variable to avoid API churn.
- Preserved the Goal 6A behavior: real Windows smoke still resolved 9 roots, kept 3 browser extension roots, listed Chrome/Edge/Firefox roots, and completed with 19,438 files considered, 1,042 package records, 0 findings, 5 duplicates, 4 informational diagnostics, no timeout, and no summary error.
- No new Goal 6 capability item was claimed by this refactor; junction/reparse,
  `dirKey`, current-user redirected `Documents`, broader redirected known
  folders, and ACL-denied-path work remained tracked separately.

Smoke receipt from 2026-05-24 Windows HTTP sink validation:

- Added a local loopback HTTP receiver to `scripts\windows-smoke.ps1`; no external ingest service is required.
- The smoke generates a temporary project fixture, supplies a generated bearer token through `BUMBLEBEE_SMOKE_HTTP_TOKEN`, and verifies the receiver sees valid bearer auth without writing the token to the redacted summary.
- `scan --profile project --output http` completed successfully against the local receiver; the receiver saw 2 requests, 0 auth failures, 1 package record, and 1 `scan_summary`.
- The HTTP `scan_summary` had `status=complete`, `profile=project`, `http_batches_attempted=1`, `http_batches_succeeded=1`, `http_last_status=200`, and no raw received NDJSON was written to the repo.
- The same smoke run also preserved the real-profile baseline proof: 9 roots, 3 browser extension roots, 1,042 package records, 0 findings, 5 duplicates, 4 informational diagnostics, no timeout, and no summary error.

Full smoke receipt from 2026-05-24 Windows compatibility validation:

- CI-parity validation passed with the local Go toolchain: `go vet ./...`, `go test ./...`, `go test -race ./...`, Windows build, and `bumblebee.exe selftest`. Formatting was validated against a clean LF checkout-index export to avoid local CRLF working-tree noise.
- Race tests used MSYS2 UCRT64 GCC installed under the local `bumblebee-tools` tool directory. The MSYS2 installer was downloaded from the official MSYS2 GitHub release, verified against the official SHA-256 file, and GCC reported the MSYS2 16.1.0 build.
- Windows release-build smoke produced both `windows/amd64` and `windows/arm64` binaries in temporary evidence storage. Runtime execution was validated on `amd64` only; `arm64` remains build-only smoke coverage on this host.
- `scripts\windows-smoke.ps1` passed with raw evidence outside the repo: 9 baseline roots, 3 browser extension roots, 1 editor extension root, 5 MCP config roots, no bare `%USERPROFILE%` root, 1,042 baseline package records, 1 complete baseline `scan_summary`, 0 findings, 5 duplicates, 4 diagnostics, no timeout, and no summary error.
- The same script validated local HTTP output with bearer auth: 2 loopback requests, 0 auth failures, 0 parse failures, 1 package record, 1 complete project `scan_summary`, 1 attempted/succeeded HTTP batch, 0 failed HTTP batches, and HTTP 200 as the last status.
- Focused explicit-root fixture smokes passed for both `project` and `deep` profiles over a path containing spaces. Each emitted 1 package record and a complete `scan_summary`; `project` stamped `root_kind=project_root`, while `deep` stamped `root_kind=unknown`.
- Focused endpoint smoke passed with `--device-id-env`: package and `scan_summary` endpoint fields matched, `device_id` was present, `uid` was Windows-SID-shaped rather than `-1`, and `username` was present. Raw identity values were not written to this receipt.
- Focused controlled-root smoke passed for implemented root families: Go user root, VS Code-family editor extension root, Claude MCP config root, Chrome extension root, Edge extension root, and Firefox profile root. The fixture emitted 5 package records across editor-extension, MCP, and browser-extension ecosystems with a complete baseline `scan_summary`.
- Focused controlled `--all-users` smoke passed with a temporary users directory: 2 real user homes expanded, service/system profile names filtered, 0 bare home roots emitted, 2 package records emitted, and the baseline `scan_summary` completed.
- Raw NDJSON, HTTP payloads, SIDs, hostnames, usernames, tokens, and full profile paths stayed outside the repo. This receipt records only redacted aggregate evidence and does not close the remaining Windows gaps below.

Known limitations / current support boundary:

- Tested support: the current Windows compatibility layer builds
  `bumblebee.exe`, passes `selftest`, previews baseline roots, scans explicit
  project roots, scans a real current-user baseline, scans implemented browser
  extension roots when present, scans Windows Claude Desktop MCP config roots
  when present, writes file output in append mode, sends HTTP output to a local
  endpoint, and emits schema-compatible NDJSON with
  `scan_summary.status=complete` for healthy runs.
- Platform support boundary: the Windows compatibility layer documents a Go
  runtime floor of Windows 10 or Windows Server 2016 and newer, but the
  operator support claim is limited to actively Microsoft-serviced Windows
  client and Windows Server releases. CI currently validates GitHub Actions
  `windows-latest`, which is Windows Server 2025 x64 at the time of
  verification. Windows `amd64` and `arm64` artifacts are built, but runtime
  smoke validation is currently Windows `amd64`; `arm64`, older Windows, and
  non-serviced Windows 10 paths remain unclaimed unless separately tested and
  serviced.
- Root coverage boundary: baseline roots currently cover the implemented
  Windows Go user root, editor extension roots, MCP config roots,
  npm global root under `%APPDATA%\npm\node_modules`, Python user site roots
  under `%APPDATA%\Python\Python*\site-packages`, pipx venv roots under
  `%USERPROFILE%\pipx\venvs`, `%LOCALAPPDATA%\pipx\venvs`, and
  `%USERPROFILE%\.local\pipx\venvs`, plus the shared cross-platform
  `%USERPROFILE%\.local\share\pipx\venvs` candidate, Chrome/Edge/Brave/
  Chromium/Vivaldi extension roots, and Firefox/LibreWolf/Waterfox profile
  roots. Arbitrary virtualenv discovery, custom npm/Python/pipx prefixes,
  Ruby/Bundler, and Composer roots remain deferred until separate
  compatibility-layer or cross-platform baseline decisions are made.
- Browser boundary: Chrome, Edge, Brave, Chromium, Vivaldi, Firefox,
  LibreWolf, and Waterfox are the exercised browser families so far. Chromium
  was validated from an official snapshot archive rather than a normal stable
  installer, and Waterfox keeps both documented and observed profile-parent
  variants in scope.
- Multi-user boundary: Windows `--all-users` uses local profile-directory
  enumeration only, matching the compatibility-layer approach. Registry, SID,
  domain, Azure AD, OneDrive, and redirected-profile discovery are not claimed;
  current-user `Documents` known-folder resolution is not applied to other
  profiles. Elevated deployment may still be needed operationally to read other
  users' profiles.
- Walker/privacy boundary: Windows sensitive-path excludes, directory
  reparse-point skipping, junction loop safety, ACL-denied diagnostics, and
  current-user redirected `Documents` handling for curated PowerShell module
  roots are implemented. Broader OneDrive/redirected-known-folder behavior
  remains open Goal 6 work.
- Endpoint identity boundary: Windows `endpoint.uid` is documented as the
  scanner-process SID, and `endpoint.device_id` remains the preferred stable
  machine identity supplied through `--device-id-env`. The Windows smoke
  validates scanner-process `endpoint.username` presence, redacted shape class,
  and package/`scan_summary` consistency on the current host. Live
  representative-host validation for local-account and Azure AD account
  provider shapes remains open Goal 7 work.
- Deployment/docs boundary: Windows deployment guidance is now captured in
  `docs/deployment-windows.md` for Task Scheduler, Intune/RMM/SCCM,
  incident-response, recurring baseline, file/log-shipper, HTTPS secret,
  permission, cadence, and verification guidance.
- User-facing docs boundary: README and inventory-source docs now describe the
  Windows compatibility layer, but they intentionally keep unimplemented roots,
  native ecosystems, WSL, all-users redirected known folders, and broad
  OneDrive discovery outside the support claim.
- Native ecosystem boundary: Goal 10 now decides the Windows-native ecosystem
  scope. NuGet project/deep metadata and PowerShell `.psd1` module manifest
  inventory are the implemented Windows-native slices. NuGet global cache
  roots remain deferred. Chocolatey, Scoop, winget/MSIX/AppX, and Visual
  Studio extensions are deferred or out of scope for this phase. Cargo, Maven,
  and Gradle remain cross-platform follow-ups. Unsupported and deferred native
  ecosystems must stay explicitly documented rather than implied by "Windows
  support."
- WSL boundary: no WSL filesystem coverage is claimed from the Windows binary.
  Run the Linux Bumblebee binary inside each distro for WSL package inventory.
- Diagnostics boundary: the Windows deployment guide clarifies that
  `diagnostics_count` includes informational diagnostics, warnings, and errors.

Use `powershell -ExecutionPolicy Bypass -File scripts\windows-smoke.ps1` for
future real-profile validation runs. The script builds into `%TEMP%`, runs the
core Windows smoke checks, keeps raw NDJSON outside the repo, and writes only
`smoke-summary.redacted.json` for review.

Known gaps from the smoke run:

- [x] Run a real-profile baseline on a machine with `%USERPROFILE%\go` present, because the first real smoke did not exercise a real `user_package_root`.
- [x] Run a real-profile MCP smoke on a machine with a Windows Claude Desktop config root present.
- [x] Decide whether operator-facing docs should clarify that `diagnostics_count` includes informational diagnostics, not only warnings or errors.
- [x] Add a redacted smoke-test receipt pattern for future Windows validation runs so raw NDJSON inventory is never checked in.
- [x] Keep the remaining browser families open until separately exercised.

Real-profile `%USERPROFILE%\go` smoke receipt from 2026-05-25:

- [x] Updated `scripts\windows-smoke.ps1` to create a run-specific temporary
  Go module fixture under the actual `%USERPROFILE%\go` baseline root.
- [x] The smoke now fails if `%USERPROFILE%\go` is not listed by
  `bumblebee roots --profile baseline`, or if the temporary `go.mod`
  dependency is not emitted as a `go-mod` package in the baseline scan.
- [x] Verified on this host with a redacted summary under the outside-repo
  smoke evidence directory:
  `userprofile_go_smoke_fixture_created=true`,
  `userprofile_go_exists=true`, `userprofile_go_listed=true`,
  `userprofile_go_smoke_package_emitted=true`, and `failures=0`.
- [x] Post-run cleanup left `leftover_go_smoke_project_count=0`; because
  `%USERPROFILE%\go` did not preexist on this host, the smoke removed the
  temporary root after validation.

Goal 3A strict-parity receipt:

- [x] Source-validated `%APPDATA%\npm\node_modules`, `%APPDATA%\Python\Python*\site-packages`, `%USERPROFILE%\pipx\venvs`, `%LOCALAPPDATA%\pipx\venvs`, and `%USERPROFILE%\.local\pipx\venvs`.
- [x] Added only literal, existence-filtered Windows baseline root candidates in the Windows platform hook.
- [x] Kept `%USERPROFILE%\.local\share\pipx\venvs` in shared baseline handling rather than duplicating it in the Windows hook.
- [x] Added controlled tests and smoke fixtures that prove npm and PyPI records emit from the new Windows roots.
- [x] Kept Ruby/Bundler, Composer, custom npm/Python/pipx prefixes, arbitrary virtualenv discovery, registry reads, WSL, and redirected known folders out of the current support claim.

## Goal 13: Keep The Fork Easy To Update

- [x] Keep repo layout unchanged unless an upstreamable refactor requires it.
- [x] Keep metadata parsers shared across macOS, Linux, and Windows.
- [x] Keep the output schema identical across platforms.
- [x] Keep exposure matching shared across platforms.
- [x] Keep sink behavior shared across platforms.
- [x] Isolate Windows root discovery behind small OS-specific functions.
- [x] Prefer `*_windows.go` files or narrow `runtime.GOOS == "windows"` branches for platform-specific behavior.
- [x] Avoid Windows-only semantic changes to existing profile meanings.
- [x] Add tests that prove Windows behavior without weakening macOS/Linux tests.
- [x] Keep Windows docs additive instead of rewriting existing macOS/Linux docs.
- [x] Structure commits so CI/release, explicit-root support, baseline defaults, browser roots, and docs can be reviewed separately.
- [ ] Rebase regularly from upstream `main` while Windows support is still fork-only.
- [ ] Open upstream PRs in small slices when possible.
- [ ] Treat conflicts in shared parser/schema/output code as a warning sign that the compatibility layer is leaking.

Why: the fork should continue receiving upstream parser, schema, threat-catalog,
and safety improvements with minimal conflict. The durable boundary is platform
discovery and deployment guidance, not a separate Windows scanner.

Current branch model:

- `upstream/main` is the Perplexity source of truth.
- `origin/windows/compat-layer` is the TakeThree-maintained Windows fork branch.
- Local `main` should remain a clean mirror of `upstream/main`.
- Local `windows/compat-layer` should track `origin/windows/compat-layer`.
- Upstream PRs to Perplexity are optional and not part of the current workflow.

Routine update workflow:

```powershell
git fetch upstream
git switch main
git merge --ff-only upstream/main
git switch windows/compat-layer
git rebase main
go test ./cmd/bumblebee ./internal/...
powershell -ExecutionPolicy Bypass -File scripts\windows-smoke.ps1
git push --force-with-lease origin windows/compat-layer
```

Maintenance receipt from 2026-05-24 base-code cleanup:

- Removed the shared `StableID` path-normalization change; record identity is
  back to base semantics while Windows path fields remain preserved as emitted.
- Moved Windows root-classification and browser-candidate literals behind
  `*_windows.go` hooks so shared root resolution stays orchestration-focused.
- Moved Windows-only root, scanner, and walker tests into build-tagged Windows
  test files where practical; shared tests now cover shared behavior.
- Verified with `go test ./cmd/bumblebee ./internal/...`, the Windows smoke
  script, whitespace diff checks, and targeted searches for removed shared drift.
