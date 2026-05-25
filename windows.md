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
- [ ] Add user npm/global package locations where reliable.
- [ ] Add user Python locations under `%APPDATA%`, `%LOCALAPPDATA%`, and common install roots.
- [ ] Add pipx and virtualenv-style user package locations where reliable.
- [ ] Add Ruby/Bundler user package locations if present.
- [ ] Add Composer user/global package locations if present.
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
- [ ] Add Brave profile extension roots under `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data`.
- [ ] Add Chromium profile extension roots under `%LOCALAPPDATA%\Chromium\User Data`.
- [ ] Add Vivaldi profile extension roots under `%LOCALAPPDATA%\Vivaldi\User Data`.
- [x] Add Firefox profile roots under `%APPDATA%\Mozilla\Firefox\Profiles`.
- [ ] Add LibreWolf and Waterfox Windows profile roots if their layouts are reliable.
- [x] Preserve the current narrow profile strategy: `Default` and `Profile 1` through `Profile 9`.
- [x] Avoid scanning cookies, login databases, local storage, cache, and history.
- [x] Add Windows browser root tests.

Why: the browser scanners are mostly portable, but the curated baseline roots
currently know only macOS and Linux profile locations.

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
- [ ] Account for OneDrive and redirected known folders.
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
- Broader redirected-known-folder policy remains unchecked; this tranche proves
  traversal safety and ACL diagnostics only.

## Goal 7: Normalize Windows Endpoint Identity

- [x] Decide whether `endpoint.uid` should contain a Windows SID.
- [x] Update docs so downstream consumers do not assume numeric Unix UIDs.
- [ ] Verify `endpoint.username` shape for local, domain, and Azure AD users.
- [x] Keep `endpoint.device_id` as the preferred stable identity.
- [x] Document Windows device ID provisioning through environment variables.
- [x] Add tests around endpoint fields that are stable on Windows.

Why: Windows identity is not POSIX UID-based. The schema can remain compatible,
but the meaning needs to be clear.

Receipt: Windows endpoint identity keeps the shared endpoint schema unchanged.
`endpoint.uid` is the Windows user SID returned by Go `os/user` for the
scanner process when user lookup succeeds; the Windows fallback intentionally
leaves `uid` empty rather than emitting Go's `os.Getuid()` value of `-1`.
`endpoint.username` remains the scanner-process account name, but the exact
shape is account-provider dependent and still needs broader local, domain, and
Azure AD validation before that checklist item is closed. `endpoint.device_id`
remains the preferred stable machine correlation key and is populated only from
the environment variable named by `--device-id-env`. Windows operators should
provision that value from an existing fleet identity such as MDM, RMM, EDR,
Microsoft Entra, Intune, or a provisioning script. Bumblebee must not
automatically derive `device_id` from `MachineGuid`, SMBIOS UUID, hostname,
registry state, Entra state, Intune state, or hardware identifiers.

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
remain untouched until Goal 9.

## Goal 9: Update User-Facing Docs

- [ ] Update `README.md` scope from macOS/Linux only once Windows support is real.
- [ ] Update `docs/inventory-sources.md` with Windows profile-to-source mapping.
- [ ] Add Windows root examples to quick start.
- [ ] Add Windows notes for MCP config locations.
- [ ] Add Windows notes for browser extension profile locations.
- [ ] Add Windows notes for editor extension roots.
- [ ] Clearly label unsupported Windows ecosystems.
- [ ] Add a Windows troubleshooting section.

Why: docs should not claim Windows support until the behavior is implemented
and tested, but the fork needs a checklist for every doc touchpoint.

## Goal 10: Decide Windows-Native Ecosystem Scope

- [ ] Decide whether NuGet is in scope.
- [ ] Decide whether PowerShell modules are in scope.
- [ ] Decide whether Chocolatey packages are in scope.
- [ ] Decide whether Scoop packages are in scope.
- [ ] Decide whether winget/MSIX/AppX inventory is in scope.
- [ ] Decide whether Visual Studio extensions are in scope.
- [ ] Decide whether Cargo/Maven/Gradle should be handled as cross-platform follow-ups rather than Windows-specific work.
- [ ] Create one parser task per accepted ecosystem.
- [ ] Keep unsupported ecosystems explicitly documented.

Why: Windows support can ship without Windows-native ecosystems, but full
developer endpoint coverage probably needs at least NuGet and PowerShell
module inventory.

## Goal 11: Define WSL Behavior

- [ ] Decide whether the Windows binary should inspect WSL filesystems.
- [ ] If not, document that WSL should run the Linux Bumblebee binary inside each distro.
- [ ] If yes, define how distro paths are discovered.
- [ ] Avoid claiming WSL coverage from Windows user-profile scanning alone.
- [ ] Add tests or fixtures only after the behavior is explicitly chosen.

