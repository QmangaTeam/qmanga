$ErrorActionPreference = 'Stop'
$destination = Join-Path (Split-Path $PSScriptRoot -Parent) '.env'
if (Test-Path -LiteralPath $destination) { Write-Output '.env already exists; preserved.'; exit 0 }
function New-Secret {
    $randomBytes = New-Object byte[] 32
    $generator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try { $generator.GetBytes($randomBytes) } finally { $generator.Dispose() }
    return ([System.BitConverter]::ToString($randomBytes)).Replace('-', '').ToLowerInvariant()
}
$settings = @(
    'APP_ENV=development'
    'APP_PORT=8080'
    'APP_ORIGIN=http://localhost:8080'
    "POSTGRES_PASSWORD=$(New-Secret)"
    "S3_ACCESS_KEY=qmanga$((New-Secret).Substring(0, 16))"
    "S3_SECRET_KEY=$(New-Secret)"
    "SESSION_SECRET=$(New-Secret)"
    "CSRF_SECRET=$(New-Secret)"
)
[System.IO.File]::WriteAllLines($destination, $settings, (New-Object System.Text.UTF8Encoding $false))
Write-Output 'Created local .env with random secrets.'

