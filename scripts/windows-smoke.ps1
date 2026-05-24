param(
    [string]$EvidenceRoot = "",
    [int]$MaxDurationSeconds = 120,
    [switch]$KeepBinary
)

Set-StrictMode -Version 3.0
$ErrorActionPreference = "Stop"

$isWindowsValue = if (Get-Variable -Name IsWindows -Scope Global -ErrorAction SilentlyContinue) { $IsWindows } else { $env:OS -eq "Windows_NT" }
if (-not $isWindowsValue) {
    throw "windows-smoke.ps1 must be run on Windows."
}

if ($MaxDurationSeconds -le 0) {
    throw "-MaxDurationSeconds must be greater than zero."
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
if ([string]::IsNullOrWhiteSpace($EvidenceRoot)) {
    $EvidenceRoot = Join-Path $env:TEMP "bumblebee-windows-smoke\$stamp"
}
$EvidenceRoot = [System.IO.Path]::GetFullPath($EvidenceRoot)
New-Item -ItemType Directory -Force -Path $EvidenceRoot | Out-Null

$exePath = Join-Path $EvidenceRoot "bumblebee.exe"
$summaryPath = Join-Path $EvidenceRoot "smoke-summary.redacted.json"
$rootsOut = Join-Path $EvidenceRoot "roots-baseline.tsv"
$rootsErr = Join-Path $EvidenceRoot "roots-baseline.stderr.txt"
$scanOut = Join-Path $EvidenceRoot "baseline.ndjson"
$scanErr = Join-Path $EvidenceRoot "baseline.stderr.txt"

function Invoke-Captured {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$StdoutPath,
        [Parameter(Mandatory = $true)][string]$StderrPath,
        [string]$WorkingDirectory = $repoRoot
    )

    $proc = Start-Process `
        -FilePath $FilePath `
        -ArgumentList $Arguments `
        -WorkingDirectory $WorkingDirectory `
        -NoNewWindow `
        -Wait `
        -PassThru `
        -RedirectStandardOutput $StdoutPath `
        -RedirectStandardError $StderrPath
    return $proc.ExitCode
}

function Add-CommandResult {
    param(
        [Parameter(Mandatory = $true)][hashtable]$Commands,
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][int]$ExitCode
    )

    $Commands[$Name] = [ordered]@{
        exit_code = $ExitCode
        status = $(if ($ExitCode -eq 0) { "pass" } else { "fail" })
    }
}

function Read-JsonLines {
    param([Parameter(Mandatory = $true)][string]$Path)

    if (-not (Test-Path $Path)) {
        return @()
    }
    $records = New-Object System.Collections.Generic.List[object]
    foreach ($line in [System.IO.File]::ReadLines($Path)) {
        if ([string]::IsNullOrWhiteSpace($line)) {
            continue
        }
        try {
            $records.Add(($line | ConvertFrom-Json))
        } catch {
            throw "Failed to parse JSON line in $Path"
        }
    }
    return $records.ToArray()
}

function Get-JsonProperty {
    param(
        [object]$Object,
        [string]$Name
    )
    if ($null -eq $Object) {
        return $null
    }
    $prop = $Object.PSObject.Properties[$Name]
    if ($null -eq $prop) {
        return $null
    }
    return $prop.Value
}

$commands = @{}
$failures = New-Object System.Collections.Generic.List[string]

$goTestOut = Join-Path $EvidenceRoot "go-test.stdout.txt"
$goTestErr = Join-Path $EvidenceRoot "go-test.stderr.txt"
$code = Invoke-Captured "go" @("test", "-count=1", "./cmd/bumblebee", "./internal/model", "./internal/scanner") $goTestOut $goTestErr
Add-CommandResult $commands "go_test_key_packages" $code
if ($code -ne 0) { $failures.Add("go test failed") }

$buildOut = Join-Path $EvidenceRoot "go-build.stdout.txt"
$buildErr = Join-Path $EvidenceRoot "go-build.stderr.txt"
$code = Invoke-Captured "go" @("build", "-buildvcs=false", "-o", $exePath, "./cmd/bumblebee") $buildOut $buildErr
Add-CommandResult $commands "go_build_windows_binary" $code
if ($code -ne 0) { $failures.Add("go build failed") }

