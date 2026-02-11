# Sprint 2 — v0.2 "Kindling" (Frontend Shell + Campfire Chat)

> **Sprint Goal:** A working React frontend that connects to the PocketBase backend. Design system implemented. Campfire (fading chat) is the **first user-facing feature**. Users can log in, join a room, send messages, and watch them fade.
>
> **Target:** June 2026
> **Owner:** Builder Agent
> **Research Prerequisites:** ✅ R-001 through R-006 + R-008 complete. Frontend fully unblocked.
> **Dependency:** v0.1 Ember (backend API) must be running.

---

## Research References

Every code pattern in this spec is verified against current documentation. Do NOT use patterns from AI training data — use these reports:

| Report | Relevance | Key Patterns |
|--------|-----------|-------------|
| [R-004](../research/R-004-pocketbase-js-sdk.md) | PocketBase JS SDK — auth, CRUD, real-time | `pb.collection().subscribe()`, SSE reconnect, `AuthProvider`, optimistic updates |
| [R-008](../research/R-008-css-animation-performance.md) | CSS animation performance at scale | `content-visibility: auto`, compositor opacity, `animationend` cleanup, no `will-change` |
| [Master Plan §4](../../vesta_master_plan.md) | Design system — Subtle Warmth palette, typography, motion | Color tokens, font stack, shape language, sound cues |
| [Master Plan §6](../../vesta_master_plan.md) | Frontend engineering — optimistic UI, visual decay, connectivity | Time sync, heartbeat, reconnect patterns |

⚠️ **PocketBase SDK Note:** PocketBase uses **SSE (Server-Sent Events)**, not WebSocket. The SDK abstracts this (`pb.collection().subscribe()`), but debugging uses `GET /api/realtime`, not WS upgrade. SSE is unidirectional — client→server uses regular HTTP POST.

---

## File Tree (Target State After Sprint 2)

```
hearth/
├── backend/                     # (from Sprint 1 — unchanged)
│   ├── main.go
│   ├── go.mod / go.sum
│   └── hooks/                   # All Sprint 1 hooks
├── frontend/
│   ├── index.html               # Vite entry point
│   ├── package.json             # React, Vite, TailwindCSS, PocketBase SDK
│   ├── tsconfig.json            # TypeScript strict mode
│   ├── vite.config.ts           # Dev server proxy, build config
│   ├── tailwind.config.ts       # Subtle Warmth design tokens
│   ├── postcss.config.js        # Tailwind + autoprefixer
│   ├── public/
│   │   └── sounds/              # Foley samples (if R-007 done)
│   ├── src/
│   │   ├── main.tsx             # React root + providers
│   │   ├── App.tsx              # Router shell
│   │   ├── lib/
│   │   │   ├── pocketbase.ts    # Singleton PB client
│   │   │   ├── time-sync.ts     # Server time offset calculation
│   │   │   └── constants.ts     # Config, API URLs
│   │   ├── hooks/
│   │   │   ├── useAuth.ts       # Auth context + hook
│   │   │   ├── useMessages.ts   # Real-time message subscription
│   │   │   ├── usePresence.ts   # Heartbeat + presence display
│   │   │   ├── useTimeSync.ts   # Server time offset hook
│   │   │   └── useReconnect.ts  # SSE reconnect + state resync
│   │   ├── components/
│   │   │   ├── ui/              # Design system primitives
│   │   │   │   ├── Button.tsx
│   │   │   │   ├── Card.tsx
│   │   │   │   ├── Input.tsx
│   │   │   │   ├── Avatar.tsx
│   │   │   │   └── Spinner.tsx
│   │   │   ├── auth/
│   │   │   │   ├── LoginForm.tsx
│   │   │   │   └── AuthGuard.tsx
│   │   │   ├── campfire/
│   │   │   │   ├── CampfireRoom.tsx       # Room container
│   │   │   │   ├── MessageList.tsx        # Scrollable message list
│   │   │   │   ├── MessageBubble.tsx      # Individual fading message
│   │   │   │   ├── MessageInput.tsx       # Compose + send
│   │   │   │   ├── MumblingIndicator.tsx  # Typing indicator
│   │   │   │   └── PresenceBar.tsx        # Who's here
│   │   │   └── layout/
│   │   │       ├── Shell.tsx              # App chrome
│   │   │       └── RoomList.tsx           # Room navigation
│   │   ├── pages/
│   │   │   ├── HomePage.tsx
│   │   │   ├── RoomPage.tsx
│   │   │   └── LoginPage.tsx
│   │   └── styles/
│   │       ├── globals.css        # @tailwind directives + custom properties
│   │       └── campfire.css       # Fading animation keyframes
│   └── .env.example               # VITE_PB_URL, VITE_DOMAIN
├── config/                        # (from Sprint 1 — unchanged)
├── docker/                        # (from Sprint 1 — add Dockerfile.frontend)
│   ├── Dockerfile.pocketbase
│   ├── Dockerfile.frontend        # NEW: multi-stage Vite build
│   └── caddy/
└── docker-compose.yaml            # Updated: add frontend static serving
```

---

## Phase 0: Frontend Scaffolding

**Goal:** Vite + React + TypeScript + TailwindCSS project compiles and renders a "Hello Hearth" screen. Design tokens configured. Dev server proxies API requests to PocketBase.

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **0.1** | K-003 | Scaffold `frontend/` project | `npm run dev` starts Vite on `:5173`. TypeScript strict. React 18+. |
| **0.2** | K-004 | Implement Subtle Warmth design tokens | Tailwind config has custom palette, fonts, shadows, border-radius. Dark mode works. |
| **0.3** | K-005 | Component library: Button, Card, Input, Avatar | All components render correctly. Pillow buttons, soft shadows, rounded shapes. |
| **0.4** | K-006 | Motion primitives | Ease-in-out on all interactive elements. Float-in animation. Squash & stretch on buttons. |

