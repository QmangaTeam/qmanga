param(
    [Parameter(Mandatory)][string]$ReleaseId,
    [string]$AppUrl = 'http://localhost:18080',
    [string]$OutputFile = 'tmp/reader-measurement.json'
)
$ErrorActionPreference = 'Stop'
if (-not ([uri]$AppUrl).IsLoopback) { throw 'Run measurements against the local acceptance environment.' }
$client = [Net.Http.HttpClient]::new()
$client.Timeout = [TimeSpan]::FromSeconds(30)
try {
    $watch = [Diagnostics.Stopwatch]::StartNew()
    $response = $client.GetAsync("$AppUrl/api/v1/releases/$([uri]::EscapeDataString($ReleaseId))/manifest").GetAwaiter().GetResult()
    $response.EnsureSuccessStatusCode() | Out-Null
    $json = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
    $watch.Stop()
    $manifest = $json | ConvertFrom-Json
    $manifestMs = $watch.Elapsed.TotalMilliseconds
    $tiers = @()
    foreach ($tier in @('eco','standard','high')) {
        $files = @($manifest.pages | ForEach-Object { $_.variants.$tier })
        if (-not $files.Count) { throw "No $tier assets" }
        $bytes = ($files | Measure-Object -Property bytes -Sum).Sum
        $first = $files[0]
        if ($first.url -notmatch '^/api/v1/media/[A-Za-z0-9-]+$') { throw 'Unexpected media URL' }
        $samples = @()
        for ($i=0; $i -lt 3; $i++) {
            $watch.Restart()
            $data = $client.GetByteArrayAsync("$AppUrl$($first.url)").GetAwaiter().GetResult()
            $watch.Stop()
            if ($data.Length -ne $first.bytes) { throw 'Asset size mismatch' }
            $hash = [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($data)).ToLowerInvariant()
            if ($hash -ne $first.checksum) { throw 'Asset checksum mismatch' }
            $samples += [math]::Round($watch.Elapsed.TotalMilliseconds,2)
        }
        $tiers += [ordered]@{
            tier=$tier; fragments=$files.Count; chapter_bytes=$bytes; first_fragment_bytes=$first.bytes
            first_fragment_http_ms=$samples
            # Transfer-only mathematical estimates, not emulated browser timings.
            transfer_floor_seconds_at_1_6_mbps=[math]::Round($bytes*8/1600000,3)
            transfer_floor_seconds_at_9_mbps=[math]::Round($bytes*8/9000000,3)
            transfer_floor_seconds_at_30_mbps=[math]::Round($bytes*8/30000000,3)
            transfer_gib_for_1000_complete_reads=[math]::Round($bytes*1000/1GB,4)
        }
    }
    $result = [ordered]@{
        measured_at_utc=[DateTime]::UtcNow.ToString('o'); release_id=$ReleaseId
        pages=@($manifest.pages).Count; manifest_http_ms=[math]::Round($manifestMs,2)
        manifest_utf8_bytes=[Text.Encoding]::UTF8.GetByteCount($json)
        method='Loopback HTTP; 3 sequential requests per tier; checksummed synthetic assets. HTTP completion is not browser paint/TTFI. Transfer estimates exclude latency, retry, cache and protocol overhead.'
        tiers=$tiers
    }
    $absolute=[IO.Path]::GetFullPath($OutputFile)
    [IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($absolute)) | Out-Null
    [IO.File]::WriteAllText($absolute,($result | ConvertTo-Json -Depth 8))
    Write-Output "Saved reader measurements: $OutputFile"
} finally { $client.Dispose(); if ($response) { $response.Dispose() } }
