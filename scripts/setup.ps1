param(
    [ValidatePattern('^[a-z0-9][a-z0-9_-]*$')][string]$ProjectName = $(if ($env:COMPOSE_PROJECT_NAME) { $env:COMPOSE_PROJECT_NAME } else { 'qmanga' }),
    [string]$EnvFile = '.env',
    [ValidateRange(1024,65535)][int]$AppPort = 8080,
    [ValidateRange(1024,65535)][int]$MailpitPort = 8025
)
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$destination = if ([IO.Path]::IsPathRooted($EnvFile)) { $EnvFile } else { Join-Path $root $EnvFile }
if (Test-Path -LiteralPath $destination) { Write-Output '.env already exists; preserved.'; exit 0 }
$volumes = @(docker volume ls --filter "label=com.docker.compose.project=$ProjectName" --format '{{.Name}}')
if ($LASTEXITCODE -ne 0) { throw 'Cannot inspect Docker volumes. Start Docker Desktop and retry; no credentials were generated.' }
if ($volumes.Count) { throw "Existing volumes found for '$ProjectName', but no env file. Restore its original credentials or choose a NEW -ProjectName and -EnvFile. Do not delete data to start the app." }
if (-not (Test-Path -LiteralPath (Split-Path $destination -Parent))) { $null = New-Item -ItemType Directory -Path (Split-Path $destination -Parent) -Force }
function New-Secret {
    $randomBytes = New-Object byte[] 32
    $generator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try { $generator.GetBytes($randomBytes) } finally { $generator.Dispose() }
    return ([System.BitConverter]::ToString($randomBytes)).Replace('-', '').ToLowerInvariant()
}
$settings = @(
    'APP_ENV=development'
    "COMPOSE_PROJECT_NAME=$ProjectName"
    "APP_PORT=$AppPort"
    "APP_ORIGIN=http://localhost:$AppPort"
    "MAILPIT_PORT=$MailpitPort"
    "POSTGRES_PASSWORD=$(New-Secret)"
    "S3_ACCESS_KEY=qmanga$((New-Secret).Substring(0, 16))"
    "S3_SECRET_KEY=$(New-Secret)"
    "SESSION_SECRET=$(New-Secret)"
    "CSRF_SECRET=$(New-Secret)"
)
[System.IO.File]::WriteAllLines($destination, $settings, (New-Object System.Text.UTF8Encoding $false))
Write-Output 'Created local .env with random secrets.'

