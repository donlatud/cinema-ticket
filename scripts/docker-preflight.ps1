# Run before: docker compose up --build
# Usage: .\scripts\docker-preflight.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent

Write-Host "Checking Docker prerequisites in $root" -ForegroundColor Cyan

$ok = $true

$keyPath = Join-Path $root "backend\firebase-key.json"
if (-not (Test-Path $keyPath)) {
    Write-Host "[FAIL] Missing backend\firebase-key.json" -ForegroundColor Red
    Write-Host "       Download Firebase Service Account key and save it there."
    $ok = $false
} elseif ((Get-Item $keyPath) -is [System.IO.DirectoryInfo]) {
    Write-Host "[FAIL] backend\firebase-key.json is a FOLDER (Windows Docker bug)" -ForegroundColor Red
    Write-Host "       Delete the folder and add the real JSON file."
    $ok = $false
} else {
    $first = Get-Content $keyPath -TotalCount 1 -ErrorAction SilentlyContinue
    if ($first -notmatch '^\s*\{') {
        Write-Host "[FAIL] backend\firebase-key.json is not valid JSON" -ForegroundColor Red
        $ok = $false
    } else {
        Write-Host "[OK]   backend\firebase-key.json" -ForegroundColor Green
    }
}

$envPath = Join-Path $root ".env"
if (-not (Test-Path $envPath)) {
    Write-Host "[FAIL] Missing .env at project root (copy from .env.example)" -ForegroundColor Red
    $ok = $false
} else {
    $envContent = Get-Content $envPath -Raw
    foreach ($var in @("VITE_FIREBASE_API_KEY", "VITE_FIREBASE_AUTH_DOMAIN", "VITE_FIREBASE_PROJECT_ID")) {
        if ($envContent -notmatch "$var=.+") {
            Write-Host "[FAIL] $var is empty in .env" -ForegroundColor Red
            $ok = $false
        }
    }
    if ($ok) {
        Write-Host "[OK]   root .env (Firebase client config)" -ForegroundColor Green
    }
}

try {
    docker version *> $null
    Write-Host "[OK]   Docker CLI responds" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] Docker is not running. Start Docker Desktop first." -ForegroundColor Red
    $ok = $false
}

if (-not $ok) {
    Write-Host ""
    Write-Host "Fix the issues above, then run: docker compose up --build" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "Ready. Run: docker compose up --build" -ForegroundColor Green
