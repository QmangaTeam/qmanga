$ErrorActionPreference = 'Stop'
$repository = Split-Path $PSScriptRoot -Parent
$temporaryRoot = [IO.Path]::GetFullPath((Join-Path $repository ('tmp/setup-check-' + [guid]::NewGuid().ToString('N'))))
if (-not $temporaryRoot.StartsWith([IO.Path]::GetFullPath((Join-Path $repository 'tmp')) + [IO.Path]::DirectorySeparatorChar)) { throw 'Unsafe test directory.' }
$null = New-Item -ItemType Directory -Path (Join-Path $temporaryRoot 'scripts') -Force
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'setup.ps1') -Destination (Join-Path $temporaryRoot 'scripts/setup.ps1')
$scriptPath = Join-Path $temporaryRoot 'scripts/setup.ps1'
$destination = Join-Path $temporaryRoot '.env'
$global:qmangaTestVolumes = 'qmanga_postgres-data'
$global:qmangaTestDockerFailure = $false
function docker {
 if ($global:qmangaTestDockerFailure) { $global:LASTEXITCODE = 1; return }
 $global:LASTEXITCODE = 0
 if ($global:qmangaTestVolumes) { $global:qmangaTestVolumes }
}
try {
 $refused = $false
 try { & $scriptPath } catch { $refused = $true }
 if (-not $refused -or (Test-Path -LiteralPath $destination)) { throw 'Setup must refuse new secrets when old data exists.' }
 $global:qmangaTestVolumes = ''
 $global:qmangaTestDockerFailure = $true
 $refused = $false
 try { & $scriptPath } catch { $refused = $true }
 if (-not $refused -or (Test-Path -LiteralPath $destination)) { throw 'Setup must not guess when Docker cannot inspect volumes.' }
 $global:qmangaTestDockerFailure = $false
 & $scriptPath -ProjectName qmanga-setup-test
 $settings = [IO.File]::ReadAllText($destination)
 if ($settings -notmatch 'COMPOSE_PROJECT_NAME=qmanga-setup-test' -or $settings -notmatch 'POSTGRES_PASSWORD=[a-f0-9]{64}') { throw 'New environment configuration is incomplete.' }
 $global:qmangaTestDockerFailure = $true
 & $scriptPath
 if ([IO.File]::ReadAllText($destination) -ne $settings) { throw 'Setup changed existing credentials.' }
 Write-Output 'Setup recovery guards and credential preservation passed.'
} finally { Remove-Item -LiteralPath $temporaryRoot -Recurse -Force }
