# Windows Firefox Browser Root

## Original Request

Prepare a GoalBuddy board for the next `windows.md` Goal 4 slice: add Windows
Firefox browser-extension coverage after the Chrome/Edge slice.

## Interpreted Outcome

The next `/goal` run should implement and verify Windows Firefox profile-root
coverage as a compatibility-layer root discovery slice.

## Goal Oracle

The tranche is complete when tests and the Windows smoke harness prove:

- `baseline` resolves `%APPDATA%\Mozilla\Firefox\Profiles` as
  `browser_extension_root` when present.
- Firefox extension discovery uses the existing `extensions.json` parser path.
- Firefox cache, cookies, history, storage, session, and other profile data are
  not added as scan roots.
- The redacted Windows smoke receipt reports Firefox profile-root presence,
  root listing, aggregate profile counts, and `scan_summary.status=complete`
  without raw add-on/profile details.
- `windows.md` marks only the Firefox Goal 4 and smoke evidence actually proven
  by this tranche.

## Non-Negotiable Constraints

- Keep Windows support as a compatibility layer, not a scanner fork.
- Do not change schemas, emitted record fields, sink behavior, CLI flags,
  profile names, root kinds, exposure matching, or endpoint identity.
- Do not add Brave, Chromium, Vivaldi, LibreWolf, Waterfox, all-users, WSL,
  native Windows ecosystems, package-manager roots, or deployment docs in this
  tranche.
- Do not add Firefox cache, cookies, history, local storage, session, or
  per-extension XPI/source directories as roots.
- Keep raw NDJSON and browser inventory details out of the repo.

## Existing Plan Facts

- Add Firefox root under `%APPDATA%\Mozilla\Firefox\Profiles`.
- Use `APPDATA`; if unset, no Firefox candidate should be emitted.
- Rely on the existing browser extension parser and scanner dispatch for
  `extensions.json`.
- Update `scripts/windows-smoke.ps1` with redacted Firefox aggregate fields.
- Local pre-plan probe found `%APPDATA%\Mozilla\Firefox\Profiles` exists with
  two profiles, and one profile has `extensions.json`, so real smoke proof
  should be available.

## Likely Misfire

The main misfire is treating Firefox like Chromium and scanning the wrong
directory shape, or widening the slice into LibreWolf/Waterfox and broad browser
data. Firefox support should add only the profile-parent root required to reach
per-profile `extensions.json`.

## Enough For This Tranche

Firefox baseline browser extension root discovery is implemented, tested,
smoke-verified on this machine, and documented in `windows.md`. LibreWolf,
Waterfox, and remaining Chromium-family browsers stay unchecked.
