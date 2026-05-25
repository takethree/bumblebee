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
$nugetFixtureRoot = Join-Path $EvidenceRoot "nuget fixture project"
$nugetScanOut = Join-Path $EvidenceRoot "nuget-project.ndjson"
$nugetScanErr = Join-Path $EvidenceRoot "nuget-project.stderr.txt"
$powershellFixtureRoot = Join-Path $EvidenceRoot "PowerShell fixture modules"
$powershellScanOut = Join-Path $EvidenceRoot "powershell-modules.ndjson"
$powershellScanErr = Join-Path $EvidenceRoot "powershell-modules.stderr.txt"
$browserFixtureRoot = Join-Path $EvidenceRoot "browser fixture roots"
$browserScanOut = Join-Path $EvidenceRoot "browser-extensions.ndjson"
$browserScanErr = Join-Path $EvidenceRoot "browser-extensions.stderr.txt"
$strictParityFixtureHome = Join-Path $EvidenceRoot "strict-parity-home"
$strictParityRootsOut = Join-Path $EvidenceRoot "strict-parity-roots.tsv"
$strictParityRootsErr = Join-Path $EvidenceRoot "strict-parity-roots.stderr.txt"
$strictParityScanOut = Join-Path $EvidenceRoot "strict-parity-scan.ndjson"
$strictParityScanErr = Join-Path $EvidenceRoot "strict-parity-scan.stderr.txt"

