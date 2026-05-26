# Deploying bumblebee on Windows

`bumblebee` is a one-shot binary with no built-in scheduler. On Windows
the usual pattern is to run `bumblebee.exe` from Task Scheduler, Intune,
Configuration Manager, or the RMM/EDR remote-execution tool the fleet already
uses.

This document is for the Windows compatibility layer. It describes deployment
patterns for behavior already implemented and tested. It does not claim WSL
coverage, unsupported Windows-native package ecosystems, all-users redirected
known-folder coverage, or browser families that are still open in `windows.md`.

For WSL package inventory, deploy or run the Linux Bumblebee binary inside each
distro. The Windows binary does not discover distro filesystems or treat WSL
package state as Windows baseline coverage.

Cadence is the runner's choice. The profile only controls what gets walked:

- `baseline`: bounded package/tool roots, editor extensions, MCP config roots,
  and implemented browser-extension roots. Recurring, typically every 6 hours.
- `project`: configured developer/project roots. Recurring, typically daily or
  every 12 hours.
- `deep`: exposure scan over operator-supplied roots. Usually on demand during
  a campaign; recurring is acceptable only for bounded roots and is usually
  paired with `--findings-only`.

For backend data-modeling considerations see [state-model.md](state-model.md).
For sink behavior and HTTP wire details see [transport.md](transport.md).

## Current-user vs all-user scans

Pick one operating shape for a scheduled run:

- **Current-user task**: runs as the developer account. This is the safer
  default when inventory should represent that user's own profile. It can read
  what that account can read and emits `endpoint.username` / `endpoint.uid` for
  the scanner process account.
- **Elevated all-user task**: runs under a privileged service account or
  `SYSTEM` and passes `--all-users`. Use this when one host-level job should
  expand the baseline or project defaults across real local profile directories
  under `C:\Users`. Records still carry the scanner process identity, so use
  `source_file`, `project_path`, and `scan_summary.roots` to attribute records
  to a user profile.

`--all-users` on Windows intentionally uses local profile-directory
enumeration. It filters service/system profile names and does not query the
registry, enumerate SIDs, discover domain or Azure AD accounts, inspect OneDrive
redirected known folders, or crawl bare user homes. It cannot be combined with
explicit `--root` entries or `--profile deep`.

Current-user baseline scans resolve the Windows `Documents` known-folder path
for PowerShell module roots, so a user's redirected or OneDrive-moved Documents
folder can still contribute `PowerShell\Modules` and
`WindowsPowerShell\Modules` roots when those directories exist. This does not
extend to `--all-users` redirected known-folder discovery.

The walker is read-only. ACL-denied or unavailable paths are reported as
diagnostics and do not by themselves mean the scan failed. `diagnostics_count`
in `scan_summary` counts all emitted diagnostics, including informational
diagnostics, warnings, and errors.

## Stable endpoint identity

Prefer a stable opaque `endpoint.device_id` supplied through
`--device-id-env`. The value should come from a fleet identity already trusted
by the operator, such as an MDM, RMM, EDR, Microsoft Entra, Intune, or
provisioning-script asset identifier.

Do not derive `device_id` inside Bumblebee from `MachineGuid`, SMBIOS UUID,
hostname, registry state, Entra state, Intune state, or hardware identifiers.
Hostnames and account names are useful context, not stable receiver keys.

## Output choice

- **Existing log shipper**: write local NDJSON and let the shipper own
  forwarding, retries, compression, and buffering.

  ```powershell
  & "C:\Program Files\Bumblebee\bumblebee.exe" scan `
    --profile baseline `
    --max-duration 5m `
    --output file `
    --output-file "C:\ProgramData\Bumblebee\inventory.ndjson" `
    --append `
    --device-id-env BUMBLEBEE_DEVICE_ID
  ```

- **Direct HTTPS ingest**: use `--output http` when there is no endpoint log
  pipeline. HTTPS is required for non-loopback hosts unless
  `--http-allow-insecure` is explicitly used for testing. Bearer tokens and
  HMAC keys are read from environment variables, not from CLI literals.

  ```powershell
  & "C:\Program Files\Bumblebee\bumblebee.exe" scan `
    --profile baseline `
    --max-duration 5m `
    --output http `
    --http-url https://inventory.example.com/v1/ingest `
    --http-auth bearer `
    --http-token-env BUMBLEBEE_TOKEN `
    --http-gzip `
    --device-id-env BUMBLEBEE_DEVICE_ID
  ```

For HMAC mode, replace the bearer options with:

