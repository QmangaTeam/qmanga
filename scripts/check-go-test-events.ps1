param([Parameter(Mandatory)][string]$Path)
$ErrorActionPreference = 'Stop'
$events = @(
    foreach ($line in Get-Content -LiteralPath $Path) {
        # Compose/build progress is not part of the go test JSON stream.
        if (-not $line.TrimStart().StartsWith('{')) { continue }
        $event = $line | ConvertFrom-Json -ErrorAction Stop
        if (-not $event.Action -or -not $event.Package) { throw 'Malformed Go test event' }
        $event
    }
)
$failed = @($events | Where-Object Action -eq 'fail')
$skipped = @($events | Where-Object { $_.Action -eq 'skip' -and $_.Test })
$passed = @($events | Where-Object { $_.Action -eq 'pass' -and $_.Test })
$packages = @($events | Where-Object { $_.Action -eq 'pass' -and -not $_.Test })
if ($failed.Count) { throw "Go test failures: $($failed.Package -join ', ')" }
if ($skipped.Count) { throw "Actual tests skipped: $($skipped.Test -join ', ')" }
if (-not $passed.Count -or -not $packages.Count) { throw 'Missing completed Go test/package results' }
foreach ($package in @($events.Package | Sort-Object -Unique)) {
    if (-not @($events | Where-Object { $_.Package -eq $package -and -not $_.Test -and $_.Action -in @('pass','skip') }).Count) {
        throw "Incomplete Go test package: $package"
    }
}
Write-Output "$($passed.Count) test/subtest passes; no actual skips; $($packages.Count) packages passed."
