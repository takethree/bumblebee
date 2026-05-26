# Windows Arc and Comet Browser Support

## Original Request

Fix the README gap and add Windows browser-extension support for Arc and Comet, correcting "comment" to Comet.

## Interpreted Outcome

Bumblebee's Windows compatibility layer accurately documents implemented Python install-prefix roots and supports Windows Comet and Arc browser-extension roots when source/local evidence proves their profile layouts, without broad browser-profile crawling or divergence from the shared scanner.

## Goal Type

specific / existing_plan

## Oracle

The tranche is complete only when:

- README, inventory docs, and `windows.md` accurately describe the supported Windows Python and browser-root boundary.
- Windows root discovery includes source-backed Comet roots and Arc roots only if the official install/profile layout is validated.
- Privacy/sensitive-path protections cover any new Windows browser profile parents.
- Tests and Windows smoke prove the implemented roots emit browser-extension records without reading full browser profile data.

## Constraints

- Keep this as a Windows compatibility layer, not a separate Windows scanner.
- Do not change the public output schema.
- Do not execute browsers, package managers, or registry discovery from Bumblebee scanner code.
- Do not fake Arc validation. If Arc cannot be installed or profile data cannot be created legitimately, mark Arc live validation as skipped and document the boundary.
- Use only reputable official sources for any download/install validation.
- Keep raw browser profile data, usernames, SIDs, hostnames, tokens, and full evidence payloads out of committed receipts.

## Existing Plan Facts

- README currently omits the implemented Windows Python install-prefix roots.
- Comet is locally present on this host under `%LOCALAPPDATA%\Perplexity\Comet\User Data`, with real Chromium-style extension manifests under `Default\Extensions`.
- Arc was not locally present during intake; its Windows layout needs source-backed validation before implementation.
- macOS shared browser roots already include Arc and Comet; Windows currently covers Chrome, Edge, Brave, Chromium, Vivaldi, Firefox, LibreWolf, and Waterfox.
- Candidate Comet root: `%LOCALAPPDATA%\Perplexity\Comet\User Data\<Default|Profile 1..9>\Extensions`.
- Candidate Arc root must be validated before use; expected package-family layout is `%LOCALAPPDATA%\Packages\TheBrowserCompany.Arc_ttt1ap7aakyb4\LocalCache\Local\Arc\User Data\<Default|Profile 1..9>\Extensions`.

## Likely Misfire

The goal could add undocumented or unvalidated browser paths, claim "full Windows browser parity" too broadly, or update docs without adding tests and smoke proof.

## Enough For This Tranche

Fix the README Python-root doc gap, implement and test Comet support, implement Arc support only if validation is legitimate, update docs/receipts, run verification, and finish with a Judge/PM audit that explicitly says whether Arc was supported or skipped.

