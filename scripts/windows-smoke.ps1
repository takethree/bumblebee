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
$httpFixtureRoot = Join-Path $EvidenceRoot "http-fixture-project"
$httpReceivedOut = Join-Path $EvidenceRoot "http-received.ndjson"
$httpReceiverResult = Join-Path $EvidenceRoot "http-receiver-result.json"

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

function New-SmokePackageFixture {
    param([Parameter(Mandatory = $true)][string]$Root)

    New-Item -ItemType Directory -Force -Path $Root | Out-Null
    $packageLock = Join-Path $Root "package-lock.json"
    $body = @'
{
  "lockfileVersion": 3,
  "packages": {
    "": {"name":"windows-http-smoke","version":"1.0.0"},
    "node_modules/lodash": {"version":"4.17.21"}
  }
}
'@
    [System.IO.File]::WriteAllText($packageLock, $body, [System.Text.UTF8Encoding]::new($false))
}

function Get-FreeTcpPort {
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
    $listener.Start()
    try {
        return ([System.Net.IPEndPoint]$listener.LocalEndpoint).Port
    } finally {
        $listener.Stop()
    }
}

function Start-SmokeHttpReceiver {
    param(
        [Parameter(Mandatory = $true)][int]$Port,
        [Parameter(Mandatory = $true)][string]$BearerToken,
        [Parameter(Mandatory = $true)][string]$OutputPath,
        [Parameter(Mandatory = $true)][string]$ResultPath,
        [Parameter(Mandatory = $true)][string]$ReadyPath,
        [Parameter(Mandatory = $true)][int]$TimeoutSeconds
    )

    Start-Job -ArgumentList $Port, $BearerToken, $OutputPath, $ResultPath, $ReadyPath, $TimeoutSeconds -ScriptBlock {
        param($Port, $BearerToken, $OutputPath, $ResultPath, $ReadyPath, $TimeoutSeconds)

        Set-StrictMode -Version 3.0
        $ErrorActionPreference = "Stop"
        $requestCount = 0
        $authFailures = 0
        $parseFailures = 0
        $sawSummary = $false
        $errorText = ""
        $listener = $null

        function Write-ReceiverResult {
            param([string]$ErrorValue = "")
            [ordered]@{
                request_count = $requestCount
                auth_failures = $authFailures
                parse_failures = $parseFailures
                saw_summary = $sawSummary
                error_present = -not [string]::IsNullOrWhiteSpace($ErrorValue)
            } | ConvertTo-Json -Depth 5 | Set-Content -Path $ResultPath -Encoding UTF8
        }

        try {
            if (Test-Path $OutputPath) {
                Remove-Item -LiteralPath $OutputPath -Force
            }
            $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, $Port)
            $listener.Start()
            "ready" | Set-Content -Path $ReadyPath -Encoding ASCII
            $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)

            while ([DateTime]::UtcNow -lt $deadline -and -not $sawSummary) {
                if (-not $listener.Pending()) {
                    Start-Sleep -Milliseconds 50
                    continue
                }

                $client = $listener.AcceptTcpClient()
                try {
                    $stream = $client.GetStream()
                    $headerBytes = New-Object System.Collections.Generic.List[byte]
                    $window = New-Object System.Collections.Generic.Queue[byte]
                    while ($true) {
                        $b = $stream.ReadByte()
                        if ($b -lt 0) {
                            break
                        }
                        $headerBytes.Add([byte]$b)
                        $window.Enqueue([byte]$b)
                        while ($window.Count -gt 4) {
                            [void]$window.Dequeue()
                        }
                        if ($window.Count -eq 4) {
                            $arr = $window.ToArray()
                            if ($arr[0] -eq 13 -and $arr[1] -eq 10 -and $arr[2] -eq 13 -and $arr[3] -eq 10) {
                                break
                            }
                        }
                    }

                    $headerText = [System.Text.Encoding]::ASCII.GetString($headerBytes.ToArray())
                    $headers = @{}
                    foreach ($line in ($headerText -split "`r`n")) {
                        $idx = $line.IndexOf(":")
                        if ($idx -gt 0) {
                            $headers[$line.Substring(0, $idx).Trim().ToLowerInvariant()] = $line.Substring($idx + 1).Trim()
                        }
                    }

                    $length = 0
                    if ($headers.ContainsKey("content-length")) {
                        [void][int]::TryParse([string]$headers["content-length"], [ref]$length)
                    }
                    $bodyBytes = New-Object byte[] $length
                    $offset = 0
                    while ($offset -lt $length) {
                        $read = $stream.Read($bodyBytes, $offset, $length - $offset)
                        if ($read -le 0) {
                            break
                        }
                        $offset += $read
                    }

                    $requestCount++
                    $wantAuth = "Bearer $BearerToken"
                    if (-not $headers.ContainsKey("authorization") -or $headers["authorization"] -ne $wantAuth) {
                        $authFailures++
                    }
                    $body = [System.Text.Encoding]::UTF8.GetString($bodyBytes, 0, $offset)
                    if (-not [string]::IsNullOrWhiteSpace($body)) {
                        Add-Content -Path $OutputPath -Value $body -Encoding UTF8
                        if ($body.Contains('"record_type":"scan_summary"') -or $body.Contains("scan_summary")) {
                            $sawSummary = $true
                        }
                    }

                    $responseBody = [System.Text.Encoding]::UTF8.GetBytes("ok")
                    $responseHeader = "HTTP/1.1 200 OK`r`nContent-Length: $($responseBody.Length)`r`nConnection: close`r`n`r`n"
                    $responseHeaderBytes = [System.Text.Encoding]::ASCII.GetBytes($responseHeader)
                    $stream.Write($responseHeaderBytes, 0, $responseHeaderBytes.Length)
                    $stream.Write($responseBody, 0, $responseBody.Length)
                } catch {
                    $parseFailures++
                } finally {
                    if ($client) {
                        $client.Close()
                    }
                }
            }
        } catch {
            $errorText = $_.Exception.Message
        } finally {
            if ($listener) {
                $listener.Stop()
            }
            Write-ReceiverResult $errorText
        }
    }
}

