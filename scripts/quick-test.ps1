# Hearth Quick Test — expose your local PocketBase to a friend via Cloudflare Tunnel
# Usage: .\scripts\quick-test.ps1
# Requires: cloudflared (winget install Cloudflare.cloudflared)

$Port = if ($env:HEARTH_PORT) { $env:HEARTH_PORT } else { "8090" }

Write-Host ""
Write-Host "  Hearth Quick Test Mode" -ForegroundColor Yellow
Write-Host "  ========================="
Write-Host ""
Write-Host "  Starting Cloudflare Tunnel to localhost:$Port..."
Write-Host "  Your friend will get a URL to connect."
Write-Host "  Press Ctrl+C to stop."
Write-Host ""

# Check if cloudflared is installed
if (-not (Get-Command "cloudflared" -ErrorAction SilentlyContinue)) {
    Write-Host "  cloudflared not found." -ForegroundColor Red
    Write-Host ""
    Write-Host "  Install it: winget install Cloudflare.cloudflared"
    Write-Host ""
    exit 1
}

# Check if PocketBase is running
try {
    $null = Invoke-WebRequest -Uri "http://localhost:$Port/api/health" -UseBasicParsing -TimeoutSec 3
} catch {
    Write-Host "  PocketBase doesn't seem to be running on localhost:$Port" -ForegroundColor DarkYellow
    Write-Host "  Start it first: cd backend; go run . serve --http=0.0.0.0:$Port"
    Write-Host ""
    Write-Host "  Continuing anyway (Cloudflare Tunnel will wait for it)..."
    Write-Host ""
}

cloudflared tunnel --url "http://localhost:$Port"
