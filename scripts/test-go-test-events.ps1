$ErrorActionPreference = 'Stop'
$checker = Join-Path $PSScriptRoot 'check-go-test-events.ps1'
$fixture = Join-Path ([System.IO.Path]::GetTempPath()) ('qmanga-go-events-' + [guid]::NewGuid() + '.jsonl')
try {
    $passed = '{"Action":"pass","Package":"example","Test":"TestExample"}'
    [IO.File]::WriteAllText($fixture, "# Docker build diagnostic`n$passed`n" + '{"Action":"pass","Package":"example"}')
    & $checker -Path $fixture
    foreach ($invalid in @('', '# diagnostics only', '{invalid json}', '{"Action":"skip","Package":"example","Test":"TestDatabase"}', '{"Action":"fail","Package":"example"}', $passed)) {
        [IO.File]::WriteAllText($fixture, $invalid)
        $rejected = $false
        try { & $checker -Path $fixture } catch { $rejected = $true }
        if (-not $rejected) { throw "Checker accepted incomplete or failed test output: $invalid" }
    }
    Write-Output 'Go event checker regressions passed.'
} finally { Remove-Item -LiteralPath $fixture -ErrorAction SilentlyContinue }