Why: WSL is Linux userland with Linux package/tool layouts. Treating it as
ordinary Windows filesystem coverage would be misleading.

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
- OneDrive basename and `OneDrive - <tenant>` folder pruning is implemented, but the broader redirected-known-folder policy remains unchecked.

Maintenance receipt from 2026-05-24 Windows privacy platform-hook refactor:

- Moved Windows-only privacy policy out of shared walker/scanner bodies and behind `*_windows.go` / `*_nonwindows.go` compatibility hooks.
- Kept exported `walk.DefaultExcludes` as a variable to avoid API churn.
- Preserved the Goal 6A behavior: real Windows smoke still resolved 9 roots, kept 3 browser extension roots, listed Chrome/Edge/Firefox roots, and completed with 19,438 files considered, 1,042 package records, 0 findings, 5 duplicates, 4 informational diagnostics, no timeout, and no summary error.
- No new Goal 6 capability item is claimed by this refactor; junction/reparse, `dirKey`, broader redirected known folders, and ACL-denied-path work remain unchecked.

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
  project roots, scans a real current-user baseline, scans Chrome/Edge/Firefox
  browser extension roots when present, scans Windows Claude Desktop MCP config
  roots when present, writes file output in append mode, sends HTTP output to a
  local endpoint, and emits schema-compatible NDJSON with
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
  Windows Go user root, editor extension roots, MCP config roots, Chrome/Edge
  extension roots, and Firefox profile roots. User npm/global, Python, pipx,
  virtualenv, Ruby/Bundler, and Composer roots remain planned but unimplemented
  until Goal 3 items are completed.
- Browser boundary: Chrome, Edge, and Firefox are the exercised browser families
  so far. Brave, Chromium, Vivaldi, LibreWolf, and Waterfox roots remain
  unclaimed until their Goal 4 items are implemented and tested.
- Multi-user boundary: Windows `--all-users` uses local profile-directory
  enumeration only, matching the compatibility-layer approach. Registry, SID,
  domain, Azure AD, OneDrive, and redirected-profile discovery are not claimed;
  elevated deployment may still be needed operationally to read other users'
  profiles.
- Walker/privacy boundary: Windows sensitive-path excludes, directory
  reparse-point skipping, junction loop safety, and ACL-denied diagnostics are
  implemented. Broader OneDrive/redirected-known-folder behavior remains open
  Goal 6 work.
- Endpoint identity boundary: Windows `endpoint.uid` is documented as the
  scanner-process SID, and `endpoint.device_id` remains the preferred stable
  machine identity supplied through `--device-id-env`. Broader
  `endpoint.username` shape validation across local, domain, and Azure AD
  users remains open Goal 7 work.
- Deployment/docs boundary: Windows deployment guidance is now captured in
  `docs/deployment-windows.md` for Task Scheduler, Intune/RMM/SCCM,
  incident-response, recurring baseline, file/log-shipper, HTTPS secret,
  permission, cadence, and verification guidance. README and broader
  user-facing docs should not broadly claim Windows support until Goal 9 is
  completed.
- Native ecosystem boundary: NuGet, PowerShell modules, Chocolatey, Scoop,
  winget/MSIX/AppX, Visual Studio extensions, Cargo, Maven, and Gradle scope
  decisions are still open in Goal 10. Unsupported native ecosystems must stay
  explicitly documented rather than implied by "Windows support."
- WSL boundary: no WSL filesystem coverage is claimed. Until Goal 11 makes an
  explicit decision, WSL users should not assume the Windows binary inventories
  Linux distro package state.
- Diagnostics boundary: the Windows deployment guide clarifies that
  `diagnostics_count` includes informational diagnostics, warnings, and errors.

Use `powershell -ExecutionPolicy Bypass -File scripts\windows-smoke.ps1` for
future real-profile validation runs. The script builds into `%TEMP%`, runs the
core Windows smoke checks, keeps raw NDJSON outside the repo, and writes only
`smoke-summary.redacted.json` for review.

Known gaps from the smoke run:

- [ ] Run a real-profile baseline on a machine with `%USERPROFILE%\go` present, because the first real smoke did not exercise a real `user_package_root`.
- [x] Run a real-profile MCP smoke on a machine with a Windows Claude Desktop config root present.
- [x] Decide whether operator-facing docs should clarify that `diagnostics_count` includes informational diagnostics, not only warnings or errors.
- [x] Add a redacted smoke-test receipt pattern for future Windows validation runs so raw NDJSON inventory is never checked in.
- [ ] Keep the remaining browser families open until separately exercised.

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
