#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Expose your Hearth House to the internet via Cloudflare Tunnel.
.DESCRIPTION
    Uses `cloudflared tunnel --url` (quick tunnel, no account needed) to expose
    PocketBase at localhost:8090 over a random *.trycloudflare.com HTTPS URL.

    PocketBase serves both the API and the SPA (from pb_public/), so the tunnel
    URL is all a friend needs — no CORS issues, no separate frontend.

    Prerequisites:
    - PocketBase must be running: cd backend && go run . serve --http=0.0.0.0:8090
    - cloudflared must be installed: winget install Cloudflare.cloudflared
.NOTES
    The tunnel URL changes every time you restart. For a permanent URL,
    set up a named tunnel with `cloudflared tunnel create`.
#>

param(
    [int]$Port = 8090,
    [string]$Metrics = "localhost:8091"
)

$ErrorActionPreference = "Stop"

# Verify cloudflared is installed
if (-not (Get-Command cloudflared -ErrorAction SilentlyContinue)) {
    Write-Host "cloudflared not found. Install it:" -ForegroundColor Red
    Write-Host "  winget install Cloudflare.cloudflared" -ForegroundColor Yellow
    exit 1
}

# Verify PocketBase is running
try {
    $health = Invoke-RestMethod -Uri "http://localhost:$Port/api/health" -TimeoutSec 3
    Write-Host "PocketBase is running (health: $($health.code))" -ForegroundColor Green
} catch {
    Write-Host "PocketBase is not running on port $Port." -ForegroundColor Red
    Write-Host "Start it first:" -ForegroundColor Yellow
    Write-Host "  cd backend; go run . serve --http=0.0.0.0:$Port" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Hearth — Cloudflare Tunnel" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Starting tunnel to localhost:$Port..." -ForegroundColor White
Write-Host "Watch for the URL below (*.trycloudflare.com)" -ForegroundColor Yellow
Write-Host "Share it with a friend — they can open it in any browser." -ForegroundColor White
Write-Host ""
Write-Host "Press Ctrl+C to stop the tunnel." -ForegroundColor DarkGray
Write-Host ""

# Run the tunnel — cloudflared prints the URL to stderr
cloudflared tunnel --url "http://localhost:$Port" --metrics "$Metrics"