if ($code -eq 0) {
    $selftestOut = Join-Path $EvidenceRoot "selftest.stdout.txt"
    $selftestErr = Join-Path $EvidenceRoot "selftest.stderr.txt"
    $selftestCode = Invoke-Captured $exePath @("selftest") $selftestOut $selftestErr
    Add-CommandResult $commands "selftest" $selftestCode
    if ($selftestCode -ne 0) { $failures.Add("selftest failed") }

    $rootsCode = Invoke-Captured $exePath @("roots", "--profile", "baseline") $rootsOut $rootsErr
    Add-CommandResult $commands "roots_baseline" $rootsCode
    if ($rootsCode -ne 0) { $failures.Add("roots baseline failed") }

    $duration = "$($MaxDurationSeconds)s"
    $scanCode = Invoke-Captured $exePath @(
        "scan",
        "--profile", "baseline",
        "--output", "file",
        "--output-file", $scanOut,
        "--max-duration", $duration
    ) (Join-Path $EvidenceRoot "scan.stdout.txt") $scanErr
    Add-CommandResult $commands "scan_baseline_file" $scanCode
    if ($scanCode -ne 0) { $failures.Add("baseline scan failed") }
}

$roots = @()
if (Test-Path $rootsOut) {
    foreach ($line in [System.IO.File]::ReadLines($rootsOut)) {
        if ([string]::IsNullOrWhiteSpace($line)) {
            continue
        }
        $parts = $line -split "`t", 2
        if ($parts.Count -ne 2) {
            $failures.Add("roots output contained a non-TSV line")
            continue
        }
        $roots += [pscustomobject]@{ Kind = $parts[0]; Path = $parts[1] }
    }
}

$kindCounts = [ordered]@{}
foreach ($group in ($roots | Group-Object Kind | Sort-Object Name)) {
    $kindCounts[$group.Name] = $group.Count
}

