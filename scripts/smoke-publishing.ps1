param([Parameter(Mandatory=$true)][string]$SourceDirectory, [int]$PageLimit = 5, [string]$AppUrl = 'http://localhost:8080', [string]$MailboxUrl = 'http://localhost:8025', [string]$ProjectName = '', [string]$EnvFile = '.env', [string]$ResultFile = '', [switch]$ConfirmRights, [string]$SourceNote = 'Local development fixture; publication permission confirmed by the test operator.', [string]$Credits = 'Qmanga test operator')
$ErrorActionPreference = 'Stop'
$compose = @('compose', '--env-file', $EnvFile)
if ($ProjectName) { $compose += @('-p', $ProjectName) }
if ($ResultFile) {
    git check-ignore -q -- $ResultFile
    if ($LASTEXITCODE -ne 0) { throw 'ResultFile contains test credentials and must be inside a Git-ignored directory (for example tmp/stage1/fixture.json).' }
}
if (-not ([uri]$AppUrl).IsLoopback -or -not ([uri]$MailboxUrl).IsLoopback) { throw 'Local development only.' }
if (-not $ConfirmRights -or $SourceNote.Trim().Length -lt 8 -or $Credits.Trim().Length -lt 2) { throw 'Use permitted test images and explicitly pass -ConfirmRights with their -SourceNote and -Credits.' }
if ($PageLimit -lt 1 -or $PageLimit -gt 300) { throw 'PageLimit must be 1..300.' }
$files = @(Get-ChildItem -LiteralPath (Resolve-Path -LiteralPath $SourceDirectory).Path -File | Where-Object { $_.Extension -in @('.jpg','.jpeg','.png','.webp') } | Sort-Object Name | Select-Object -First $PageLimit)
if (-not $files.Count) { throw 'No supported images found.' }
$clients = [Collections.Generic.List[System.Net.Http.HttpClient]]::new()
function New-TestClient {
    $handler = [System.Net.Http.HttpClientHandler]::new()
    $handler.CookieContainer = [System.Net.CookieContainer]::new()
    $client = [System.Net.Http.HttpClient]::new($handler)
    $client.Timeout = [TimeSpan]::FromMinutes(10)
    $clients.Add($client)
    return $client
}
function Send-Api($Client, [string]$Method, [string]$Path, $Body = $null, [string]$Csrf = '', [int]$Expected = 200) {
    $request = [System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::new($Method), "$AppUrl/api/v1/$Path")
    $request.Headers.Add('Origin', ([uri]$AppUrl).GetLeftPart([System.UriPartial]::Authority))
    if ($Csrf) { $request.Headers.Add('X-CSRF-Token', $Csrf) }
    if ($null -ne $Body) { $request.Content = [System.Net.Http.StringContent]::new(($Body | ConvertTo-Json -Depth 12 -Compress), [Text.Encoding]::UTF8, 'application/json') }
    try {
        $response = $Client.SendAsync($request).GetAwaiter().GetResult()
        try {
            $text = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
            $data = if ($text) { $text | ConvertFrom-Json } else { $null }
            if ([int]$response.StatusCode -ne $Expected) { throw "API $Method $Path expected $Expected, received $([int]$response.StatusCode): $($data.error.code)" }
            return $data
        } finally { $response.Dispose() }
    } finally { $request.Dispose() }
}
function Verification-Token([string]$Email) {
    for ($attempt=0; $attempt -lt 25; $attempt++) {
        $mailbox = Invoke-RestMethod -TimeoutSec 5 "$MailboxUrl/api/v1/messages"
        $matching = @($mailbox.messages | Where-Object { @($_.To | Where-Object { $_.Address -eq $Email }).Count -gt 0 })
        if ($matching.Count) {
            $message = Invoke-RestMethod -TimeoutSec 5 "$MailboxUrl/api/v1/message/$($matching[0].ID)"
            if ($message.Text -match '/verify-email\?token=([A-Za-z0-9_-]{43})') { return $Matches[1] }
        }
        Start-Sleep -Milliseconds 200
    }
    throw 'Verification email not received.'
}
try {
    $admin = New-TestClient
    $owner = New-TestClient
    $anonymous = New-TestClient
    $suffix = [guid]::NewGuid().ToString('N')
    $adminEmail = "publishing-admin-$suffix@example.test"
    $adminPassword = [guid]::NewGuid().ToString('N') + '-Aa1'
    $ownerEmail = "publishing-owner-$suffix@example.test"
    $ownerPassword = [guid]::NewGuid().ToString('N') + '-Bb2'
    $adminPassword | docker @compose exec -T -e "ADMIN_EMAIL=$adminEmail" api qmanga-admin | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Bootstrap administrator failed.' }
    $null = Send-Api $admin POST 'auth/login' @{email=$adminEmail; password=$adminPassword}
    $adminCsrf = (Send-Api $admin GET 'auth/csrf').csrf_token
    $null = Send-Api $owner POST 'auth/register' @{email=$ownerEmail; password=$ownerPassword} -Expected 202
    $null = Send-Api $owner POST 'auth/verify-email' @{token=(Verification-Token $ownerEmail)}
    $null = Send-Api $owner POST 'auth/login' @{email=$ownerEmail; password=$ownerPassword}
    $ownerCsrf = (Send-Api $owner GET 'auth/csrf').csrf_token
    $team = Send-Api $owner POST 'studio/teams' @{slug="test-$($suffix.Substring(0,12))";name='Жергілікті сынақ тобы';description='Local development test'} $ownerCsrf 201
    $null = Send-Api $admin POST "admin/teams/$($team.id)/approve" @{} $adminCsrf 204
    $slug = "local-reader-test-$($suffix.Substring(0,12))"
    $title = Send-Api $admin POST 'admin/titles' @{slug=$slug;name='Жергілікті оқу сынағы';original_name='';aliases=@();kind='manga';original_status='ongoing';translation_status='ongoing';description='Пайдаланушы берген файлдармен жергілікті сынақ. Бұл файлдар GitHub-қа жіберілмейді.';direction='rtl';genre_ids=@();demo=$true} $adminCsrf 201
    $chapter = Send-Api $admin POST "admin/titles/$($title.id)/chapters" @{label='1';order=1;volume='1';name='Сынақ тарауы'} $adminCsrf 201
    $project = Send-Api $admin POST "admin/titles/$($title.id)/projects" @{team_id=$team.id} $adminCsrf 201
    $coverRequest = [System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::Put, "$AppUrl/api/v1/studio/projects/$($project.id)/cover-file")
    $coverRequest.Headers.Add('Origin', ([uri]$AppUrl).GetLeftPart([System.UriPartial]::Authority))
    $coverRequest.Headers.Add('X-CSRF-Token', $ownerCsrf)
    $coverStream = [IO.File]::OpenRead($files[0].FullName)
    $coverRequest.Content = [System.Net.Http.StreamContent]::new($coverStream)
    $coverRequest.Content.Headers.ContentType = [System.Net.Http.Headers.MediaTypeHeaderValue]::new('application/octet-stream')
    try {
        $coverResponse = $owner.SendAsync($coverRequest).GetAwaiter().GetResult()
        try {
            $coverResult = $coverResponse.Content.ReadAsStringAsync().GetAwaiter().GetResult() | ConvertFrom-Json
            if ([int]$coverResponse.StatusCode -ne 201 -or -not $coverResult.cover_id) { throw "Independent cover upload failed: $([int]$coverResponse.StatusCode)" }
        } finally { $coverResponse.Dispose() }
    } finally { $coverRequest.Dispose(); $coverStream.Dispose() }
    $studioProject = Send-Api $owner GET "studio/projects/$($project.id)"
    if ($studioProject.cover_id -ne $coverResult.cover_id) { throw 'Independent cover did not appear in studio project.' }
    $coverMedia = $anonymous.GetAsync("$AppUrl$($coverResult.url)").GetAwaiter().GetResult()
    try { if ([int]$coverMedia.StatusCode -ne 200 -or $coverMedia.Content.Headers.ContentType.MediaType -ne 'image/webp') { throw 'Independent cover is not public WebP media.' } }
    finally { $coverMedia.Dispose() }
    $draft = Send-Api $owner POST "studio/projects/$($project.id)/revisions" @{chapter_id=$chapter.id} $ownerCsrf 201
    $null = Send-Api $anonymous GET "titles/$slug" -Expected 404
    $position = 0
    foreach ($file in $files) {
        $position++
        $request = [System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::Put, "$AppUrl/api/v1/studio/revisions/$($draft.id)/pages?position=$position&key=local-file-$position")
        $request.Headers.Add('Origin', ([uri]$AppUrl).GetLeftPart([System.UriPartial]::Authority))
        $request.Headers.Add('X-CSRF-Token', $ownerCsrf)
        $stream = [IO.File]::OpenRead($file.FullName)
        $request.Content = [System.Net.Http.StreamContent]::new($stream)
        $request.Content.Headers.ContentType = [System.Net.Http.Headers.MediaTypeHeaderValue]::new('application/octet-stream')
        try {
            $response = $owner.SendAsync($request).GetAwaiter().GetResult()
            try { if ([int]$response.StatusCode -notin @(200,201)) { throw "Upload $position failed: $([int]$response.StatusCode)" } } finally { $response.Dispose() }
        } finally { $request.Dispose(); $stream.Dispose() }
        Write-Output "Accepted page $position/$($files.Count)"
    }
    $ready = $false
    for ($attempt=0; $attempt -lt 240; $attempt++) {
        $draft = Send-Api $owner GET "studio/revisions/$($draft.id)"
        if ($draft.state -eq 'ready') { $ready=$true; break }
        if (@($draft.pages | Where-Object status -eq 'failed').Count) { throw 'Image processing failed.' }
        Start-Sleep -Seconds 1
    }
    if (-not $ready) { throw 'Processing deadline exceeded.' }
    $null = Send-Api $owner POST "studio/revisions/$($draft.id)/submit" @{version=$draft.version} $ownerCsrf 409
    $null = Send-Api $owner PUT "studio/revisions/$($draft.id)/attribution" @{version=$draft.version; source_note=$SourceNote; credits=$Credits; rights_confirmed=$true} $ownerCsrf 204
    $draft = Send-Api $owner GET "studio/revisions/$($draft.id)"
    if (-not $draft.rights_confirmed -or -not $draft.source_note -or -not $draft.credits) { throw 'Attribution was not persisted.' }
    $null = Send-Api $owner POST "studio/revisions/$($draft.id)/submit" @{version=$draft.version} $ownerCsrf 204
    $null = Send-Api $admin POST "moderation/revisions/$($draft.id)/review" @{decision='publish';reason='Local test'} $adminCsrf 204
    $manifest = Send-Api $anonymous GET "releases/$($draft.release_id)/manifest"
    if (@($manifest.pages).Count -ne $files.Count -or $manifest.cover_url -ne $coverResult.url) { throw 'Independent cover changed chapter pages or disappeared from manifest.' }
    foreach ($page in $manifest.pages) {
        foreach ($tier in @('eco','standard','high')) {
            $variants = @($page.variants.$tier)
            if (-not $variants.Count) { throw "Missing $tier variants." }
            foreach ($media in $variants) {
                if ($media.mime -ne 'image/webp' -or $media.url -notmatch '^/api/v1/media/[a-zA-Z0-9-]+$') { throw 'Unexpected reader format or direct storage URL.' }
                $response = $anonymous.GetAsync("$AppUrl$($media.url)").GetAwaiter().GetResult()
                try {
                    $bytes = $response.Content.ReadAsByteArrayAsync().GetAwaiter().GetResult()
                    $checksum = [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)).ToLowerInvariant()
                    if ([int]$response.StatusCode -ne 200 -or $response.Content.Headers.ContentType.MediaType -ne 'image/webp' -or $bytes.Length -ne $media.bytes -or $checksum -ne $media.checksum -or $bytes.Length -lt 12 -or [Text.Encoding]::ASCII.GetString($bytes,0,4) -ne 'RIFF' -or [Text.Encoding]::ASCII.GetString($bytes,8,4) -ne 'WEBP') { throw 'Go media WebP bytes, MIME or checksum mismatch.' }
                } finally { $response.Dispose() }
            }
        }
    }
    Write-Output "Publishing smoke passed: $($files.Count) local pages, private source, actual WebP in all three profiles, worker, moderation and Go media."
    Write-Output "Reader: $AppUrl/read/$($draft.release_id)"
    if ($ResultFile) {
        $result = @{ app_url=$AppUrl; admin_email=$adminEmail; admin_password=$adminPassword; owner_email=$ownerEmail; owner_password=$ownerPassword; title_id=$title.id; project_id=$project.id; release_id=$draft.release_id; env_file=[IO.Path]::GetFullPath($EnvFile); project_name=$ProjectName }
        [IO.File]::WriteAllText([IO.Path]::GetFullPath($ResultFile), ($result | ConvertTo-Json), [Text.UTF8Encoding]::new($false))
    }
} finally { foreach ($client in $clients) { $client.Dispose() } }