function Wait-ForPath {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][int]$TimeoutSeconds
    )
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    while ([DateTime]::UtcNow -lt $deadline) {
        if (Test-Path -LiteralPath $Path) {
            return $true
        }
        Start-Sleep -Milliseconds 100
    }
    return $false
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

    New-SmokePackageFixture $httpFixtureRoot
    $httpReadyPath = Join-Path $EvidenceRoot "http-receiver.ready"
    $httpPort = Get-FreeTcpPort
    $httpToken = "smoke-" + ([guid]::NewGuid().ToString("N"))
    $receiverJob = Start-SmokeHttpReceiver `
        -Port $httpPort `
        -BearerToken $httpToken `
        -OutputPath $httpReceivedOut `
        -ResultPath $httpReceiverResult `
        -ReadyPath $httpReadyPath `
        -TimeoutSeconds $MaxDurationSeconds

    if (-not (Wait-ForPath $httpReadyPath 10)) {
        Add-CommandResult $commands "scan_project_http" 1
        $failures.Add("http smoke receiver did not become ready")
        Stop-Job -Job $receiverJob -ErrorAction SilentlyContinue
        Remove-Job -Job $receiverJob -Force -ErrorAction SilentlyContinue
    } else {
        $oldSmokeToken = $env:BUMBLEBEE_SMOKE_HTTP_TOKEN
        $env:BUMBLEBEE_SMOKE_HTTP_TOKEN = $httpToken
        try {
            $httpScanCode = Invoke-Captured $exePath @(
                "scan",
                "--profile", "project",
                "--root", $httpFixtureRoot,
                "--output", "http",
                "--http-url", "http://127.0.0.1:$httpPort/ingest",
                "--http-auth", "bearer",
                "--http-token-env", "BUMBLEBEE_SMOKE_HTTP_TOKEN",
                "--http-batch-size", "1",
                "--max-duration", $duration
            ) (Join-Path $EvidenceRoot "scan-http.stdout.txt") (Join-Path $EvidenceRoot "scan-http.stderr.txt")
            Add-CommandResult $commands "scan_project_http" $httpScanCode
            if ($httpScanCode -ne 0) { $failures.Add("project HTTP scan failed") }
        } finally {
            if ($null -eq $oldSmokeToken) {
                Remove-Item Env:BUMBLEBEE_SMOKE_HTTP_TOKEN -ErrorAction SilentlyContinue
            } else {
                $env:BUMBLEBEE_SMOKE_HTTP_TOKEN = $oldSmokeToken
            }
        }

        $receiverDone = Wait-Job -Job $receiverJob -Timeout 10
        if ($null -eq $receiverDone) {
            $failures.Add("http smoke receiver did not finish after scan")
            Stop-Job -Job $receiverJob -ErrorAction SilentlyContinue
        }
        Receive-Job -Job $receiverJob -ErrorAction SilentlyContinue | Out-Null
        Remove-Job -Job $receiverJob -Force -ErrorAction SilentlyContinue
    }
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

