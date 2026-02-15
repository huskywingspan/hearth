#!/usr/bin/env bash
#
# Expose your Hearth House to the internet via Cloudflare Tunnel.
#
# Uses `cloudflared tunnel --url` (quick tunnel, no account needed) to expose
# PocketBase at localhost:8090 over a random *.trycloudflare.com HTTPS URL.
#
# Prerequisites:
# - PocketBase must be running: cd backend && go run . serve --http=0.0.0.0:8090
# - cloudflared must be installed: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/
#

set -euo pipefail

PORT="${1:-8090}"
METRICS="localhost:8091"

# Verify cloudflared is installed
if ! command -v cloudflared &> /dev/null; then
    echo "❌ cloudflared not found. Install it:"
    echo "   https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/"
    exit 1
fi

# Verify PocketBase is running
if ! curl -sf "http://localhost:$PORT/api/health" > /dev/null 2>&1; then
    echo "❌ PocketBase is not running on port $PORT."
    echo "   Start it first: cd backend && go run . serve --http=0.0.0.0:$PORT"
    exit 1
fi

echo ""
echo "========================================"
echo "  Hearth — Cloudflare Tunnel"
echo "========================================"
echo ""
echo "Starting tunnel to localhost:$PORT..."
echo "Watch for the URL below (*.trycloudflare.com)"
echo "Share it with a friend — they can open it in any browser."
echo ""
echo "Press Ctrl+C to stop the tunnel."
echo ""

cloudflared tunnel --url "http://localhost:$PORT" --metrics "$METRICS"
