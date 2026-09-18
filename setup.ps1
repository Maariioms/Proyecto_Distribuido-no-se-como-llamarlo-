Write-Host "== XSIAM-sito: script de automatizacion =="

$docker = Get-Command docker -ErrorAction SilentlyContinue
if (-not $docker) {
    Write-Host "Docker esta en el gulag"
    exit 1
}

docker info > $null 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Docker esta en el server pero no ha hecho spawn"
    exit 1
}

Write-Host "Docker si esta, ahorita despierta a todos"
docker compose up --build

Write-Host "Frontend :3000, middleware en :9000"
