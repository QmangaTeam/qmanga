param([string]$AppUrl = 'http://localhost:8080', [string]$MailboxUrl = 'http://localhost:8025')
$ErrorActionPreference = 'Stop'
$appUri = [uri]$AppUrl
$mailUri = [uri]$MailboxUrl
if (-not $appUri.IsLoopback -or -not $mailUri.IsLoopback) { throw 'This test only runs on localhost.' }
Add-Type -AssemblyName System.Net.Http
$handler = New-Object System.Net.Http.HttpClientHandler
$handler.CookieContainer = New-Object System.Net.CookieContainer
$client = New-Object System.Net.Http.HttpClient($handler)
$client.Timeout = [TimeSpan]::FromSeconds(20)
$email = "qmanga-smoke-$([guid]::NewGuid().ToString('N'))@example.test"
$password = [guid]::NewGuid().ToString('N') + '-Aa1'
$changedPassword = [guid]::NewGuid().ToString('N') + '-Bb2'
function Send-Auth([string]$Method, [string]$Path, [object]$Body, [string]$Csrf = '') {
    $message = New-Object System.Net.Http.HttpRequestMessage((New-Object System.Net.Http.HttpMethod($Method)), "$AppUrl/api/v1/auth/$Path")
    $message.Headers.Add('Origin', $appUri.GetLeftPart([System.UriPartial]::Authority))
    if ($Csrf) { $message.Headers.Add('X-CSRF-Token', $Csrf) }
    if ($null -ne $Body) { $message.Content = New-Object System.Net.Http.StringContent(($Body | ConvertTo-Json -Compress), [Text.Encoding]::UTF8, 'application/json') }
    try {
        $response = $client.SendAsync($message).GetAwaiter().GetResult()
        $text = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        $data = if ($text) { $text | ConvertFrom-Json } else { $null }
        return @{ Status = [int]$response.StatusCode; Data = $data }
    } finally { $message.Dispose() }
}
function Assert-Status($Result, [int]$Expected) { if ($Result.Status -ne $Expected) { throw "Expected HTTP $Expected, received $($Result.Status)" } }
function Read-Token([string]$Recipient, [string]$Page) {
    for ($attempt = 0; $attempt -lt 20; $attempt++) {
        $mailbox = Invoke-RestMethod -TimeoutSec 5 "$MailboxUrl/api/v1/messages"
        $matching = @($mailbox.messages | Where-Object { @($_.To | Where-Object { $_.Address -eq $Recipient }).Count -gt 0 })
        if ($matching.Count) {
            $message = Invoke-RestMethod -TimeoutSec 5 "$MailboxUrl/api/v1/message/$($matching[0].ID)"
            $tokenPattern = '/' + [regex]::Escape($Page) + '\?token=([A-Za-z0-9_-]{43})'
            if ($message.Text -match $tokenPattern) { return $Matches[1] }
        }
        Start-Sleep -Milliseconds 200
    }
    throw 'Test email was not received.'
}
try {
    Assert-Status (Send-Auth POST register @{ email = $email; password = $password; display_name = 'Smoke Reader' }) 202
    $verification = Read-Token $email 'verify-email'
    Assert-Status (Send-Auth POST verify-email @{ token = $verification }) 200
    Assert-Status (Send-Auth POST login @{ email = $email; password = $password }) 200
    $accountResult = Send-Auth GET me $null
    Assert-Status $accountResult 200
    if ($accountResult.Data.user.display_name -ne 'Smoke Reader') { throw 'Registration nickname was not persisted.' }
    if (-not $accountResult.Data.user.email_verified) { throw 'Email was not verified.' }
    Assert-Status (Send-Auth POST logout @{}) 403
    $csrf = (Send-Auth GET csrf $null).Data.csrf_token
    Assert-Status (Send-Auth POST logout @{} $csrf) 204
    Assert-Status (Send-Auth GET me $null) 401
    Assert-Status (Send-Auth POST forgot-password @{ email = $email }) 202
    $reset = Read-Token $email 'reset-password'
    Assert-Status (Send-Auth POST reset-password @{ token = $reset; password = $changedPassword }) 200
    Assert-Status (Send-Auth POST reset-password @{ token = $reset; password = $changedPassword }) 410
    Assert-Status (Send-Auth POST login @{ email = $email; password = $password }) 401
    Assert-Status (Send-Auth POST login @{ email = $email; password = $changedPassword }) 200
    Write-Output 'Auth smoke passed: Go/Gin + PostgreSQL + real SMTP capture, verification, CSRF, logout and password reset.'
} finally { $client.Dispose(); $handler.Dispose() }
