#!/bin/bash
# Hearth Quick Test — expose your local PocketBase to a friend via Cloudflare Tunnel
# Usage: ./scripts/quick-test.sh
# Requires: cloudflared (https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/)

set -e

PORT="${HEARTH_PORT:-8090}"

echo ""
echo "  🔥 Hearth Quick Test Mode"
echo "  ========================="
echo ""
echo "  Starting Cloudflare Tunnel to localhost:${PORT}..."
echo "  Your friend will get a URL to connect."
echo "  Press Ctrl+C to stop."
echo ""

# Check if cloudflared is installed
if ! command -v cloudflared &> /dev/null; then
    echo "  ❌ cloudflared not found."
    echo ""
    echo "  Install it:"
    echo "    macOS:   brew install cloudflared"
    echo "    Linux:   curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o /usr/local/bin/cloudflared && chmod +x /usr/local/bin/cloudflared"
    echo "    Windows: winget install Cloudflare.cloudflared"
    echo ""
    exit 1
fi

# Check if PocketBase is running
if ! curl -s "http://localhost:${PORT}/api/health" > /dev/null 2>&1; then
    echo "  ⚠️  PocketBase doesn't seem to be running on localhost:${PORT}"
    echo "  Start it first: cd backend && go run . serve --http=0.0.0.0:${PORT}"
    echo ""
    echo "  Continuing anyway (Cloudflare Tunnel will wait for it)..."
    echo ""
fi

cloudflared tunnel --url "http://localhost:${PORT}"