$browserRootCount = @($roots | Where-Object { $_.Kind -eq "browser_extension_root" }).Count
$bareUserProfileRootCount = 0
if ($env:USERPROFILE) {
    $bareUserProfileRootCount = @($roots | Where-Object {
        [string]::Equals(
            [System.IO.Path]::GetFullPath($_.Path).TrimEnd('\'),
            [System.IO.Path]::GetFullPath($env:USERPROFILE).TrimEnd('\'),
            [System.StringComparison]::OrdinalIgnoreCase
        )
    }).Count
}

$goRoot = if ($env:USERPROFILE) { Join-Path $env:USERPROFILE "go" } else { "" }
$appDataClaude = if ($env:APPDATA) { Join-Path $env:APPDATA "Claude" } else { "" }
$msixClaude = if ($env:LOCALAPPDATA) { Join-Path $env:LOCALAPPDATA "Packages\Claude_pzs8sxrjxfjjc\LocalCache\Roaming\Claude" } else { "" }
$chromeDefaultExtensions = if ($env:LOCALAPPDATA) { Join-Path $env:LOCALAPPDATA "Google\Chrome\User Data\Default\Extensions" } else { "" }
$edgeDefaultExtensions = if ($env:LOCALAPPDATA) { Join-Path $env:LOCALAPPDATA "Microsoft\Edge\User Data\Default\Extensions" } else { "" }
$firefoxProfiles = if ($env:APPDATA) { Join-Path $env:APPDATA "Mozilla\Firefox\Profiles" } else { "" }
$goRootExists = if ($goRoot) { Test-Path $goRoot } else { $false }
$goRootListed = if ($goRoot) { @($roots | Where-Object { $_.Path -eq $goRoot }).Count -gt 0 } else { $false }
$appDataClaudeExists = if ($appDataClaude) { Test-Path $appDataClaude } else { $false }
$appDataClaudeListed = if ($appDataClaude) { @($roots | Where-Object { $_.Path -eq $appDataClaude }).Count -gt 0 } else { $false }
$msixClaudeExists = if ($msixClaude) { Test-Path $msixClaude } else { $false }
$msixClaudeListed = if ($msixClaude) { @($roots | Where-Object { $_.Path -eq $msixClaude }).Count -gt 0 } else { $false }
$chromeDefaultExtensionsExists = if ($chromeDefaultExtensions) { Test-Path $chromeDefaultExtensions } else { $false }
$chromeDefaultExtensionsListed = if ($chromeDefaultExtensions) { @($roots | Where-Object { $_.Path -eq $chromeDefaultExtensions }).Count -gt 0 } else { $false }
$edgeDefaultExtensionsExists = if ($edgeDefaultExtensions) { Test-Path $edgeDefaultExtensions } else { $false }
$edgeDefaultExtensionsListed = if ($edgeDefaultExtensions) { @($roots | Where-Object { $_.Path -eq $edgeDefaultExtensions }).Count -gt 0 } else { $false }
$firefoxProfilesExists = if ($firefoxProfiles) { Test-Path $firefoxProfiles } else { $false }
$firefoxProfilesListed = if ($firefoxProfiles) { @($roots | Where-Object { $_.Path -eq $firefoxProfiles }).Count -gt 0 } else { $false }
$firefoxProfileCount = 0
$firefoxExtensionsJsonProfileCount = 0
if ($firefoxProfilesExists) {
    $firefoxProfileDirs = @(Get-ChildItem -LiteralPath $firefoxProfiles -Directory -ErrorAction SilentlyContinue)
    $firefoxProfileCount = $firefoxProfileDirs.Count
    $firefoxExtensionsJsonProfileCount = @($firefoxProfileDirs | Where-Object {
        Test-Path -LiteralPath (Join-Path $_.FullName "extensions.json")
    }).Count
}

if ($bareUserProfileRootCount -gt 0) {
    $failures.Add("bare USERPROFILE appeared as a baseline root")
}

$records = Read-JsonLines $scanOut
$recordTypeCounts = [ordered]@{}
$summary = $null
foreach ($record in $records) {
    $recordType = [string]$record.record_type
    if ([string]::IsNullOrWhiteSpace($recordType)) {
        continue
    }
    if (-not $recordTypeCounts.Contains($recordType)) {
        $recordTypeCounts[$recordType] = 0
    }
    $recordTypeCounts[$recordType]++
    if ($recordType -eq "scan_summary") {
        $summary = $record
    }
}

if ($commands.Contains("scan_baseline_file") -and $commands["scan_baseline_file"].exit_code -eq 0 -and $null -eq $summary) {
    $failures.Add("scan_summary was missing from baseline NDJSON")
}
if ($null -ne $summary -and $summary.status -ne "complete") {
    $failures.Add("scan_summary status was not complete")
}

$redacted = [ordered]@{
    generated_at = (Get-Date).ToUniversalTime().ToString("o")
    evidence_dir = $EvidenceRoot
    max_duration_seconds = $MaxDurationSeconds
    commands = $commands
    roots_preview = [ordered]@{
        root_count = @($roots).Count
        kind_counts = $kindCounts
        browser_root_count = $browserRootCount
        bare_userprofile_root_count = $bareUserProfileRootCount
        userprofile_go_exists = [bool]$goRootExists
        userprofile_go_listed = [bool]$goRootListed
        claude_desktop_appdata_exists = [bool]$appDataClaudeExists
        claude_desktop_appdata_listed = [bool]$appDataClaudeListed
        claude_desktop_msix_exists = [bool]$msixClaudeExists
        claude_desktop_msix_listed = [bool]$msixClaudeListed
        appdata_claude_exists = [bool]$appDataClaudeExists
        appdata_claude_listed = [bool]$appDataClaudeListed
        chrome_default_extensions_exists = [bool]$chromeDefaultExtensionsExists
        chrome_default_extensions_listed = [bool]$chromeDefaultExtensionsListed
        edge_default_extensions_exists = [bool]$edgeDefaultExtensionsExists
        edge_default_extensions_listed = [bool]$edgeDefaultExtensionsListed
        firefox_profiles_exists = [bool]$firefoxProfilesExists
        firefox_profiles_listed = [bool]$firefoxProfilesListed
        firefox_profile_count = $firefoxProfileCount
        firefox_extensions_json_profile_count = $firefoxExtensionsJsonProfileCount
    }
    baseline_scan = [ordered]@{
        record_type_counts = $recordTypeCounts
        summary_present = $null -ne $summary
        summary_status = Get-JsonProperty $summary "status"
        summary_profile = Get-JsonProperty $summary "profile"
        roots_in_summary = if ($summary -and (Get-JsonProperty $summary "roots")) { @((Get-JsonProperty $summary "roots")).Count } else { 0 }
        files_considered = Get-JsonProperty $summary "files_considered"
        package_records = Get-JsonProperty $summary "package_records_emitted"
        findings = Get-JsonProperty $summary "findings_emitted"
        duplicates = Get-JsonProperty $summary "duplicates"
        diagnostics = Get-JsonProperty $summary "diagnostics_count"
        timed_out = Get-JsonProperty $summary "timed_out"
        error_present = if ($summary) { -not [string]::IsNullOrWhiteSpace([string](Get-JsonProperty $summary "error")) } else { $null }
    }
    failures = @($failures)
}

$redacted | ConvertTo-Json -Depth 8 | Set-Content -Path $summaryPath -Encoding UTF8

if (-not $KeepBinary -and (Test-Path $exePath)) {
    Remove-Item $exePath
}

Write-Host "Redacted smoke summary: $summaryPath"
if ($failures.Count -gt 0) {
    Write-Error ("Windows smoke failed: " + ($failures -join "; "))
    exit 1
}
exit 0
