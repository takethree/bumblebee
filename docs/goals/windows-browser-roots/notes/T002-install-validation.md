# T002 Install Validation

This note records redacted install/run-once validation for the Windows browser
root tranche. Raw downloads and extracted browser files are outside the repo
under `F:\bumblebee-tools\browser-root-validation`.

## Download and Integrity Evidence

| Browser | Source | Verification |
| --- | --- | --- |
| Brave | `https://brave.com/download/`; fallback official Brave GitHub release `v1.90.124` | Web installer Authenticode valid for `Brave Software, Inc.`. Web installer failed non-interactively on this host, so validation used official release zip `brave-v1.90.124-win32-x64.zip`; zip SHA256 matched Brave's release `.sha256` file. |
| Chromium | `https://www.chromium.org/getting-involved/download-chromium/` | Official Chromium snapshot flow resolved Win_x64 `LAST_CHANGE=1635789` and `chrome-win.zip` from `commondatastorage.googleapis.com`; recorded SHA256 locally. This is a snapshot archive, not a normal signed stable installer. |
| Vivaldi | `https://vivaldi.com/download/` | `Vivaldi.x64.exe` downloaded from `downloads.vivaldi.com`; Authenticode valid for `Vivaldi Technologies AS`. Initial partial 20 MB download was rejected as unsigned and resumed to the full signed file before execution. |
| LibreWolf | `https://librewolf.net/installation/windows/` | Installer and portable executable Authenticode valid for `OSSign (Scheibling Consulting AB)`. LibreWolf's official Windows docs and FAQ identify that signer as expected. Installer path was not forced after a user-canceled/elevation gate; validation used the official portable zip to observe profile layout. |
| Waterfox | `https://www.waterfox.com/download/` | `Waterfox Setup 6.6.13.exe` Authenticode valid for `BrowserWorks Ltd`; SHA512 matched the official download page. |

## Observed Root Creation

| Browser | Expected root from plan | Observed result |
| --- | --- | --- |
| Brave | `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data` | Created on first launch from official Brave release zip. |
| Chromium | `%LOCALAPPDATA%\Chromium\User Data` | Created on first launch from official Chromium snapshot. |
| Vivaldi | `%LOCALAPPDATA%\Vivaldi\User Data` | Created by signed Vivaldi installer. |
| LibreWolf | `%APPDATA%\LibreWolf\Profiles` | Created on first launch from official LibreWolf portable build. |
| Waterfox | `%APPDATA%\Waterfox\Waterfox\Profiles` | Not created. Signed Waterfox installer succeeded, then `-CreateProfile default` and a normal first launch created `%APPDATA%\Waterfox\Profiles` instead. |

## Post-Implementation Live Data Validation

After implementation, the initial empty-directory check for Chromium-family
`Default\Extensions` roots was rejected as invalid. The empty test directories
were removed and replaced with a live scan against browser-created extension
data.

Command:

```powershell
go run ./cmd/bumblebee scan --profile baseline --ecosystem browser-extension --output file --output-file $outFile
```

Redacted output path:

`%TEMP%\bumblebee-live-browser-test-20260525-154256\browser-extension-scan.ndjson`

Observed package records from real profile data:

| Browser/path | Package records | Evidence type |
| --- | ---: | --- |
| `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data` | 2 | Real `Default\Extensions\<id>\<version>\manifest.json` files, including Google Docs Offline and Adobe Acrobat. |
| `%LOCALAPPDATA%\Chromium\User Data` | 1 | Real `Default\Extensions\<id>\<version>\manifest.json` file for Adobe Acrobat. |
| `%LOCALAPPDATA%\Vivaldi\User Data` | 1 | Real `Default\Extensions\<id>\<version>\manifest.json` file for Adobe Acrobat. |
| `%APPDATA%\LibreWolf\Profiles` | 8 | Real profile `extensions.json`, including uBlock Origin from LibreWolf's bundled AMO policy. |
| `%APPDATA%\Waterfox\Profiles` | 7 | Real profile `extensions.json` from the signed Waterfox install. |
| `%APPDATA%\Waterfox\Waterfox\Profiles` | 0 | Not observed on this host; retained because it is source-documented. |

The Chromium-family extension evidence came from actual browser writes. Existing
Chrome-compatible external-extension registry entries on the machine caused the
Adobe extension to materialize in Chromium and Vivaldi; Brave also installed
Google Docs Offline after the Chrome Web Store page's `Add to Chrome` action was
clicked through the browser. No extension directories were manually seeded.

## Safety Notes

- No browser sign-in, sync setup, import, or default-browser change was
  performed.
- The initial install/run-once validation proved root layout. The later
  corrective live-profile validation proved package-record emission from real
  browser-created extension metadata.
- Waterfox went through the Judge decision before implementation because
  observed local behavior conflicted with the initially planned nested profile
  path, despite Waterfox's support docs listing the nested path. The implemented
  compatibility layer keeps both the documented nested path and the locally
  observed `%APPDATA%\Waterfox\Profiles` path.
