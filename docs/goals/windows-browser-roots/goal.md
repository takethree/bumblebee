# Windows Browser Roots

## Original Request

Prepare a GoalBuddy run for the next `windows.md` tranche: finish the
remaining Windows browser-extension roots while keeping Bumblebee's Windows work
as an upstream-friendly compatibility layer.

## Outcome

Bumblebee should add and verify Windows baseline root discovery for Brave,
Chromium, Vivaldi, LibreWolf, and Waterfox where reliable. The work must use
official browser sources for real-profile validation because those browsers are
not currently installed on this machine.

## Oracle

The tranche is complete when:

- `windows.md` Goal 4 reflects implemented browser-family support and remaining
  limitations.
- The Windows root resolver emits only narrow browser-extension roots or
  Firefox-family profile parents, not broad browser profile directories.
- Controlled tests prove the new roots classify correctly and do not weaken
  sensitive browser-profile skipping.
- Real install/run-once validation records redacted evidence that official
  installs create the expected profile locations, or explicitly defers any
  family whose layout cannot be source-backed and observed.
- Windows smoke and focused Go tests pass.

## Constraints

- Treat this as a compatibility-layer slice: root discovery only, shared
  browser-extension scanner behavior, no output schema changes, no browser APIs,
  no browser command output as inventory, and no personal profile data in repo.
- Install browsers only from official/reputable sources and verify installer or
  archive integrity before execution where the publisher provides a signature or
  checksum.
- Keep browsers installed after validation unless a source/install failure makes
  that unsafe.
- Chromium validation uses an official Chromium snapshot, not a third-party
  Chromium installer.
- Do not seed synthetic extensions into real browser profiles. Use controlled
  fixtures for package-record emission and real installs only for layout/root
  presence validation.
- If observed paths disagree with the proposed plan, stop and update the board
  before implementation.

## Existing Plan Facts

- Add Chromium-family candidates for `Default` and `Profile 1` through
  `Profile 9` under:
  - `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data`
  - `%LOCALAPPDATA%\Chromium\User Data`
  - `%LOCALAPPDATA%\Vivaldi\User Data`
- Add Firefox-family profile parents for:
  - `%APPDATA%\LibreWolf\Profiles`
  - `%APPDATA%\Waterfox\Waterfox\Profiles`
- Validate official sources:
  - Brave: `https://brave.com/download/`
  - Vivaldi: `https://vivaldi.com/download/`
  - LibreWolf: `https://librewolf.net/installation/windows/`
  - Waterfox: `https://www.waterfox.com/download/`
  - Chromium: `https://www.chromium.org/getting-involved/download-chromium/`
- Expected implementation areas are Windows root candidates/classification,
  Windows browser-sensitive path policy if needed, Windows smoke coverage, and
  additive docs.