### 0.1 — Scaffold Frontend (K-003)

```bash
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install tailwindcss @tailwindcss/vite pocketbase
```

**File:** `frontend/vite.config.ts`

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
      '/_': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'es2022',
    sourcemap: true,
  },
});
```

**File:** `frontend/tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"]
}
```

**Notes for Builder:**
- Use `@tailwindcss/vite` plugin (v4 approach) — NOT `tailwindcss` PostCSS plugin (v3 approach). Tailwind v4 uses native Vite integration.
- TypeScript `strict: true` is mandatory per project rules. Also enable `noUncheckedIndexedAccess` for extra safety.
- The Vite proxy forwards `/api/*` and `/_/*` to PocketBase, solving CORS in development (production uses Caddy).
- Path alias `@/*` → `src/*` for clean imports.

### 0.2 — Design Tokens: Subtle Warmth (K-004)

**File:** `frontend/src/styles/globals.css`

```css
@import "tailwindcss";

/* === Subtle Warmth Design Tokens === */

@theme {
  /* Color Palette — Dark Mode (default) */
  --color-bg-primary: #2B211E;        /* Espresso */
  --color-bg-secondary: #3E2C29;      /* Warm Charcoal */
  --color-bg-elevated: #4A3632;       /* Raised card */
  --color-bg-input: #352723;          /* Input fields */

  --color-text-primary: #F2E2D9;      /* Cream */
  --color-text-secondary: #B8A69A;    /* Muted cream */
  --color-text-muted: #7A6B62;        /* Very muted */

  --color-accent-amber: #E6A44F;      /* Active / Focus / Ember */
  --color-accent-amber-hover: #D4933E;
  --color-accent-gold: #F0C674;       /* Highlight */
  --color-alert-clay: #C45C3A;        /* Burnt Clay — alerts */
  --color-alert-clay-hover: #A8492E;
  --color-accent-sage: #7A9A7E;       /* Sage Green — secondary accent */
  --color-accent-terracotta: #C47A5A; /* Terracotta */
  --color-accent-slate: #6B8A9A;      /* Slate Blue */

  /* Typography */
  --font-body: 'Inter', system-ui, -apple-system, sans-serif;
  --font-display: 'Merriweather', Georgia, serif;

  /* Shape */
  --radius-sm: 0.5rem;
  --radius-md: 0.75rem;
  --radius-lg: 1rem;
  --radius-xl: 1.5rem;
  --radius-pill: 9999px;

  /* Shadows (candlelight — warm, diffused) */
  --shadow-sm: 0 1px 3px rgba(43, 33, 30, 0.3);
  --shadow-md: 0 4px 12px rgba(43, 33, 30, 0.4);
  --shadow-lg: 0 8px 24px rgba(43, 33, 30, 0.5);
  --shadow-glow: 0 0 20px rgba(230, 164, 79, 0.3);  /* Ember glow */

  /* Spacing scale (generous whitespace) */
  --space-xs: 0.25rem;
  --space-sm: 0.5rem;
  --space-md: 1rem;
  --space-lg: 1.5rem;
  --space-xl: 2rem;
  --space-2xl: 3rem;

  /* Transitions (no linear — always ease-in-out) */
  --ease-default: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-bounce: cubic-bezier(0.34, 1.56, 0.64, 1);
  --duration-fast: 150ms;
  --duration-normal: 250ms;
  --duration-slow: 400ms;
}

/* Light mode override */
@media (prefers-color-scheme: light) {
  @theme {
    --color-bg-primary: #FAF9D1;
    --color-bg-secondary: #F2E2D9;
    --color-bg-elevated: #FFFFFF;
    --color-bg-input: #F5EDE7;
    --color-text-primary: #2B211E;
    --color-text-secondary: #5A4A42;
    --color-text-muted: #8A7A72;
  }
}

/* Base styles */
body {
  background-color: var(--color-bg-primary);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  line-height: 1.6;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
```

**Notes for Builder:**
- Use Tailwind v4 `@theme` directive to define custom tokens. These become Tailwind utility classes automatically (`bg-bg-primary`, `text-accent-amber`, etc.).
- Dark mode is the **default** — Hearth is a "firelit room," not a white office. Light mode is provided via `prefers-color-scheme` media query.
- Import Inter and Merriweather from Google Fonts via `<link>` in `index.html`, or use `@fontsource/inter` + `@fontsource/merriweather` npm packages (preferred — no external requests, privacy-first).
- Shadows use warm brown rgba, NOT gray/black. This creates the "candlelight" effect.

### 0.3 — Component Library (K-005)

#### Button — "Pillow" Style

```tsx
// frontend/src/components/ui/Button.tsx
import { type ButtonHTMLAttributes, type ReactNode } from 'react';

type Variant = 'primary' | 'secondary' | 'ghost';
type Size = 'sm' | 'md' | 'lg';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  children: ReactNode;
}