```powershell
--http-auth hmac-sha256 `
--http-hmac-key-env BUMBLEBEE_HMAC_KEY
```

There is no built-in S3, GCS, or Azure Blob uploader. Send to a local file plus
a shipper, or to a small internal HTTP relay.

## Task Scheduler baseline example

Use a wrapper script so the scheduled task command line contains no secrets.
Fleet tooling should provision `BUMBLEBEE_TOKEN` and `BUMBLEBEE_DEVICE_ID` as
machine or process environment variables before the script runs.

Example `C:\ProgramData\Bumblebee\run-baseline.ps1`:

```powershell
$ErrorActionPreference = "Stop"

$requiredEnv = @("BUMBLEBEE_TOKEN", "BUMBLEBEE_DEVICE_ID")
foreach ($name in $requiredEnv) {
  if ([string]::IsNullOrWhiteSpace([Environment]::GetEnvironmentVariable($name))) {
    throw "$name is not set"
  }
}

& "C:\Program Files\Bumblebee\bumblebee.exe" scan `
  --profile baseline `
  --max-duration 5m `
  --output http `
  --http-url https://inventory.example.com/v1/ingest `
  --http-auth bearer `
  --http-token-env BUMBLEBEE_TOKEN `
  --http-gzip `
  --device-id-env BUMBLEBEE_DEVICE_ID

exit $LASTEXITCODE
```

Register a current-user recurring task:

```powershell
$action = New-ScheduledTaskAction `
  -Execute "PowerShell.exe" `
  -Argument "-NoProfile -ExecutionPolicy Bypass -File `"C:\ProgramData\Bumblebee\run-baseline.ps1`""
$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(5) `
  -RepetitionInterval (New-TimeSpan -Hours 6)
$settings = New-ScheduledTaskSettingsSet -StartWhenAvailable

Register-ScheduledTask `
  -TaskName "Bumblebee Baseline" `
  -TaskPath "\Bumblebee\" `
  -Action $action `
  -Trigger $trigger `
  -Settings $settings `
  -Description "Run Bumblebee baseline inventory every 6 hours"
```

Register an elevated all-user variant by changing the wrapper command to add
`--all-users`, then register the task with an elevated principal chosen by the
operator, such as a managed service account or `SYSTEM`:

```powershell
$principal = New-ScheduledTaskPrincipal `
  -UserId "SYSTEM" `
  -LogonType ServiceAccount `
  -RunLevel Highest

Register-ScheduledTask `
  -TaskName "Bumblebee Baseline All Users" `
  -TaskPath "\Bumblebee\" `
  -Action $action `
  -Trigger $trigger `
  -Settings $settings `
  -Principal $principal `
  -Description "Run Bumblebee baseline inventory across local user profiles"
```

`schtasks.exe` is also acceptable when a deployment system standardizes on it.
Schedule the wrapper script rather than embedding secrets in `/tr`:

```cmd
schtasks /create /tn "\Bumblebee\Baseline" /sc hourly /mo 6 /f ^
  /tr "PowerShell.exe -NoProfile -ExecutionPolicy Bypass -File C:\ProgramData\Bumblebee\run-baseline.ps1"
```

## Intune, RMM, and Configuration Manager assumptions

Bumblebee does not require a resident Windows service. Fleet tools only need to
deliver the binary, deliver a wrapper script, provide environment variables,
and run the wrapper on the chosen schedule or on demand.

Use this shape:

1. Package `bumblebee.exe` and wrapper scripts into the deployment artifact.
2. Install them under stable paths such as `C:\Program Files\Bumblebee\` and
   `C:\ProgramData\Bumblebee\`.
3. Provision `BUMBLEBEE_DEVICE_ID` and any HTTP auth secret through the fleet
   tool's secret or environment mechanism.
4. For recurring inventory, create or update a scheduled task.
5. For on-demand campaigns, run a one-shot wrapper that passes explicit
   `--root` values.
6. Treat a zero exit code plus a delivered `scan_summary.status=complete` as
   the successful run signal.

For Intune Win32 app deployment, use a silent installer or PowerShell installer
script and a detection rule for the installed binary or wrapper version. For
Intune Remediations, use detection/remediation scripts for on-demand or
recurring scan wrappers when that better matches the tenant's operations model.
For Configuration Manager, use an application or package deployment with a
script detection method. For RMM tools, use the same wrapper scripts and keep
secrets out of command-line arguments.

## One-shot incident response

`deep` has no default roots. Always pass at least one explicit `--root`.

Current-user example:

```powershell
& "C:\Program Files\Bumblebee\bumblebee.exe" scan `
  --profile deep `
  --root "$env:USERPROFILE\source" `
  --exposure-catalog "C:\ProgramData\Bumblebee\campaign-catalog.json" `
  --max-duration 10m `
  --output http `
  --http-url https://inventory.example.com/v1/ingest `
  --http-auth bearer `
  --http-token-env BUMBLEBEE_TOKEN `
  --device-id-env BUMBLEBEE_DEVICE_ID
