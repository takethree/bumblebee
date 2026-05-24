# Windows Chrome/Edge Browser Roots

## Original Request

Prepare a GoalBuddy board for Goal 4A from `windows.md`: add Windows browser
extension coverage for Chrome and Edge while preserving Bumblebee's Windows
compatibility-layer boundary.

## Interpreted Outcome

The next `/goal` run should implement and verify the first Windows browser
extension baseline slice: Chrome and Edge profile `Extensions` roots under
`%LOCALAPPDATA%`, with tests and a redacted real-profile smoke receipt.

## Goal Oracle

The tranche is complete when tests and the Windows smoke harness prove:

- `baseline` resolves Windows Chrome and Edge `Default\Extensions` and
  `Profile 1` through `Profile 9\Extensions` roots when present.
- Browser roots are tagged `browser_extension_root`.
- Parent browser profile directories, cookies, login databases, local storage,
  cache, and history are not emitted as roots.
- The real Windows smoke receipt shows Chrome/Edge browser roots listed on this
  machine, `scan_summary.status=complete`, no bare `%USERPROFILE%` root, and no
  raw browser inventory in the redacted summary.
- `windows.md` marks only the Goal 4 and Goal 12 items proven by this tranche.

## Non-Negotiable Constraints

- Keep Windows support as a compatibility layer, not a scanner fork.
- Do not change schemas, emitted record fields, parser behavior, sink behavior,
  CLI flags, profile names, root kinds, exposure matching, or endpoint identity.
- Do not add Brave, Chromium, Vivaldi, Firefox, LibreWolf, Waterfox, all-users,
  WSL, native Windows ecosystems, package-manager roots, or deployment docs in
  this tranche.
- Do not scan browser profile parents as roots. Only Chromium-family
  `Extensions` directories are in scope.
- Keep raw NDJSON and browser inventory details out of the repo.

## Existing Plan Facts

- Add Chrome roots under `%LOCALAPPDATA%\Google\Chrome\User Data`.
- Add Edge roots under `%LOCALAPPDATA%\Microsoft\Edge\User Data`.
- Use only `Default` and `Profile 1` through `Profile 9`.
- Update `scripts/windows-smoke.ps1` so browser roots are expected after this
  slice and the receipt stays redacted.
- Use the existing browser extension parser; do not modify browser parsing
  unless Scout/Judge finds a direct blocker.

## Likely Misfire

The main misfire is widening Goal 4A into all browser support or scanning full
browser profile directories. That would risk cookies, history, login databases,
local storage, cache, and unrelated browser state. This tranche must stay at
the root discovery boundary.

## Enough For This Tranche

Chrome and Edge baseline browser extension roots are implemented, tested,
smoke-verified on this machine, and documented in `windows.md`. Other browser
families remain unchecked for later slices.