export function Button({
  variant = 'primary',
  size = 'md',
  className = '',
  children,
  ...props
}: ButtonProps) {
  const base = [
    'inline-flex items-center justify-center',
    'font-medium rounded-pill',
    'transition-all duration-normal ease-default',
    'active:scale-95',               // Squash on press
    'focus-visible:outline-2 focus-visible:outline-offset-2',
    'focus-visible:outline-accent-amber',
    'disabled:opacity-50 disabled:cursor-not-allowed',
  ].join(' ');

  const variants: Record<Variant, string> = {
    primary: 'bg-accent-amber text-bg-primary hover:bg-accent-amber-hover shadow-md hover:shadow-lg',
    secondary: 'bg-bg-elevated text-text-primary border border-bg-elevated hover:bg-bg-secondary',
    ghost: 'text-text-secondary hover:text-text-primary hover:bg-bg-secondary',
  };

  const sizes: Record<Size, string> = {
    sm: 'text-sm px-3 py-1.5',
    md: 'text-base px-5 py-2.5',
    lg: 'text-lg px-7 py-3',
  };

  return (
    <button
      className={`${base} ${variants[variant]} ${sizes[size]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
```

**Key design requirements:**
- `rounded-pill` (border-radius: 9999px) — "lozenge" / "pillow" shape per master plan
- `active:scale-95` — squash on press (Disney "squash & stretch" principle)
- No linear transitions — all interactions use `ease-default`
- Amber/gold as primary action color
- Warm shadows that grow on hover

#### Additional UI Components (K-005)

Builder should create `Card`, `Input`, `Avatar`, and `Spinner` following the same patterns:
- **Card:** `bg-bg-elevated rounded-xl shadow-md` with generous padding
- **Input:** `bg-bg-input rounded-lg border-none` with focus glow (`shadow-glow`)
- **Avatar:** Rounded-full, ember glow ring when user is active/speaking
- **Spinner:** Subtle warm pulsing animation (not a cold spinning wheel)

### 0.4 — Motion Primitives (K-006)

**File:** `frontend/src/styles/globals.css` (append)

```css
/* === Motion Primitives === */

/* Float in — messages and cards enter like a leaf falling */
@keyframes float-in {
  0% {
    opacity: 0;
    transform: translateY(-12px) scale(0.97);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.animate-float-in {
  animation: float-in var(--duration-normal) var(--ease-default) both;
}

/* Squash & stretch for button press */
@keyframes squash {
  0% { transform: scale(1); }
  50% { transform: scale(0.95, 1.02); }
  100% { transform: scale(1); }
}

/* Pulse for loading/active states */
@keyframes warm-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.animate-warm-pulse {
  animation: warm-pulse 2s var(--ease-default) infinite;
}

/* Slide pane for page transitions */
@keyframes slide-in-right {
  0% {
    opacity: 0;
    transform: translateX(24px);
  }
  100% {
    opacity: 1;
    transform: translateX(0);
  }
}

.animate-slide-in {
  animation: slide-in-right var(--duration-slow) var(--ease-default) both;
}
```

**Rules:**
- No `linear` timing function anywhere in the project
- All transitions use `var(--ease-default)` or `var(--ease-bounce)`
- All animated properties MUST be `opacity` or `transform` only (compositor-thread, R-008)
- Use CSS animations, not JS `requestAnimationFrame`, for visual effects

---

## Phase 1: Auth & Connection

**Goal:** User can log in with email/password. PocketBase SDK connects. Auth persists across refreshes. Token auto-refreshes.

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **1.1** | K-010 | PocketBase SDK client + auth integration | Login works. Token stored. User record available in React context. |
| **1.2** | SEC-005 | httpOnly cookie auth store | Tokens NOT in `localStorage`. Custom `AuthStore` implementation. |
| **1.3** | SEC-006 | SSE reconnect auth validation | On `PB_CONNECT`, validate token. Refresh if needed. Redirect to login if invalid. |

### 1.1 — PocketBase Client Setup (K-010)

**File:** `frontend/src/lib/pocketbase.ts`

```typescript
import PocketBase from 'pocketbase';

// Single global instance — NEVER create per-component
const pb = new PocketBase(import.meta.env.VITE_PB_URL || '/');

export default pb;
```

**File:** `frontend/src/hooks/useAuth.ts`

```typescript
import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import pb from '@/lib/pocketbase';
import type { RecordModel } from 'pocketbase';

interface AuthContextType {
  user: RecordModel | null;
  isValid: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState(pb.authStore.record);
  const [isValid, setIsValid] = useState(pb.authStore.isValid);

  useEffect(() => {
    // Listen for all auth state changes
    return pb.authStore.onChange((_token, record) => {
      setUser(record);
      setIsValid(pb.authStore.isValid);
    });
  }, []);

  // Auto-refresh token every 10 minutes
  useEffect(() => {
    const interval = setInterval(async () => {
      if (pb.authStore.isValid) {
        try {
          await pb.collection('users').authRefresh();
        } catch {
          pb.authStore.clear();
        }
      }
    }, 10 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  const login = async (email: string, password: string) => {
    await pb.collection('users').authWithPassword(email, password);
  };

  const logout = () => {
    pb.authStore.clear();
    pb.realtime.disconnect(); // Immediately kill SSE connection
  };

  return (
    <AuthContext.Provider value={{ user, isValid, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
```

**Notes for Builder (from R-004):**
- `pb.authStore.onChange()` is the canonical way to react to auth changes — returns an unsubscribe function.
- Call `pb.realtime.disconnect()` on logout to immediately kill the SSE connection.
- `authRefresh()` extends the token's validity. If it throws, the token is invalid — clear and redirect to login.
- For the `VITE_PB_URL` env var: in production, use `/` (Caddy proxies). In dev, Vite proxy handles it.

### 1.2 — httpOnly Cookie Auth Store (SEC-005)

**Why:** PocketBase's default `LocalAuthStore` puts the JWT in `localStorage`, which is readable by ANY JavaScript on the page — including XSS payloads. Moving to `httpOnly` cookies makes the token invisible to JavaScript entirely.

**Implementation approach:**

```typescript
// This requires a backend endpoint to set/clear the cookie.
// PocketBase doesn't natively support httpOnly cookie auth,
// so we need a thin proxy hook.

// Backend hook: POST /api/hearth/auth/login
// 1. Proxies to PocketBase auth
// 2. Sets httpOnly cookie with the JWT
// 3. Returns user record (without token) to frontend

// Backend hook: POST /api/hearth/auth/logout
// 1. Clears the httpOnly cookie
// 2. Returns 200

// Frontend: Custom AuthStore that doesn't persist token client-side
import PocketBase, { AsyncAuthStore } from 'pocketbase';

const store = new AsyncAuthStore({
  save: async (_serialized) => {
    // Token is in httpOnly cookie — don't save to localStorage
    // Just keep in-memory for the SDK to function
  },
  clear: async () => {
    await fetch('/api/hearth/auth/logout', { method: 'POST' });
  },
  initial: '', // Token comes from cookie on each request
});

const pb = new PocketBase('/', store);
```

**⚠️ Builder decision needed:** This is a significant arch change — the backend needs new auth proxy hooks. Builder should evaluate complexity and may defer to a later sprint if the scope is too large. At minimum, document the `localStorage` usage as a known SEC-005 item.

**Fallback if deferred:** Keep `LocalAuthStore` but add `token` to the Content Security Policy (already configured in Caddy via E-041) so inline scripts can't steal it. This isn't as strong but reduces risk.

### 1.3 — SSE Reconnect Validation (SEC-006)

**File:** `frontend/src/hooks/useReconnect.ts`

```typescript
import { useEffect, useRef, useCallback } from 'react';
import pb from '@/lib/pocketbase';

/**
 * Handles SSE reconnect state synchronization.
 * On every PB_CONNECT after the first, re-fetches data to catch missed events.
 * From R-004: "Reconnect does NOT replay events — missed events are lost."
 */
export function useReconnect(onResync: () => void) {
  const isFirstConnect = useRef(true);

  useEffect(() => {
    const unsubscribe = pb.realtime.subscribe('PB_CONNECT', async () => {
      if (isFirstConnect.current) {
        isFirstConnect.current = false;
        return; // First connect — no resync needed
      }

      // Non-initial connect: we missed events during disconnection
      // Validate auth is still good
      try {
        await pb.collection('users').authRefresh();
      } catch {
        // Auth expired during disconnection
        pb.authStore.clear();
        return;
      }

      // Re-fetch current state
      onResync();
    });

    return () => { unsubscribe.then(fn => fn()); };
  }, [onResync]);
}
```

**Notes for Builder (from R-004):**
- `PB_CONNECT` fires on EVERY connection, including the first. Use a `useRef` flag to distinguish first-connect from reconnect.
- After reconnect, ALL subscriptions are automatically re-established by the SDK — BUT missed events during disconnection are NOT replayed.
- The `onResync` callback should re-fetch the message list and presence state for the current room.

---

## Phase 2: Campfire (Fading Chat)

**Goal:** Users see a real-time message feed where messages visually fade from 100% → 0% opacity over their TTL. Messages float in, fade, and disappear. The server's message GC and the client's CSS decay align.

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **2.1** | K-011 | Chat message list with real-time subscription | Messages appear in real-time. Scroll is anchored to bottom. `delete` events remove messages. |
| **2.2** | K-012 | CSS transparency decay engine | 4-stage fade (Fresh→Fading→Echo→Gone). `animationend` removes from DOM. `content-visibility: auto` on offscreen messages. |
| **2.3** | K-013 | Time sync with server | Client calculates offset from `Date` header. `--age-offset` is accurate ±500ms. |
| **2.4** | K-014 | Optimistic message sending | Message appears instantly. Reverts with error toast on failure. |
| **2.5** | K-015 | "Mumbling" typing indicator | Blurred waveform/scribble animation. Not "User is typing..." |
| **2.6** | K-016 | Exponential backoff reconnection | Uses PB SDK built-in. `useReconnect` hook triggers resync. Visual "reconnecting..." indicator. |
| **2.7** | K-017 | Heartbeat presence display | 30s heartbeat. Shows active users in room. Offline after 2 missed beats. |

### 2.1 — Real-time Message List (K-011)

**File:** `frontend/src/hooks/useMessages.ts`

```typescript
import { useEffect, useState, useCallback, useRef } from 'react';
import pb from '@/lib/pocketbase';
import { useReconnect } from './useReconnect';

export interface Message {
  id: string;
  text: string;
  room: string;
  author: string;
  expires_at: string;
  created: string;
  expand?: {
    author?: { display_name: string; avatar_url: string };
  };
}

export function useMessages(roomId: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch current messages for this room
  const fetchMessages = useCallback(async () => {
    if (!roomId) return;
    try {
      const result = await pb.collection('messages').getList<Message>(1, 200, {
        filter: `room = "${roomId}"`,
        sort: 'created',
        expand: 'author',
        requestKey: `messages-${roomId}`, // Prevent duplicate requests
      });
      setMessages(result.items);
    } catch (err) {
      console.error('Failed to fetch messages:', err);
    } finally {
      setIsLoading(false);
    }
  }, [roomId]);

  // Initial fetch
  useEffect(() => {
    fetchMessages();
  }, [fetchMessages]);

  // Real-time subscription
  useEffect(() => {
    if (!roomId) return;

    const unsubPromise = pb.collection('messages').subscribe<Message>('*', (data) => {
      if (data.record.room !== roomId) return; // Filter to current room

      switch (data.action) {
        case 'create':
          setMessages((prev) => [...prev, data.record]);
          break;
        case 'update':
          setMessages((prev) =>
            prev.map((m) => (m.id === data.record.id ? data.record : m))
          );
          break;
        case 'delete':
          // Server GC deleted an expired message — remove from state
          setMessages((prev) => prev.filter((m) => m.id !== data.record.id));
          break;
      }
    });

    return () => {
      unsubPromise.then((unsub) => unsub());
    };
  }, [roomId]);

  // Reconnect: re-fetch everything (missed events are not replayed)
  useReconnect(fetchMessages);

  // Remove a message client-side (e.g., after animationend)
  const removeMessage = useCallback((id: string) => {
    setMessages((prev) => prev.filter((m) => m.id !== id));
  }, []);

  // Optimistic send
  const sendMessage = useCallback(
    async (text: string) => {
      const tempId = `temp-${Date.now()}`;
      const optimistic: Message = {
        id: tempId,
        text,
        room: roomId,
        author: pb.authStore.record?.id ?? '',
        expires_at: new Date(Date.now() + 300_000).toISOString(), // 5 min default
        created: new Date().toISOString(),
      };

      // Immediately add to state
      setMessages((prev) => [...prev, optimistic]);

      try {
        const real = await pb.collection('messages').create<Message>({
          text,
          room: roomId,
          // Server enforces expires_at via TTL field
        });
        // Replace optimistic with real record
        setMessages((prev) =>
          prev.map((m) => (m.id === tempId ? real : m))
        );
      } catch {
        // Revert on failure
        setMessages((prev) => prev.filter((m) => m.id !== tempId));
        // TODO: show error toast
      }
    },
    [roomId]
  );

  return { messages, isLoading, sendMessage, removeMessage };
}
```

**Notes for Builder (from R-004):**
- Use `requestKey` to prevent the SDK from cancelling duplicate fetches when the user switches rooms quickly.
- The `subscribe('*', cb)` callback receives ALL changes in the collection — filter by `roomId` client-side.
- `expand: 'author'` tells PocketBase to inline the author's record (display_name, avatar) in the response.
- On `delete` event: the server's GC cron deleted an expired message. CSS may have already made it invisible, but we still need to clean up React state.

### 2.2 — CSS Transparency Decay Engine (K-012)

**This is the heart of Campfire.** Per R-008, the approach is: one CSS animation per message, pure `opacity` on the compositor thread, `content-visibility: auto` for offscreen optimization, `animationend` for DOM cleanup.

**File:** `frontend/src/styles/campfire.css`

```css
/* === Campfire: 4-Stage Transparency Decay === */

@keyframes campfire-fade {
  0%   { opacity: 1; }     /* Fresh — active conversation */
  40%  { opacity: 0.5; }   /* Fading — past context */
  80%  { opacity: 0.1; }   /* Echo — visual texture */
  100% { opacity: 0; }     /* Gone — vanished */
}

.campfire-message {
  /* Browser-native virtualization (R-008 Strategy 2) */
  content-visibility: auto;
  contain-intrinsic-size: auto 80px;

  /* Single CSS animation — compositor-thread only (R-008 Strategy 3) */
  animation: campfire-fade var(--fade-duration) ease-out forwards;
  animation-delay: var(--age-offset);

  /* Float-in entrance */
  transform-origin: center;
}

/* Optimistic messages don't fade until confirmed */
.campfire-message[data-optimistic="true"] {
  animation: none;
  opacity: 0.85; /* Slightly dim to indicate "sending" */
}
```

**File:** `frontend/src/components/campfire/MessageBubble.tsx`

```tsx
import { useCallback, type CSSProperties } from 'react';
import { useTimeSync } from '@/hooks/useTimeSync';
import type { Message } from '@/hooks/useMessages';

interface Props {
  message: Message;
  onGone: (id: string) => void;
}

export function MessageBubble({ message, onGone }: Props) {
  const { getServerNow } = useTimeSync();

  const isOptimistic = message.id.startsWith('temp-');
  const serverNow = getServerNow();

  // Calculate animation timing
  const createdMs = new Date(message.created).getTime();
  const expiresMs = new Date(message.expires_at).getTime();
  const fadeDuration = (expiresMs - createdMs) / 1000; // Total TTL in seconds
  const ageSeconds = (serverNow - createdMs) / 1000;   // How old the message is
  const ageOffset = -ageSeconds;                         // Negative = start mid-fade

  // If message is already past its TTL, don't render
  if (ageSeconds >= fadeDuration) return null;

  const handleAnimationEnd = useCallback(() => {
    onGone(message.id);
  }, [message.id, onGone]);

  const style: CSSProperties = {
    '--fade-duration': `${fadeDuration}s`,
    '--age-offset': `${ageOffset}s`,
  } as CSSProperties;

  return (
    <div
      className="campfire-message animate-float-in"
      style={style}
      data-optimistic={isOptimistic}
      onAnimationEnd={(e) => {
        // Only handle the fade animation, not the float-in
        if (e.animationName === 'campfire-fade') {
          handleAnimationEnd();
        }
      }}
    >
      <div className="flex items-start gap-3 p-3">
        <div className="font-medium text-sm text-accent-amber">
          {message.expand?.author?.display_name ?? 'Wanderer'}
        </div>
        <div className="text-text-primary text-base">
          {message.text}
        </div>
      </div>
    </div>
  );
}
```

**File:** `frontend/src/components/campfire/MessageList.tsx`

```tsx
import { useEffect, useRef } from 'react';
import { MessageBubble } from './MessageBubble';
import type { Message } from '@/hooks/useMessages';

interface Props {
  messages: Message[];
  onMessageGone: (id: string) => void;
}

export function MessageList({ messages, onMessageGone }: Props) {
  const listRef = useRef<HTMLDivElement>(null);
  const shouldAutoScroll = useRef(true);

  // Auto-scroll to bottom on new messages (if already at bottom)
  useEffect(() => {
    const el = listRef.current;
    if (!el || !shouldAutoScroll.current) return;
    el.scrollTop = el.scrollHeight;
  }, [messages.length]);

  // Track whether user has scrolled up
  const handleScroll = () => {
    const el = listRef.current;
    if (!el) return;
    const threshold = 100; // px from bottom
    shouldAutoScroll.current =
      el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
  };

  return (
    <div
      ref={listRef}
      onScroll={handleScroll}
      className="flex-1 overflow-y-auto overscroll-contain p-4 space-y-2"
    >
      {messages.map((msg) => (
        <MessageBubble
          key={msg.id}
          message={msg}
          onGone={onMessageGone}
        />
      ))}
    </div>
  );
}
```

**Key architecture points (from R-008):**
- `content-visibility: auto` on each message = browser-native virtualization. Offscreen messages skip rendering entirely.
- `contain-intrinsic-size: auto 80px` = estimated height for scroll calculations. `auto` keyword remembers actual height after first render.
- ONE `@keyframes` animation applied via CSS custom properties per message. No JS animation loop.
- `animation-delay` negative value starts the animation mid-fade for old messages on page load.
- `animationend` event fires when a message reaches 0% opacity → remove from React state → DOM cleanup.
- `onAnimationEnd` filters by `animationName` to ignore the float-in entrance animation.
- **DO NOT** use `will-change: opacity`. The browser auto-promotes animated elements to compositor layers.

### 2.3 — Time Sync (K-013)

**File:** `frontend/src/hooks/useTimeSync.ts`

```typescript
import { useEffect, useRef, useCallback } from 'react';
import pb from '@/lib/pocketbase';

/**
 * Calculates offset between client clock and server clock
 * using the Date header from PocketBase API responses.
 *
 * serverTime ≈ clientTime + offset
 */
export function useTimeSync() {
  const offsetMs = useRef(0);

  useEffect(() => {
    async function sync() {
      const before = Date.now();
      // Use a lightweight endpoint to measure round-trip
      const response = await fetch(
        (import.meta.env.VITE_PB_URL || '') + '/api/health'
      );
      const after = Date.now();
      const rtt = after - before;

      const serverDateStr = response.headers.get('Date');
      if (!serverDateStr) return;

      const serverTime = new Date(serverDateStr).getTime();
      // Estimate: server sent the header at the midpoint of the RTT
      const estimatedClientTimeAtServer = before + rtt / 2;
      offsetMs.current = serverTime - estimatedClientTimeAtServer;
    }

    sync();
    // Re-sync every 5 minutes
    const interval = setInterval(sync, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  const getServerNow = useCallback(() => {
    return Date.now() + offsetMs.current;
  }, []);

  return { getServerNow, offsetMs: offsetMs.current };
}
```

**Notes for Builder:**
- The `Date` HTTP header is present in all responses (PocketBase/Caddy both set it).
- RTT/2 estimation is ~500ms accurate, which is fine for visual fading (humans can't perceive <1s opacity differences).
- Falls back gracefully if `Date` header is missing (offset stays 0 — client clock used directly).

### 2.4 — Optimistic Sending (K-014)

Already implemented in `useMessages.ts` above (see `sendMessage`). Key points:
- Temporary message gets `id: "temp-{timestamp}"`
- `data-optimistic="true"` pauses the fade animation and dims to 85% opacity
- On server confirmation: replace temp with real record (which starts fading normally)
- On server rejection: remove temp and show error toast

**Error toast component pattern:**
```tsx
// Builder should implement a simple toast system, e.g.:
// - toast queue in React context
// - Auto-dismiss after 3 seconds
// - Warm clay color (#C45C3A) for errors
// - Float-in animation from bottom
```

### 2.5 — "Mumbling" Typing Indicator (K-015)

The master plan specifies: "Instead of 'User is typing...' show a blurred waveform or abstract 'scribbles' indicating rhythm, length, and intensity without revealing content."

**File:** `frontend/src/components/campfire/MumblingIndicator.tsx`

```tsx
interface Props {
  typingUsers: string[]; // display names of currently typing users
}

export function MumblingIndicator({ typingUsers }: Props) {
  if (typingUsers.length === 0) return null;

  return (
    <div className="flex items-center gap-2 px-4 py-2 text-text-muted text-sm">
      <div className="mumbling-animation flex gap-0.5">
        {/* Three blurred "scribble" bars that undulate */}
        <span className="mumble-bar" style={{ animationDelay: '0ms' }} />
        <span className="mumble-bar" style={{ animationDelay: '150ms' }} />
        <span className="mumble-bar" style={{ animationDelay: '300ms' }} />
      </div>
      <span className="blur-[1px]">
        {typingUsers.length === 1
          ? `${typingUsers[0]} is thinking...`
          : `${typingUsers.length} people murmuring...`
        }
      </span>
    </div>
  );
}
```

**CSS for mumbling:**
```css
.mumble-bar {
  width: 3px;
  height: 12px;
  background-color: var(--color-text-muted);
  border-radius: var(--radius-pill);
  animation: mumble 1.2s var(--ease-default) infinite;
}

@keyframes mumble {
  0%, 100% { transform: scaleY(0.4); opacity: 0.3; }
  50%      { transform: scaleY(1);   opacity: 0.7; }
}
```

**Implementation notes:**
- Typing state is sent via PocketBase custom topic subscription or a debounced API call
- Debounce: send "typing" on keydown, clear after 3 seconds of no keystrokes
- The blurred text and undulating bars create a "murmuring" feeling rather than "performing"

### 2.6 — Reconnection (K-016)

PocketBase's SSE SDK has built-in reconnection with backoff: `[200, 300, 500, 1000, 1200, 1500, 2000]ms` (from R-004). The `useReconnect` hook (Phase 1.3) handles state resync after reconnect.

**Visual indicator:**

```tsx
// In Shell.tsx or a dedicated component
import { useConnectionState } from '@/hooks/useReconnect';

export function ConnectionIndicator() {
  const { isConnected, isReconnecting } = useConnectionState();

  if (isConnected) return null;

  return (
    <div className="fixed top-0 inset-x-0 bg-alert-clay text-text-primary text-center py-1 text-sm animate-float-in z-50">
      {isReconnecting ? 'Reconnecting...' : 'Offline'}
    </div>
  );
}
```

**Builder extends `useReconnect` hook to expose connection state:**
```typescript
// Track via pb.realtime.onDisconnect callback (from R-004)
pb.realtime.onDisconnect = () => setIsConnected(false);
// PB_CONNECT → setIsConnected(true)
```

### 2.7 — Presence Display (K-017)

**File:** `frontend/src/hooks/usePresence.ts`

```typescript
import { useEffect, useState, useCallback } from 'react';
import pb from '@/lib/pocketbase';

interface PresenceEntry {
  user_id: string;
  display_name: string;
  last_seen: number;
}

export function usePresence(roomId: string) {
  const [presentUsers, setPresentUsers] = useState<PresenceEntry[]>([]);

  // Send heartbeat every 30 seconds
  useEffect(() => {
    if (!roomId) return;

    const beat = () => {
      pb.send(`/api/hearth/presence/heartbeat`, {
        method: 'POST',
        body: { room_id: roomId },
      }).catch(() => { /* offline — will retry on reconnect */ });
    };

    beat(); // Immediate first heartbeat
    const interval = setInterval(beat, 30_000);
    return () => clearInterval(interval);
  }, [roomId]);

  // Poll presence list (complements heartbeat)
  useEffect(() => {
    if (!roomId) return;

    const fetchPresence = async () => {
      try {
        const data = await pb.send(`/api/hearth/presence/${roomId}`, {
          method: 'GET',
        });
        setPresentUsers(data.users ?? []);
      } catch { /* noop */ }
    };

    fetchPresence();
    const interval = setInterval(fetchPresence, 15_000);
    return () => clearInterval(interval);
  }, [roomId]);

  return { presentUsers };
}
```

**File:** `frontend/src/components/campfire/PresenceBar.tsx`

```tsx
import type { PresenceEntry } from '@/hooks/usePresence';
import { Avatar } from '@/components/ui/Avatar';

interface Props {
  users: PresenceEntry[];
}

export function PresenceBar({ users }: Props) {
  return (
    <div className="flex items-center gap-2 px-4 py-2 border-b border-bg-elevated">
      <span className="text-text-muted text-sm">
        {users.length} {users.length === 1 ? 'person' : 'people'} here
      </span>
      <div className="flex -space-x-2">
        {users.slice(0, 8).map((u) => (
          <Avatar
            key={u.user_id}
            name={u.display_name}
            className="ring-2 ring-bg-primary"
            size="sm"
          />
        ))}
        {users.length > 8 && (
          <span className="text-text-muted text-xs ml-2">
            +{users.length - 8}
          </span>
        )}
      </div>
    </div>
  );
}
```

---

## Phase 3: Security Hardening (Frontend)

**Goal:** Harden the frontend against XSS, CSRF, and token theft. Frontend security items from the Security Concerns Tracker.

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **3.1** | SEC-005 | httpOnly cookie auth | Covered in Phase 1.2. Tokens not in `localStorage`. |
| **3.2** | SEC-006 | SSE reconnect auth | Covered in Phase 1.3. Token validated on reconnect. |

See Phase 1 for implementation details.

---

## Phase 4: Polish & Mobile

**Goal:** The app feels like a product, not a prototype. Mobile-first responsive. Code-split for fast loads.

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **4.1** | K-023 | Mobile-first responsive layout | Breakpoints: mobile (default), tablet (768px), desktop (1024px). Campfire works on a phone. |
| **4.2** | K-024 | Code-splitting via React.lazy | Route-level splits. Initial bundle <150KB gzipped. |
| **4.3** | — | Login/register page | Clean login form with pillow button. Error states. |
| **4.4** | — | Room list / navigation | User can see their rooms and switch between them. |
| **4.5** | — | Dockerfile for frontend | Multi-stage build. Static files served via Caddy. |

### 4.1 — Mobile-First Layout (K-023)

**Breakpoint strategy:**
```css
/* Default: mobile (< 768px) — single column */
/* md: tablet (>= 768px) — sidebar + content */
/* lg: desktop (>= 1024px) — wider sidebar, richer presence bar */
```

**Shell layout:**
```tsx
// Mobile: room list is a slide-out drawer
// Tablet+: room list is a fixed sidebar
export function Shell({ children }: { children: ReactNode }) {
  return (
    <div className="flex h-screen bg-bg-primary">
      {/* Room sidebar — hidden on mobile, visible on md+ */}
      <aside className="hidden md:flex md:w-64 lg:w-72 flex-col border-r border-bg-elevated">
        <RoomList />
      </aside>

      {/* Main content */}
      <main className="flex-1 flex flex-col min-w-0">
        {children}
      </main>
    </div>
  );
}
```

### 4.2 — Code Splitting (K-024)

```tsx
// frontend/src/App.tsx
import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Spinner } from '@/components/ui/Spinner';
import { AuthGuard } from '@/components/auth/AuthGuard';

const LoginPage = lazy(() => import('@/pages/LoginPage'));
const HomePage = lazy(() => import('@/pages/HomePage'));
const RoomPage = lazy(() => import('@/pages/RoomPage'));

export default function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<Spinner />}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<AuthGuard />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/room/:roomId" element={<RoomPage />} />
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
```

### 4.5 — Frontend Dockerfile

**File:** `docker/Dockerfile.frontend`

```dockerfile
# Multi-stage build
FROM node:22-alpine AS builder
WORKDIR /build
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Static files — served directly by Caddy (no separate server)
# Copy dist/ into Caddy's file_server root
FROM scratch
COPY --from=builder /build/dist /frontend-dist
```

**Notes for Builder:**
- The frontend is static files after build — no Node.js server in production
- Caddy serves the files via `file_server` directive (add to `caddy.yaml`)
- Alternatively, copy `dist/` into the PocketBase container's `pb_public/` directory — PocketBase serves static files natively

---

## Phase 5: Sound & Ambience (Deferred)

Phases K-020 through K-022 (sound library, interaction sounds, generative ambience) are **deferred to a later sprint** pending R-007 completion. They are polish items that don't block the core Campfire experience.

If R-007 completes before Sprint 2 build work reaches Phase 4, sounds can be integrated then. Otherwise, they move to a v0.2.1 patch.

---

## Dependencies & NPM Packages

| Package | Purpose | Size (gzipped) |
|---------|---------|--------|
| `react` + `react-dom` | UI framework | ~44KB |
| `pocketbase` | Backend SDK (auth, CRUD, SSE) | ~12KB |
| `react-router-dom` | Client-side routing | ~12KB |
| `@fontsource/inter` | Body font (self-hosted, no Google request) | ~100KB (subset) |
| `@fontsource/merriweather` | Display font (self-hosted) | ~80KB (subset) |

**Not needed (per R-008):**
- ❌ `@tanstack/react-virtual` — `content-visibility: auto` handles virtualization
- ❌ `react-virtuoso` — Same reason; defer unless >500 DOM elements measured
- ❌ `framer-motion` — CSS animations handle all motion needs
- ❌ `gsap` — Same; compositor-thread CSS is sufficient

**Total initial bundle estimate:** <150KB gzipped (without fonts).

---

## Acceptance Criteria (Sprint 2 Complete)

- [ ] `npm run dev` starts the frontend on `:5173`
- [ ] `npm run build` produces a static `dist/` folder
- [ ] TypeScript strict mode — zero errors
- [ ] User can register and log in with email/password
- [ ] Auth token persists across page refreshes
- [ ] User sees their rooms and can navigate between them
- [ ] Messages appear in real-time (SSE subscription working)
- [ ] New messages float in with ease-out animation
- [ ] Messages visually fade: Fresh (100%) → Fading (50%) → Echo (10%) → Gone (0%)
- [ ] Old messages on page load start mid-fade (negative `animation-delay`)
- [ ] `animationend` removes fully faded messages from DOM
- [ ] Server GC `delete` events remove messages from UI
- [ ] "Mumbling" typing indicator shows when others are typing
- [ ] Presence bar shows active users in room
- [ ] 30-second heartbeat keeps presence alive
- [ ] Reconnection shows visual indicator and resyncs state
- [ ] Mobile layout works on 375px-wide screen
- [ ] Dark mode is default; light mode responds to OS preference
- [ ] All interactive elements use ease-in-out transitions (no linear)
- [ ] Bundle size <150KB gzipped (excluding fonts)

---

## Estimated Effort

| Phase | Tasks | Estimate |
|-------|-------|----------|
| Phase 0 — Scaffolding | K-003, K-004, K-005, K-006 | 3 days |
| Phase 1 — Auth & Connection | K-010, SEC-005, SEC-006 | 2 days |
| Phase 2 — Campfire | K-011 through K-017 | 5 days |
| Phase 3 — Security | (covered in Phase 1) | — |
| Phase 4 — Polish & Mobile | K-023, K-024, pages, Docker | 3 days |
| **Total** | | **~13 days** |

---

## Ready for Implementation

**Feature:** v0.2 Kindling — Frontend Shell + Campfire Chat
**Spec:** `docs/specs/sprint-2-kindling.md`
**Complexity:** L (Large)

**Key Points for Builder:**
1. **CSS fading is compositor-thread only** — never animate properties other than `opacity` and `transform`. Use `content-visibility: auto` (not JS virtualization) for offscreen messages. (R-008)
2. **PocketBase uses SSE, not WebSocket.** `PB_CONNECT` fires on every reconnect — use it to resync state, because missed events are NOT replayed. (R-004)
3. **SEC-005 (httpOnly cookies) may need a backend hook** — evaluate complexity and defer if needed. Document the `localStorage` risk if deferred.
4. **Dark mode is the default.** Hearth is warm and dim, not clinical and bright.
5. **No external font requests** — use `@fontsource` packages. Privacy-first means no Google Fonts CDN.

**Files to Create:** See file tree above (~25 files in `frontend/`)
**Files to Modify:** `docker-compose.yaml` (add frontend serving), `config/caddy.yaml` (add `file_server` route)

**Questions Resolved:**
- Q: Will CSS animations hold up at 200 messages? → A: Yes (R-008). Compositor handles it. `content-visibility: auto` is the key optimization.
- Q: Need JS virtualization library? → A: No. Defer TanStack Virtual unless >500 DOM elements.
- Q: Which fonts? → A: Inter (body) + Merriweather (display), self-hosted via `@fontsource`.

**Deferred Decisions (Builder can decide):**
- React Router vs. TanStack Router (either works; Router is simpler)
- Toast library: build minimal or use a lightweight lib (sonner, react-hot-toast)
- SEC-005 scope: full httpOnly cookie pipeline vs. defer with docs
