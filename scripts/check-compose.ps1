param([string]$EnvFile = '.env')
$ErrorActionPreference = 'Stop'
foreach ($project in @('qmanga-check-one', 'qmanga-check-two')) {
    $raw = docker compose --env-file $EnvFile -p $project config --format json
    if ($LASTEXITCODE -ne 0) { throw 'Compose configuration failed.' }
    $config = ($raw -join "`n") | ConvertFrom-Json
    $api = $config.services.api
    $worker = $config.services.worker
    if (-not $api.image -or $worker.image -ne $api.image) { throw 'API and worker must name the same project-scoped image.' }
    if ($api.image -notlike "$project-*") { throw 'Image must be isolated by Compose project.' }
    if (-not $worker.build -or $worker.build.context -ne $api.build.context -or $worker.build.dockerfile -ne $api.build.dockerfile) { throw 'Worker must be buildable without pulling an unpublished API image.' }
    if ($worker.pull_policy -ne 'build' -or $api.pull_policy -ne 'build') { throw 'Both shared-image services must prefer the local build on a cold start.' }
}
Write-Output 'Compose shared-image/project isolation checks passed.'