```

Elevated campaign example across known local profiles:

```powershell
$usersRoot = "C:\Users"
$skip = @("All Users", "Default", "Default User", "Public")

Get-ChildItem -LiteralPath $usersRoot -Directory | Where-Object {
  $skip -notcontains $_.Name
} | ForEach-Object {
  & "C:\Program Files\Bumblebee\bumblebee.exe" scan `
    --profile deep `
    --root $_.FullName `
    --exposure-catalog "C:\ProgramData\Bumblebee\campaign-catalog.json" `
    --findings-only `
    --max-duration 10m `
    --output http `
    --http-url https://inventory.example.com/v1/ingest `
    --http-auth bearer `
    --http-token-env BUMBLEBEE_TOKEN `
    --device-id-env BUMBLEBEE_DEVICE_ID
}
```

Each invocation has its own `run_id`. Receiver-side current-state promotion
should follow the `scan_summary.status=complete` rule in
[state-model.md](state-model.md).

## Default roots per profile

Preview resolved roots before deployment:

```powershell
& "C:\Program Files\Bumblebee\bumblebee.exe" roots --profile baseline
& "C:\Program Files\Bumblebee\bumblebee.exe" roots --profile baseline --all-users
```

On Windows, the tested baseline support boundary currently includes the Go user
root, VS Code-family editor extension roots, Windows Claude Desktop MCP config
roots, Chrome extension roots, Edge extension roots, and Firefox profile roots
when those locations exist. WSL package state is intentionally handled by the
Linux binary inside each distro. Unimplemented language roots, additional
browser families, and Windows-native ecosystems remain tracked in `windows.md`.

If a profile's defaults pick up nothing on a host, the scan exits with a
helpful error rather than walking the whole profile directory.

## Verifying a deployment

On a representative host:

1. Confirm the binary runs:

   ```powershell
   & "C:\Program Files\Bumblebee\bumblebee.exe" selftest
   ```

2. Preview roots:

   ```powershell
   & "C:\Program Files\Bumblebee\bumblebee.exe" roots --profile baseline
   ```

3. Run a bounded explicit-root smoke:

   ```powershell
   & "C:\Program Files\Bumblebee\bumblebee.exe" scan `
     --profile project `
     --root "C:\ProgramData\Bumblebee\fixture project" `
     --output file `
     --output-file "C:\ProgramData\Bumblebee\smoke.ndjson" `
     --append `
     --device-id-env BUMBLEBEE_DEVICE_ID
   ```

4. Confirm the records stream includes one trailing
   `record_type=scan_summary` with `status=complete`.
5. Confirm `endpoint.device_id` is present when `--device-id-env` is expected.
6. For HTTP deployments, confirm the receiver accepted the final
   `scan_summary` for the new `run_id`; HTTP batch counters in the summary
   describe batches observed before the summary itself was flushed.
7. Run the repository smoke script on development hosts when validating the
   compatibility layer itself:

   ```powershell
   powershell -ExecutionPolicy Bypass -File scripts\windows-smoke.ps1
   ```

   On a lab or enterprise profile where Windows actually redirects the current
   user's `Documents` known folder, add `-RequireRedirectedDocuments`. That
   mode fails early on non-redirected profiles and is the validation gate for
   proving redirected-Documents policy end to end.
   The redacted summary also records endpoint username presence, shape class,
   and package/`scan_summary` consistency. Treat that as scanner-process
   context only, not as a stable endpoint key.

Keep raw NDJSON, HTTP payloads, tokens, SIDs, hostnames, usernames, and full
profile paths out of commits. Record only redacted aggregate receipts in
`windows.md`.

## Microsoft deployment references

These references are not Bumblebee requirements, but they are the source docs
behind the Windows deployment shapes above:

- [Register-ScheduledTask](https://learn.microsoft.com/en-us/powershell/module/scheduledtasks/register-scheduledtask)
- [schtasks create](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/schtasks-create)
- [Win32 app management in Microsoft Intune](https://learn.microsoft.com/en-us/intune/app-management/deployment/win32)
- [Intune Remediations](https://learn.microsoft.com/en-us/intune/device-management/tools/deploy-remediations)
- [Create applications in Configuration Manager](https://learn.microsoft.com/en-us/intune/configmgr/apps/deploy-use/create-applications)