function ConvertTo-WindowsCommandLineArgument {
    param([AllowNull()][string]$Argument)

    if ($null -eq $Argument) {
        return '""'
    }
    if ($Argument -notmatch '[\s"]') {
        return $Argument
    }

    $out = [System.Text.StringBuilder]::new()
    [void]$out.Append('"')
    $backslashes = 0
    foreach ($ch in $Argument.ToCharArray()) {
        if ($ch -eq '\') {
            $backslashes++
            continue
        }
        if ($ch -eq '"') {
            [void]$out.Append(('\' * (($backslashes * 2) + 1)))
            [void]$out.Append('"')
            $backslashes = 0
            continue
        }
        if ($backslashes -gt 0) {
            [void]$out.Append(('\' * $backslashes))
            $backslashes = 0
        }
        [void]$out.Append($ch)
    }
    if ($backslashes -gt 0) {
        [void]$out.Append(('\' * ($backslashes * 2)))
    }
    [void]$out.Append('"')
    return $out.ToString()
}

function ConvertTo-WindowsCommandLine {
    param([Parameter(Mandatory = $true)][string[]]$Arguments)

    return (($Arguments | ForEach-Object { ConvertTo-WindowsCommandLineArgument $_ }) -join " ")
}

function Invoke-Captured {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$StdoutPath,
        [Parameter(Mandatory = $true)][string]$StderrPath,
        [string]$WorkingDirectory = $repoRoot
    )

    $argumentLine = ConvertTo-WindowsCommandLine $Arguments
    $proc = Start-Process `
        -FilePath $FilePath `
        -ArgumentList $argumentLine `
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

function New-SmokeNuGetFixture {
    param([Parameter(Mandatory = $true)][string]$Root)

    New-Item -ItemType Directory -Force -Path $Root | Out-Null
    $packagesConfig = Join-Path $Root "packages.config"
    $packagesConfigBody = @'
<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Newtonsoft.Json" version="13.0.3" targetFramework="net472" />
  <package id="Serilog" version="3.1.1" />
  <package id="NoVersion" />
</packages>
'@
    [System.IO.File]::WriteAllText($packagesConfig, $packagesConfigBody, [System.Text.UTF8Encoding]::new($false))

    $lockfile = Join-Path $Root "packages.lock.json"
    $lockfileBody = @'
{
  "version": 1,
  "dependencies": {
    ".NETFramework,Version=v4.7.2": {
      "Newtonsoft.Json": {
        "type": "Direct",
        "requested": "[13.0.3, )",
        "resolved": "13.0.3",
        "contentHash": "abc"
      },
      "Serilog": {
        "type": "Transitive",
        "requested": "[3.0.0, )",
        "resolved": "3.1.1",
        "contentHash": "def"
      },
      "Local.Project": {
        "type": "Project"
      },
      "NoResolved": {
        "type": "Direct"
      }
    },
    "net8.0": {
      "Newtonsoft.Json": {
        "type": "Direct",
        "requested": "[13.0.3, )",
        "resolved": "13.0.3",
        "contentHash": "abc"
      },
      "Different.Version": {
        "type": "Transitive",
        "requested": "[2.0.0, )",
        "resolved": "2.0.0",
        "contentHash": "ghi"
      }
    }
  }
}
'@
    [System.IO.File]::WriteAllText($lockfile, $lockfileBody, [System.Text.UTF8Encoding]::new($false))
}

function New-SmokePowerShellFixture {
    param([Parameter(Mandatory = $true)][string]$Root)

    $pesterRoot = Join-Path $Root "Pester\5.7.1"
    New-Item -ItemType Directory -Force -Path $pesterRoot | Out-Null
    $pesterManifest = Join-Path $pesterRoot "Pester.psd1"
    $pesterBody = @'
@{
  RootModule = 'Pester.psm1'
  ModuleVersion = '5.7.1'
  GUID = 'a699dea5-2c73-4616-a270-1f7abb777e71'
  Author = 'PowerShell Team'
}
'@
    [System.IO.File]::WriteAllText($pesterManifest, $pesterBody, [System.Text.UTF8Encoding]::new($false))

    $missingRoot = Join-Path $Root "NoVersion"
    New-Item -ItemType Directory -Force -Path $missingRoot | Out-Null
    [System.IO.File]::WriteAllText((Join-Path $missingRoot "NoVersion.psd1"), "@{ RootModule = 'NoVersion.psm1' }", [System.Text.UTF8Encoding]::new($false))

    $expressionRoot = Join-Path $Root "ExpressionVersion"
    New-Item -ItemType Directory -Force -Path $expressionRoot | Out-Null
    [System.IO.File]::WriteAllText((Join-Path $expressionRoot "ExpressionVersion.psd1"), "@{ ModuleVersion = (Get-Date) }", [System.Text.UTF8Encoding]::new($false))
}

function New-SmokeBrowserFixture {
    param([Parameter(Mandatory = $true)][string]$Root)

    $chromiumCases = @(
        @{ Browser = "BraveSoftware\Brave-Browser"; Name = "Smoke Brave Extension"; Id = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"; Version = "1.0.0" },
        @{ Browser = "Chromium"; Name = "Smoke Chromium Extension"; Id = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"; Version = "2.0.0" },
        @{ Browser = "Vivaldi"; Name = "Smoke Vivaldi Extension"; Id = "cccccccccccccccccccccccccccccccc"; Version = "3.0.0" }
    )
    foreach ($case in $chromiumCases) {
        $manifestDir = Join-Path $Root "$($case.Browser)\User Data\Default\Extensions\$($case.Id)\$($case.Version)"
        New-Item -ItemType Directory -Force -Path $manifestDir | Out-Null
        $manifest = @{
            manifest_version = 3
            name = $case.Name
            version = $case.Version
        } | ConvertTo-Json -Compress
        [System.IO.File]::WriteAllText((Join-Path $manifestDir "manifest.json"), $manifest, [System.Text.UTF8Encoding]::new($false))
    }

    $firefoxCases = @(
        @{ Path = "LibreWolf\Profiles\abcd.default"; Name = "Smoke LibreWolf Addon"; Id = "librewolf-smoke@example.com"; Version = "4.0.0" },
        @{ Path = "Waterfox\Waterfox\Profiles\abcd.default"; Name = "Smoke Waterfox Nested Addon"; Id = "waterfox-nested-smoke@example.com"; Version = "5.0.0" },
        @{ Path = "Waterfox\Profiles\abcd.default"; Name = "Smoke Waterfox Legacy Addon"; Id = "waterfox-legacy-smoke@example.com"; Version = "6.0.0" }
    )
    foreach ($case in $firefoxCases) {
        $profileDir = Join-Path $Root $case.Path
        New-Item -ItemType Directory -Force -Path $profileDir | Out-Null
        $extensions = @{
            addons = @(
                @{
                    id = $case.Id
                    version = $case.Version
                    type = "extension"
                    active = $true
                    defaultLocale = @{ name = $case.Name }
                }
            )
        } | ConvertTo-Json -Depth 5 -Compress
        [System.IO.File]::WriteAllText((Join-Path $profileDir "extensions.json"), $extensions, [System.Text.UTF8Encoding]::new($false))
    }
}

function New-SmokeStrictParityRootFixture {
    param([Parameter(Mandatory = $true)][string]$FixtureHome)

    $appData = Join-Path $FixtureHome "AppData\Roaming"
    $localAppData = Join-Path $FixtureHome "AppData\Local"
    $npmRoot = Join-Path $appData "npm\node_modules\smoke-npm-root"
    New-Item -ItemType Directory -Force -Path $npmRoot | Out-Null
    [System.IO.File]::WriteAllText((Join-Path $npmRoot "package.json"), '{"name":"smoke-npm-root","version":"1.0.0"}', [System.Text.UTF8Encoding]::new($false))

    $pythonRoot = Join-Path $appData "Python\Python311\site-packages\SmokePythonRoot-1.0.0.dist-info"
    New-Item -ItemType Directory -Force -Path $pythonRoot | Out-Null
    [System.IO.File]::WriteAllText((Join-Path $pythonRoot "METADATA"), "Metadata-Version: 2.1`nName: SmokePythonRoot`nVersion: 1.0.0`n`n", [System.Text.UTF8Encoding]::new($false))

    $pipxCases = @(
        @{ Root = (Join-Path $FixtureHome "pipx\venvs\smoke-pipx\Lib\site-packages\SmokePipxRoot-2.0.0.dist-info"); Name = "SmokePipxRoot"; Version = "2.0.0" },
        @{ Root = (Join-Path $localAppData "pipx\venvs\smoke-pipx-local\Lib\site-packages\SmokePipxLocalRoot-3.0.0.dist-info"); Name = "SmokePipxLocalRoot"; Version = "3.0.0" },
        @{ Root = (Join-Path $FixtureHome ".local\pipx\venvs\smoke-pipx-legacy\Lib\site-packages\SmokePipxLegacyRoot-4.0.0.dist-info"); Name = "SmokePipxLegacyRoot"; Version = "4.0.0" }
    )
    foreach ($case in $pipxCases) {
        New-Item -ItemType Directory -Force -Path $case.Root | Out-Null
        [System.IO.File]::WriteAllText((Join-Path $case.Root "METADATA"), "Metadata-Version: 2.1`nName: $($case.Name)`nVersion: $($case.Version)`n`n", [System.Text.UTF8Encoding]::new($false))
    }
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

    New-SmokeNuGetFixture $nugetFixtureRoot
    $nugetScanCode = Invoke-Captured $exePath @(
        "scan",
        "--profile", "project",
        "--root", $nugetFixtureRoot,
        "--ecosystem", "nuget",
        "--max-duration", $duration
    ) $nugetScanOut $nugetScanErr
    Add-CommandResult $commands "scan_project_nuget" $nugetScanCode
    if ($nugetScanCode -ne 0) { $failures.Add("project NuGet scan failed") }

    New-SmokePowerShellFixture $powershellFixtureRoot
    $powershellScanCode = Invoke-Captured $exePath @(
        "scan",
        "--profile", "project",
        "--root", $powershellFixtureRoot,
        "--ecosystem", "powershell-module",
        "--max-duration", $duration
    ) $powershellScanOut $powershellScanErr
    Add-CommandResult $commands "scan_project_powershell" $powershellScanCode
    if ($powershellScanCode -ne 0) { $failures.Add("project PowerShell module scan failed") }

    New-SmokeBrowserFixture $browserFixtureRoot
    $browserScanCode = Invoke-Captured $exePath @(
        "scan",
        "--profile", "project",
        "--root", $browserFixtureRoot,
        "--ecosystem", "browser-extension",
        "--max-duration", $duration
    ) $browserScanOut $browserScanErr
    Add-CommandResult $commands "scan_project_browser_extensions" $browserScanCode
    if ($browserScanCode -ne 0) { $failures.Add("project browser extension scan failed") }

    New-SmokeStrictParityRootFixture $strictParityFixtureHome
    $oldUserProfile = $env:USERPROFILE
    $oldHomeDrive = $env:HOMEDRIVE
    $oldHomePath = $env:HOMEPATH
    $oldAppData = $env:APPDATA
    $oldLocalAppData = $env:LOCALAPPDATA
    try {
        $env:USERPROFILE = $strictParityFixtureHome
        $env:HOMEDRIVE = Split-Path -Qualifier $strictParityFixtureHome
        $env:HOMEPATH = Split-Path -NoQualifier $strictParityFixtureHome
        $env:APPDATA = Join-Path $strictParityFixtureHome "AppData\Roaming"
        $env:LOCALAPPDATA = Join-Path $strictParityFixtureHome "AppData\Local"

        $strictRootsCode = Invoke-Captured $exePath @(
            "roots",
            "--profile", "baseline"
        ) $strictParityRootsOut $strictParityRootsErr
        Add-CommandResult $commands "roots_strict_parity_fixture" $strictRootsCode
        if ($strictRootsCode -ne 0) { $failures.Add("strict-parity fixture roots failed") }

        $strictScanCode = Invoke-Captured $exePath @(
            "scan",
            "--profile", "baseline",
            "--ecosystem", "npm,pypi",
            "--max-duration", $duration
        ) $strictParityScanOut $strictParityScanErr
        Add-CommandResult $commands "scan_strict_parity_fixture" $strictScanCode
        if ($strictScanCode -ne 0) { $failures.Add("strict-parity fixture scan failed") }
    } finally {
        if ($null -eq $oldUserProfile) { Remove-Item Env:USERPROFILE -ErrorAction SilentlyContinue } else { $env:USERPROFILE = $oldUserProfile }
        if ($null -eq $oldHomeDrive) { Remove-Item Env:HOMEDRIVE -ErrorAction SilentlyContinue } else { $env:HOMEDRIVE = $oldHomeDrive }
        if ($null -eq $oldHomePath) { Remove-Item Env:HOMEPATH -ErrorAction SilentlyContinue } else { $env:HOMEPATH = $oldHomePath }
        if ($null -eq $oldAppData) { Remove-Item Env:APPDATA -ErrorAction SilentlyContinue } else { $env:APPDATA = $oldAppData }
        if ($null -eq $oldLocalAppData) { Remove-Item Env:LOCALAPPDATA -ErrorAction SilentlyContinue } else { $env:LOCALAPPDATA = $oldLocalAppData }
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
$braveDefaultExtensions = if ($env:LOCALAPPDATA) { Join-Path $env:LOCALAPPDATA "BraveSoftware\Brave-Browser\User Data\Default\Extensions" } else { "" }
$chromiumDefaultExtensions = if ($env:LOCALAPPDATA) { Join-Path $env:LOCALAPPDATA "Chromium\User Data\Default\Extensions" } else { "" }
$edgeDefaultExtensions = if ($env:LOCALAPPDATA) { Join-Path $env:LOCALAPPDATA "Microsoft\Edge\User Data\Default\Extensions" } else { "" }
$vivaldiDefaultExtensions = if ($env:LOCALAPPDATA) { Join-Path $env:LOCALAPPDATA "Vivaldi\User Data\Default\Extensions" } else { "" }
$firefoxProfiles = if ($env:APPDATA) { Join-Path $env:APPDATA "Mozilla\Firefox\Profiles" } else { "" }
$librewolfProfiles = if ($env:APPDATA) { Join-Path $env:APPDATA "LibreWolf\Profiles" } else { "" }
$waterfoxProfiles = if ($env:APPDATA) { Join-Path $env:APPDATA "Waterfox\Waterfox\Profiles" } else { "" }
$waterfoxLegacyProfiles = if ($env:APPDATA) { Join-Path $env:APPDATA "Waterfox\Profiles" } else { "" }
$goRootExists = if ($goRoot) { Test-Path $goRoot } else { $false }
$goRootListed = if ($goRoot) { @($roots | Where-Object { $_.Path -eq $goRoot }).Count -gt 0 } else { $false }
$appDataClaudeExists = if ($appDataClaude) { Test-Path $appDataClaude } else { $false }
$appDataClaudeListed = if ($appDataClaude) { @($roots | Where-Object { $_.Path -eq $appDataClaude }).Count -gt 0 } else { $false }
$msixClaudeExists = if ($msixClaude) { Test-Path $msixClaude } else { $false }
$msixClaudeListed = if ($msixClaude) { @($roots | Where-Object { $_.Path -eq $msixClaude }).Count -gt 0 } else { $false }
$chromeDefaultExtensionsExists = if ($chromeDefaultExtensions) { Test-Path $chromeDefaultExtensions } else { $false }
$chromeDefaultExtensionsListed = if ($chromeDefaultExtensions) { @($roots | Where-Object { $_.Path -eq $chromeDefaultExtensions }).Count -gt 0 } else { $false }
$braveDefaultExtensionsExists = if ($braveDefaultExtensions) { Test-Path $braveDefaultExtensions } else { $false }
$braveDefaultExtensionsListed = if ($braveDefaultExtensions) { @($roots | Where-Object { $_.Path -eq $braveDefaultExtensions }).Count -gt 0 } else { $false }
$chromiumDefaultExtensionsExists = if ($chromiumDefaultExtensions) { Test-Path $chromiumDefaultExtensions } else { $false }
$chromiumDefaultExtensionsListed = if ($chromiumDefaultExtensions) { @($roots | Where-Object { $_.Path -eq $chromiumDefaultExtensions }).Count -gt 0 } else { $false }
$edgeDefaultExtensionsExists = if ($edgeDefaultExtensions) { Test-Path $edgeDefaultExtensions } else { $false }
$edgeDefaultExtensionsListed = if ($edgeDefaultExtensions) { @($roots | Where-Object { $_.Path -eq $edgeDefaultExtensions }).Count -gt 0 } else { $false }
$vivaldiDefaultExtensionsExists = if ($vivaldiDefaultExtensions) { Test-Path $vivaldiDefaultExtensions } else { $false }
$vivaldiDefaultExtensionsListed = if ($vivaldiDefaultExtensions) { @($roots | Where-Object { $_.Path -eq $vivaldiDefaultExtensions }).Count -gt 0 } else { $false }
$firefoxProfilesExists = if ($firefoxProfiles) { Test-Path $firefoxProfiles } else { $false }
$firefoxProfilesListed = if ($firefoxProfiles) { @($roots | Where-Object { $_.Path -eq $firefoxProfiles }).Count -gt 0 } else { $false }
$librewolfProfilesExists = if ($librewolfProfiles) { Test-Path $librewolfProfiles } else { $false }
$librewolfProfilesListed = if ($librewolfProfiles) { @($roots | Where-Object { $_.Path -eq $librewolfProfiles }).Count -gt 0 } else { $false }
$waterfoxProfilesExists = if ($waterfoxProfiles) { Test-Path $waterfoxProfiles } else { $false }
$waterfoxProfilesListed = if ($waterfoxProfiles) { @($roots | Where-Object { $_.Path -eq $waterfoxProfiles }).Count -gt 0 } else { $false }
$waterfoxLegacyProfilesExists = if ($waterfoxLegacyProfiles) { Test-Path $waterfoxLegacyProfiles } else { $false }
$waterfoxLegacyProfilesListed = if ($waterfoxLegacyProfiles) { @($roots | Where-Object { $_.Path -eq $waterfoxLegacyProfiles }).Count -gt 0 } else { $false }
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

$nugetRecords = Read-JsonLines $nugetScanOut
$nugetRecordTypeCounts = [ordered]@{}
$nugetSourceTypeCounts = [ordered]@{}
$nugetSummary = $null
$nugetPackages = New-Object System.Collections.Generic.List[object]
foreach ($record in $nugetRecords) {
    $recordType = [string]$record.record_type
    if ([string]::IsNullOrWhiteSpace($recordType)) {
        continue
    }
    if (-not $nugetRecordTypeCounts.Contains($recordType)) {
        $nugetRecordTypeCounts[$recordType] = 0
    }
    $nugetRecordTypeCounts[$recordType]++
    if ($recordType -eq "package") {
        $nugetPackages.Add($record)
        $sourceType = [string]$record.source_type
        if (-not [string]::IsNullOrWhiteSpace($sourceType)) {
            if (-not $nugetSourceTypeCounts.Contains($sourceType)) {
                $nugetSourceTypeCounts[$sourceType] = 0
            }
            $nugetSourceTypeCounts[$sourceType]++
        }
    }
    if ($recordType -eq "scan_summary") {
        $nugetSummary = $record
    }
}

$nugetRequiredFields = @("ecosystem", "package_name", "normalized_name", "version", "package_manager", "source_type", "source_file", "project_path", "root_kind", "confidence")
$nugetMissingRequiredFields = [ordered]@{}
foreach ($field in $nugetRequiredFields) {
    $missing = @($nugetPackages | Where-Object { $null -eq (Get-JsonProperty $_ $field) -or [string]::IsNullOrWhiteSpace([string](Get-JsonProperty $_ $field)) }).Count
    $nugetMissingRequiredFields[$field] = $missing
}
$nugetRequestedSpecCount = @($nugetPackages | Where-Object { -not [string]::IsNullOrWhiteSpace([string](Get-JsonProperty $_ "requested_spec")) }).Count
$nugetDirectTrueCount = @($nugetPackages | Where-Object { (Get-JsonProperty $_ "direct_dependency") -eq $true }).Count
$nugetDirectFalseCount = @($nugetPackages | Where-Object { (Get-JsonProperty $_ "direct_dependency") -eq $false }).Count
$nugetDirectEmptyCount = @($nugetPackages | Where-Object { $null -eq (Get-JsonProperty $_ "direct_dependency") }).Count
$nugetTransitiveScopeCount = @($nugetPackages | Where-Object { (Get-JsonProperty $_ "install_scope") -eq "transitive" }).Count
$nugetProjectRootCount = @($nugetPackages | Where-Object { (Get-JsonProperty $_ "root_kind") -eq "project_root" }).Count
$nugetProjectReferenceEmitted = @($nugetPackages | Where-Object { (Get-JsonProperty $_ "package_name") -eq "Local.Project" }).Count -gt 0
$nugetMissingVersionEmitted = @($nugetPackages | Where-Object { (Get-JsonProperty $_ "package_name") -eq "NoVersion" }).Count -gt 0
$nugetMissingResolvedEmitted = @($nugetPackages | Where-Object { (Get-JsonProperty $_ "package_name") -eq "NoResolved" }).Count -gt 0

if ($commands.Contains("scan_project_nuget") -and $commands["scan_project_nuget"].exit_code -eq 0) {
    if ($nugetPackages.Count -ne 5) {
        $failures.Add("NuGet smoke emitted $($nugetPackages.Count) package records, want 5")
    }
    if ($null -eq $nugetSummary) {
        $failures.Add("NuGet smoke scan_summary was missing")
    } elseif ($nugetSummary.status -ne "complete") {
        $failures.Add("NuGet smoke scan_summary status was not complete")
    }
    foreach ($field in $nugetRequiredFields) {
        if ($nugetMissingRequiredFields[$field] -ne 0) {
            $failures.Add("NuGet smoke missing required field $field")
        }
    }
    if (-not $nugetSourceTypeCounts.Contains("nuget-packages-config") -or $nugetSourceTypeCounts["nuget-packages-config"] -ne 2) {
        $failures.Add("NuGet smoke did not emit 2 packages.config records")
    }
    if (-not $nugetSourceTypeCounts.Contains("nuget-lockfile") -or $nugetSourceTypeCounts["nuget-lockfile"] -ne 3) {
        $failures.Add("NuGet smoke did not emit 3 lockfile records")
    }
    if ($nugetRequestedSpecCount -ne 3) {
        $failures.Add("NuGet smoke requested_spec count was $nugetRequestedSpecCount, want 3")
    }
    if ($nugetDirectTrueCount -ne 1 -or $nugetDirectFalseCount -ne 2 -or $nugetDirectEmptyCount -ne 2 -or $nugetTransitiveScopeCount -ne 2) {
        $failures.Add("NuGet smoke direct/transitive counts were unexpected")
    }
    if ($nugetProjectRootCount -ne $nugetPackages.Count) {
        $failures.Add("NuGet smoke did not stamp all packages as project_root")
    }
    if ($nugetProjectReferenceEmitted -or $nugetMissingVersionEmitted -or $nugetMissingResolvedEmitted) {
        $failures.Add("NuGet smoke emitted an expected-skipped entry")
    }
}

$powershellRecords = Read-JsonLines $powershellScanOut
$powershellRecordTypeCounts = [ordered]@{}
$powershellSourceTypeCounts = [ordered]@{}
$powershellSummary = $null
$powershellPackages = New-Object System.Collections.Generic.List[object]
foreach ($record in $powershellRecords) {
    $recordType = [string]$record.record_type
    if ([string]::IsNullOrWhiteSpace($recordType)) {
        continue
    }
    if (-not $powershellRecordTypeCounts.Contains($recordType)) {
        $powershellRecordTypeCounts[$recordType] = 0
    }
    $powershellRecordTypeCounts[$recordType]++
    if ($recordType -eq "package") {
        $powershellPackages.Add($record)
        $sourceType = [string]$record.source_type
        if (-not [string]::IsNullOrWhiteSpace($sourceType)) {
            if (-not $powershellSourceTypeCounts.Contains($sourceType)) {
                $powershellSourceTypeCounts[$sourceType] = 0
            }
            $powershellSourceTypeCounts[$sourceType]++
        }
    }
    if ($recordType -eq "scan_summary") {
        $powershellSummary = $record
    }
}

$powershellRequiredFields = @("ecosystem", "package_name", "normalized_name", "version", "package_manager", "source_type", "source_file", "project_path", "root_kind", "confidence")
$powershellMissingRequiredFields = [ordered]@{}
foreach ($field in $powershellRequiredFields) {
    $missing = @($powershellPackages | Where-Object { $null -eq (Get-JsonProperty $_ $field) -or [string]::IsNullOrWhiteSpace([string](Get-JsonProperty $_ $field)) }).Count
    $powershellMissingRequiredFields[$field] = $missing
}
$powershellProjectRootCount = @($powershellPackages | Where-Object { (Get-JsonProperty $_ "root_kind") -eq "project_root" }).Count
$powershellPesterEmitted = @($powershellPackages | Where-Object {
    (Get-JsonProperty $_ "package_name") -eq "Pester" -and
    (Get-JsonProperty $_ "normalized_name") -eq "pester" -and
    (Get-JsonProperty $_ "version") -eq "5.7.1" -and
    (Get-JsonProperty $_ "ecosystem") -eq "powershell-module" -and
    (Get-JsonProperty $_ "package_manager") -eq "powershell" -and
    (Get-JsonProperty $_ "source_type") -eq "powershell-module-manifest"
}).Count -eq 1
$powershellNoVersionEmitted = @($powershellPackages | Where-Object { (Get-JsonProperty $_ "package_name") -eq "NoVersion" }).Count -gt 0
$powershellExpressionVersionEmitted = @($powershellPackages | Where-Object { (Get-JsonProperty $_ "package_name") -eq "ExpressionVersion" }).Count -gt 0

if ($commands.Contains("scan_project_powershell") -and $commands["scan_project_powershell"].exit_code -eq 0) {
    if ($powershellPackages.Count -ne 1) {
        $failures.Add("PowerShell module smoke emitted $($powershellPackages.Count) package records, want 1")
    }
    if ($null -eq $powershellSummary) {
        $failures.Add("PowerShell module smoke scan_summary was missing")
    } elseif ($powershellSummary.status -ne "complete") {
        $failures.Add("PowerShell module smoke scan_summary status was not complete")
    }
    foreach ($field in $powershellRequiredFields) {
        if ($powershellMissingRequiredFields[$field] -ne 0) {
            $failures.Add("PowerShell module smoke missing required field $field")
        }
    }
    if (-not $powershellSourceTypeCounts.Contains("powershell-module-manifest") -or $powershellSourceTypeCounts["powershell-module-manifest"] -ne 1) {
        $failures.Add("PowerShell module smoke did not emit 1 manifest record")
    }
    if ($powershellProjectRootCount -ne $powershellPackages.Count) {
        $failures.Add("PowerShell module smoke did not stamp all packages as project_root")
    }
    if (-not $powershellPesterEmitted) {
        $failures.Add("PowerShell module smoke did not emit the expected Pester manifest record")
    }
    if ($powershellNoVersionEmitted -or $powershellExpressionVersionEmitted) {
        $failures.Add("PowerShell module smoke emitted an expected-skipped entry")
    }
}

$browserRecords = Read-JsonLines $browserScanOut
$browserRecordTypeCounts = [ordered]@{}
$browserSourceTypeCounts = [ordered]@{}
$browserManagerCounts = [ordered]@{}
$browserPackages = New-Object System.Collections.Generic.List[object]
$browserSummary = $null
foreach ($record in $browserRecords) {
    $recordType = [string]$record.record_type
    if ([string]::IsNullOrWhiteSpace($recordType)) {
        continue
    }
    if (-not $browserRecordTypeCounts.Contains($recordType)) {
        $browserRecordTypeCounts[$recordType] = 0
    }
    $browserRecordTypeCounts[$recordType]++
    if ($recordType -eq "package") {
        $browserPackages.Add($record)
        $sourceType = [string]$record.source_type
        if (-not [string]::IsNullOrWhiteSpace($sourceType)) {
            if (-not $browserSourceTypeCounts.Contains($sourceType)) {
                $browserSourceTypeCounts[$sourceType] = 0
            }
            $browserSourceTypeCounts[$sourceType]++
        }
        $manager = [string]$record.package_manager
        if (-not [string]::IsNullOrWhiteSpace($manager)) {
            if (-not $browserManagerCounts.Contains($manager)) {
                $browserManagerCounts[$manager] = 0
            }
            $browserManagerCounts[$manager]++
        }
    }
    if ($recordType -eq "scan_summary") {
        $browserSummary = $record
    }
}

$browserRequiredFields = @("ecosystem", "package_name", "normalized_name", "version", "package_manager", "source_type", "source_file", "project_path", "root_kind", "confidence")
$browserMissingRequiredFields = [ordered]@{}
foreach ($field in $browserRequiredFields) {
    $missing = @($browserPackages | Where-Object { $null -eq (Get-JsonProperty $_ $field) -or [string]::IsNullOrWhiteSpace([string](Get-JsonProperty $_ $field)) }).Count
    $browserMissingRequiredFields[$field] = $missing
}
$browserProjectRootCount = @($browserPackages | Where-Object { (Get-JsonProperty $_ "root_kind") -eq "project_root" }).Count
$browserExpectedNames = @(
    "Smoke Brave Extension",
    "Smoke Chromium Extension",
    "Smoke Vivaldi Extension",
    "Smoke LibreWolf Addon",
    "Smoke Waterfox Nested Addon",
    "Smoke Waterfox Legacy Addon"
)
$browserExpectedNamesEmitted = [ordered]@{}
foreach ($name in $browserExpectedNames) {
    $browserExpectedNamesEmitted[$name] = @($browserPackages | Where-Object { (Get-JsonProperty $_ "package_name") -eq $name }).Count -eq 1
}

if ($commands.Contains("scan_project_browser_extensions") -and $commands["scan_project_browser_extensions"].exit_code -eq 0) {
    if ($browserPackages.Count -ne 6) {
        $failures.Add("browser extension smoke emitted $($browserPackages.Count) package records, want 6")
    }
    if ($null -eq $browserSummary) {
        $failures.Add("browser extension smoke scan_summary was missing")
    } elseif ($browserSummary.status -ne "complete") {
        $failures.Add("browser extension smoke scan_summary status was not complete")
    }
    foreach ($field in $browserRequiredFields) {
        if ($browserMissingRequiredFields[$field] -ne 0) {
            $failures.Add("browser extension smoke missing required field $field")
        }
    }
    if (-not $browserSourceTypeCounts.Contains("browser-extension") -or $browserSourceTypeCounts["browser-extension"] -ne 6) {
        $failures.Add("browser extension smoke did not emit 6 browser-extension source_type records")
    }
    if (-not $browserManagerCounts.Contains("chromium-extension") -or $browserManagerCounts["chromium-extension"] -ne 3) {
        $failures.Add("browser extension smoke did not emit 3 chromium-extension records")
    }
    if (-not $browserManagerCounts.Contains("firefox-extension") -or $browserManagerCounts["firefox-extension"] -ne 3) {
        $failures.Add("browser extension smoke did not emit 3 firefox-extension records")
    }
    if ($browserProjectRootCount -ne $browserPackages.Count) {
        $failures.Add("browser extension smoke did not stamp all packages as project_root")
    }
    foreach ($name in $browserExpectedNames) {
        if (-not $browserExpectedNamesEmitted[$name]) {
            $failures.Add("browser extension smoke did not emit expected package $name")
        }
    }
}

$strictParityRoots = @()
if (Test-Path $strictParityRootsOut) {
    foreach ($line in [System.IO.File]::ReadLines($strictParityRootsOut)) {
        if ([string]::IsNullOrWhiteSpace($line)) {
            continue
        }
        $parts = $line -split "`t", 2
        if ($parts.Count -ne 2) {
            $failures.Add("strict-parity roots output contained a non-TSV line")
            continue
        }
        $strictParityRoots += [pscustomobject]@{ Kind = $parts[0]; Path = $parts[1] }
    }
}

$strictParityExpectedRoots = [ordered]@{
    npm_global_modules = (Join-Path $strictParityFixtureHome "AppData\Roaming\npm\node_modules")
    python_user_site = (Join-Path $strictParityFixtureHome "AppData\Roaming\Python\Python311\site-packages")
    pipx_home_venvs = (Join-Path $strictParityFixtureHome "pipx\venvs")
    pipx_localappdata_venvs = (Join-Path $strictParityFixtureHome "AppData\Local\pipx\venvs")
    pipx_legacy_venvs = (Join-Path $strictParityFixtureHome ".local\pipx\venvs")
}
$strictParityRootListed = [ordered]@{}
foreach ($key in $strictParityExpectedRoots.Keys) {
    $path = $strictParityExpectedRoots[$key]
    $strictParityRootListed[$key] = @($strictParityRoots | Where-Object { $_.Path -eq $path -and $_.Kind -eq "user_package_root" }).Count -eq 1
}

$strictParityRecords = Read-JsonLines $strictParityScanOut
$strictParityRecordTypeCounts = [ordered]@{}
$strictParitySourceTypeCounts = [ordered]@{}
$strictParityPackages = New-Object System.Collections.Generic.List[object]
$strictParitySummary = $null
foreach ($record in $strictParityRecords) {
    $recordType = [string]$record.record_type
    if ([string]::IsNullOrWhiteSpace($recordType)) {
        continue
    }
    if (-not $strictParityRecordTypeCounts.Contains($recordType)) {
        $strictParityRecordTypeCounts[$recordType] = 0
    }
    $strictParityRecordTypeCounts[$recordType]++
    if ($recordType -eq "package") {
        $strictParityPackages.Add($record)
        $sourceType = [string]$record.source_type
        if (-not [string]::IsNullOrWhiteSpace($sourceType)) {
            if (-not $strictParitySourceTypeCounts.Contains($sourceType)) {
                $strictParitySourceTypeCounts[$sourceType] = 0
            }
            $strictParitySourceTypeCounts[$sourceType]++
        }
    }
    if ($recordType -eq "scan_summary") {
        $strictParitySummary = $record
    }
}

$strictParityExpectedNames = @(
    "smoke-npm-root",
    "SmokePythonRoot",
    "SmokePipxRoot",
    "SmokePipxLocalRoot",
    "SmokePipxLegacyRoot"
)
$strictParityExpectedNamesEmitted = [ordered]@{}
foreach ($name in $strictParityExpectedNames) {
    $strictParityExpectedNamesEmitted[$name] = @($strictParityPackages | Where-Object { (Get-JsonProperty $_ "package_name") -eq $name }).Count -eq 1
}

if ($commands.Contains("roots_strict_parity_fixture") -and $commands["roots_strict_parity_fixture"].exit_code -eq 0) {
    foreach ($key in $strictParityRootListed.Keys) {
        if (-not $strictParityRootListed[$key]) {
            $failures.Add("strict-parity roots did not list expected root $key")
        }
    }
}

if ($commands.Contains("scan_strict_parity_fixture") -and $commands["scan_strict_parity_fixture"].exit_code -eq 0) {
    if ($strictParityPackages.Count -ne 5) {
        $failures.Add("strict-parity scan emitted $($strictParityPackages.Count) package records, want 5")
    }
    if ($null -eq $strictParitySummary) {
        $failures.Add("strict-parity scan_summary was missing")
    } elseif ($strictParitySummary.status -ne "complete") {
        $failures.Add("strict-parity scan_summary status was not complete")
    }
    if (-not $strictParitySourceTypeCounts.Contains("npm-node_modules") -or $strictParitySourceTypeCounts["npm-node_modules"] -ne 1) {
        $failures.Add("strict-parity scan did not emit 1 npm node_modules package record")
    }
    if (-not $strictParitySourceTypeCounts.Contains("pypi-dist-info") -or $strictParitySourceTypeCounts["pypi-dist-info"] -ne 4) {
        $failures.Add("strict-parity scan did not emit 4 PyPI dist-info records")
    }
    foreach ($name in $strictParityExpectedNames) {
        if (-not $strictParityExpectedNamesEmitted[$name]) {
            $failures.Add("strict-parity scan did not emit expected package $name")
        }
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
        brave_default_extensions_exists = [bool]$braveDefaultExtensionsExists
        brave_default_extensions_listed = [bool]$braveDefaultExtensionsListed
        chromium_default_extensions_exists = [bool]$chromiumDefaultExtensionsExists
        chromium_default_extensions_listed = [bool]$chromiumDefaultExtensionsListed
        edge_default_extensions_exists = [bool]$edgeDefaultExtensionsExists
        edge_default_extensions_listed = [bool]$edgeDefaultExtensionsListed
        vivaldi_default_extensions_exists = [bool]$vivaldiDefaultExtensionsExists
        vivaldi_default_extensions_listed = [bool]$vivaldiDefaultExtensionsListed
        firefox_profiles_exists = [bool]$firefoxProfilesExists
        firefox_profiles_listed = [bool]$firefoxProfilesListed
        librewolf_profiles_exists = [bool]$librewolfProfilesExists
        librewolf_profiles_listed = [bool]$librewolfProfilesListed
        waterfox_profiles_exists = [bool]$waterfoxProfilesExists
        waterfox_profiles_listed = [bool]$waterfoxProfilesListed
        waterfox_legacy_profiles_exists = [bool]$waterfoxLegacyProfilesExists
        waterfox_legacy_profiles_listed = [bool]$waterfoxLegacyProfilesListed
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
    nuget_project_scan = [ordered]@{
        record_type_counts = $nugetRecordTypeCounts
        source_type_counts = $nugetSourceTypeCounts
        package_records = $nugetPackages.Count
        summary_present = $null -ne $nugetSummary
        summary_status = Get-JsonProperty $nugetSummary "status"
        summary_profile = Get-JsonProperty $nugetSummary "profile"
        requested_spec_count = $nugetRequestedSpecCount
        direct_true_count = $nugetDirectTrueCount
        direct_false_count = $nugetDirectFalseCount
        direct_empty_count = $nugetDirectEmptyCount
        transitive_scope_count = $nugetTransitiveScopeCount
        project_root_count = $nugetProjectRootCount
        missing_required_fields = $nugetMissingRequiredFields
        skipped_project_reference_emitted = $nugetProjectReferenceEmitted
        skipped_missing_version_emitted = $nugetMissingVersionEmitted
        skipped_missing_resolved_emitted = $nugetMissingResolvedEmitted
    }
    powershell_module_project_scan = [ordered]@{
        record_type_counts = $powershellRecordTypeCounts
        source_type_counts = $powershellSourceTypeCounts
        package_records = $powershellPackages.Count
        summary_present = $null -ne $powershellSummary
        summary_status = Get-JsonProperty $powershellSummary "status"
        summary_profile = Get-JsonProperty $powershellSummary "profile"
        project_root_count = $powershellProjectRootCount
        missing_required_fields = $powershellMissingRequiredFields
        expected_pester_emitted = $powershellPesterEmitted
        skipped_missing_version_emitted = $powershellNoVersionEmitted
        skipped_expression_version_emitted = $powershellExpressionVersionEmitted
    }
    browser_extension_project_scan = [ordered]@{
        record_type_counts = $browserRecordTypeCounts
        source_type_counts = $browserSourceTypeCounts
        package_manager_counts = $browserManagerCounts
        package_records = $browserPackages.Count
        summary_present = $null -ne $browserSummary
        summary_status = Get-JsonProperty $browserSummary "status"
        summary_profile = Get-JsonProperty $browserSummary "profile"
        project_root_count = $browserProjectRootCount
        missing_required_fields = $browserMissingRequiredFields
        expected_names_emitted = $browserExpectedNamesEmitted
    }
    strict_parity_fixture = [ordered]@{
        root_count = @($strictParityRoots).Count
        expected_roots_listed = $strictParityRootListed
        record_type_counts = $strictParityRecordTypeCounts
        source_type_counts = $strictParitySourceTypeCounts
        package_records = $strictParityPackages.Count
        summary_present = $null -ne $strictParitySummary
        summary_status = Get-JsonProperty $strictParitySummary "status"
        summary_profile = Get-JsonProperty $strictParitySummary "profile"
        expected_names_emitted = $strictParityExpectedNamesEmitted
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
