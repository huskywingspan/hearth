# Hearth — Local LAN Setup Guide

> **Purpose:** Run Hearth on your PC and access it from other devices on the same network.  
> **Requirements:** Go 1.24+, Node.js 22+, Windows/macOS/Linux  
> **No Docker, Caddy, or TLS needed** — PocketBase serves everything from one port.

---

## Architecture

```
Your PC (server)                    Other Devices (clients)
┌────────────────────┐              ┌──────────────────┐
│  PocketBase :8090  │◄── HTTP ────►│  Browser          │
│  ├── API (/api/)   │   over LAN   │  http://192.168.  │
│  ├── SSE (/_/)     │              │    x.x:8090       │
│  └── SPA (pb_public/)             └──────────────────┘
└────────────────────┘
        ▲
        │ also
┌───────┴────────────┐
│  Your PC's browser  │
│  http://localhost:  │
│         8090        │
└────────────────────┘
```

PocketBase serves the React SPA and the API on the same port — same-origin, so CORS isn't an issue.

---

## Step-by-Step Setup

### 1. Find your PC's LAN IP

**Windows (PowerShell):**
```powershell
ipconfig | Select-String "IPv4"
```

**macOS/Linux:**
```bash
ip addr show | grep "inet " | grep -v 127.0.0.1
# or
hostname -I
```

Note the address (e.g., `192.168.1.105`). Your other devices will use this.

---

### 2. Generate an HMAC secret

The invite system needs a signing key. Generate a 32-byte hex string:

**PowerShell:**
```powershell
-join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Maximum 256) })
```

**Bash:**
```bash
openssl rand -hex 32
```

Copy the 64-character hex output. You'll use it in step 5.

---

### 3. Install dependencies & build the frontend

```powershell
cd frontend
npm install
npm run build
cd ..
```

This creates `frontend/dist/` with the production SPA.

---

### 4. Copy the frontend into PocketBase's static directory

PocketBase serves static files from a `pb_public/` folder next to the binary:

**PowerShell:**
```powershell
New-Item -ItemType Directory -Force -Path backend\pb_public
Copy-Item -Recurse -Force frontend\dist\* backend\pb_public\
```

**Bash:**
```bash
mkdir -p backend/pb_public
cp -r frontend/dist/* backend/pb_public/
```

---

### 5. Build and run the backend

```powershell
cd backend

# Set environment variables
$env:HEARTH_DOMAIN = "localhost"
$env:HMAC_SECRET_CURRENT = "<paste your 64-char hex string from step 2>"

# Build
go build -o hearth.exe .

# Run — 0.0.0.0 listens on all network interfaces (required for LAN access)
.\hearth.exe serve --http=0.0.0.0:8090
```

**Bash equivalent:**
```bash
cd backend
export HEARTH_DOMAIN=localhost
export HMAC_SECRET_CURRENT="<your hex string>"
go build -o hearth .
./hearth serve --http=0.0.0.0:8090
```

---

### 6. First-run admin setup (one time only)

1. Open `http://localhost:8090/_/` in your browser
2. PocketBase prompts you to create an **admin account** (email + password)
3. This is the admin panel — not a Hearth user account
4. The hooks auto-create the collections (users, rooms, messages) on first startup

---

### 7. Create a room (via admin panel)

The app has user registration built in, but rooms must be created manually:

1. Go to `http://localhost:8090/_/` → **Collections** → **rooms**
2. Click **New record**
3. Fill in:
   - `name`: e.g., "The Kitchen"
   - `slug`: e.g., "the-kitchen"
   - `default_ttl`: e.g., `300` (seconds — 5-minute message decay)
4. After users register, add them to `room_members` for that room

---

### 8. Open the app

| Device | URL |
|--------|-----|
| **Server PC** | `http://localhost:8090` |
| **Laptop / Phone** | `http://<your-PC-IP>:8090` (e.g., `http://192.168.1.105:8090`) |

Register user accounts on each device, then navigate to your room. You'll see real-time fading chat.

---

## Troubleshooting

### Laptop can't connect

**Windows Firewall** — Allow port 8090 inbound. Run PowerShell **as Administrator**:

```powershell
New-NetFirewallRule -DisplayName "Hearth PocketBase" -Direction Inbound -Protocol TCP -LocalPort 8090 -Action Allow
```

To remove later:
```powershell
Remove-NetFirewallRule -DisplayName "Hearth PocketBase"
```

### "CORS error" in browser console

This shouldn't happen when accessing via `pb_public/` (same-origin). If you see CORS errors, make sure you're accessing the app through PocketBase's URL (`:8090`), not the Vite dev server (`:5173`).

### Collections not created

If you see empty data after startup, check the PocketBase admin panel (`/_/`). The hooks register collections on the first `OnServe` event. If collections are missing, restart the server — they're created idempotently.

### Messages not appearing in real-time

Ensure both users are in the room's `room_members` list (set via admin panel). SSE subscriptions filter by room membership.

---

## Development Mode (with hot reload)

For active frontend development, use Vite's dev server instead of `pb_public/`:

**Terminal 1 — Backend:**
```powershell
cd backend
$env:HEARTH_DOMAIN = "localhost"
$env:HMAC_SECRET_CURRENT = "<your hex string>"
.\hearth.exe serve --http=0.0.0.0:8090
```

**Terminal 2 — Frontend (Vite dev server):**
```powershell
cd frontend
npm run dev
```

Access via `http://localhost:5173`. Vite proxies `/api/*` and `/_/*` to PocketBase on `:8090` (configured in `vite.config.ts`).

> **Note:** Dev mode only works from the PC running Vite. LAN clients should use the `pb_public/` method (steps 3-4) for testing.

---

## What Works (v0.2 Kindling)

- ✅ User registration & login
- ✅ Real-time chat with fading messages (4-stage CSS decay)
- ✅ Message expiry & garbage collection
- ✅ Presence indicators (who's online)
- ✅ Automatic reconnection (handles Wi-Fi drops)
- ✅ Dark mode + light mode (follows OS preference)
- ✅ HMAC invite tokens (API only — no UI yet)
- ✅ Proof-of-Work bot deterrent

## What Doesn't Work Yet

- 🔲 Voice (The Portal) — v0.3, needs LiveKit server
- 🔲 Invite link UI — API exists, frontend not built
- 🔲 Sound effects — deferred to R-007
- 🔲 Mobile sidebar drawer — needs hamburger menu
- 🔲 Typing indicator broadcast — needs backend custom topic