$httpRecords = Read-JsonLines $httpReceivedOut
$httpRecordTypeCounts = [ordered]@{}
$httpSummary = $null
foreach ($record in $httpRecords) {
    $recordType = [string]$record.record_type
    if ([string]::IsNullOrWhiteSpace($recordType)) {
        continue
    }
    if (-not $httpRecordTypeCounts.Contains($recordType)) {
        $httpRecordTypeCounts[$recordType] = 0
    }
    $httpRecordTypeCounts[$recordType]++
    if ($recordType -eq "scan_summary") {
        $httpSummary = $record
    }
}
$httpReceiver = $null
if (Test-Path $httpReceiverResult) {
    $httpReceiver = Get-Content -Raw -Path $httpReceiverResult | ConvertFrom-Json
}
$httpRequestCount = Get-JsonProperty $httpReceiver "request_count"
$httpAuthFailures = Get-JsonProperty $httpReceiver "auth_failures"
$httpParseFailures = Get-JsonProperty $httpReceiver "parse_failures"
$httpReceiverSawSummary = Get-JsonProperty $httpReceiver "saw_summary"
$httpReceiverError = Get-JsonProperty $httpReceiver "error_present"
if (-not $httpRecordTypeCounts.Contains("package")) {
    $httpPackageRecords = 0
} else {
    $httpPackageRecords = $httpRecordTypeCounts["package"]
}

if ($commands.Contains("scan_project_http") -and $commands["scan_project_http"].exit_code -eq 0) {
    if ($null -eq $httpReceiver) {
        $failures.Add("http smoke receiver result was missing")
    }
    if ($null -eq $httpRequestCount -or [int]$httpRequestCount -le 0) {
        $failures.Add("http smoke receiver got no requests")
    }
    if ($null -ne $httpAuthFailures -and [int]$httpAuthFailures -gt 0) {
        $failures.Add("http smoke receiver saw bearer auth failures")
    }
    if ($null -ne $httpParseFailures -and [int]$httpParseFailures -gt 0) {
        $failures.Add("http smoke receiver saw parse failures")
    }
    if ($httpReceiverError) {
        $failures.Add("http smoke receiver reported an internal error")
    }
    if ($httpPackageRecords -le 0) {
        $failures.Add("http smoke receiver got no package records")
    }
    if ($null -eq $httpSummary) {
        $failures.Add("http smoke receiver got no scan_summary")
    } elseif ($httpSummary.status -ne "complete") {
        $failures.Add("http scan_summary status was not complete")
    }
    if ($httpReceiverSawSummary -eq $false) {
        $failures.Add("http smoke receiver did not observe scan_summary")
    }
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
    http_sink = [ordered]@{
        request_count = $httpRequestCount
        auth_failures = $httpAuthFailures
        parse_failures = $httpParseFailures
        receiver_error_present = $httpReceiverError
        receiver_saw_summary = $httpReceiverSawSummary
        record_type_counts = $httpRecordTypeCounts
        package_records = $httpPackageRecords
        summary_present = $null -ne $httpSummary
        summary_status = Get-JsonProperty $httpSummary "status"
        summary_profile = Get-JsonProperty $httpSummary "profile"
        summary_http_batches_attempted = Get-JsonProperty $httpSummary "http_batches_attempted"
        summary_http_batches_succeeded = Get-JsonProperty $httpSummary "http_batches_succeeded"
        summary_http_batches_failed = Get-JsonProperty $httpSummary "http_batches_failed"
        summary_http_last_status = Get-JsonProperty $httpSummary "http_last_status"
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
