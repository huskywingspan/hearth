# R-004: PocketBase JavaScript SDK — Real-time Integration Guide

> **Status:** Complete  
> **Date:** 2026-02-10  
> **Priority:** High | **Blocks:** K-002, K-010  
> **Source:** PocketBase official docs (`pocketbase.io/docs/`), `pocketbase/js-sdk` GitHub repository

---

## Summary

The PocketBase JS SDK (`pocketbase` npm package) provides a complete client for auth, record CRUD, file handling, and **real-time subscriptions via Server-Sent Events (SSE)**. This document covers everything needed to integrate PocketBase into Hearth's React frontend, with patterns for hooks, subscription lifecycle, auth management, and optimistic updates.

---

## 1. Installation & Client Setup

```bash
npm install pocketbase
```

### Client Initialization

```typescript
import PocketBase from 'pocketbase';

// Single global instance — do NOT create per-component
const pb = new PocketBase('https://hearth.example');

export default pb;
```

**Key properties:**
- `pb.authStore` — Manages auth state (token, user record)
- `pb.collection('name')` — Returns a `RecordService` for CRUD + subscriptions
- `pb.realtime` — Direct access to the `RealtimeService`

### Auth Store Options

| Store | Use Case | Persistence |
|-------|----------|-------------|
| `LocalAuthStore` (default) | Web browsers | `localStorage` |
| `AsyncAuthStore` | React Native, SSR, custom | Custom `save`/`clear` callbacks |

```typescript
// Custom auth store (e.g., for encrypted storage)
import PocketBase, { AsyncAuthStore } from 'pocketbase';

const store = new AsyncAuthStore({
  save: async (serialized) => secureStorage.set('pb_auth', serialized),
  clear: async () => secureStorage.remove('pb_auth'),
  initial: await secureStorage.get('pb_auth'),
});

const pb = new PocketBase('https://hearth.example', store);
```

---

## 2. Authentication

### Auth Methods

```typescript
// Email/password login
const authData = await pb.collection('users').authWithPassword(
  'user@example.com',
  'password123'
);
// authData.token, authData.record

// OAuth2 (opens popup)
const authData = await pb.collection('users').authWithOAuth2({
  provider: 'github',
});

// Refresh token (validates + extends session)
const authData = await pb.collection('users').authRefresh();

// Logout
pb.authStore.clear();
```

### Auth State Monitoring

```typescript
// React hook for auth state
function useAuth() {
  const [user, setUser] = useState(pb.authStore.record);
  const [isValid, setIsValid] = useState(pb.authStore.isValid);

  useEffect(() => {
    // Fires on any auth change (login, logout, token refresh)
    const unsubscribe = pb.authStore.onChange((token, record) => {
      setUser(record);
      setIsValid(pb.authStore.isValid);
    });
    return unsubscribe;
  }, []);

  return { user, isValid, token: pb.authStore.token };
}
```

### Auth Store API

| Property/Method | Type | Description |
|----------------|------|-------------|
| `authStore.token` | `string` | Current JWT token |
| `authStore.record` | `RecordModel \| null` | Current user record |
| `authStore.isValid` | `boolean` | Token exists and not expired (checks `exp` claim) |
| `authStore.isSuperuser` | `boolean` | Whether authenticated as superuser |
| `authStore.onChange(cb)` | `() => void` | Subscribe to auth changes; returns unsubscribe function |
| `authStore.save(token, record)` | `void` | Manually update auth state |
| `authStore.clear()` | `void` | Clear auth (logout) |

---

## 3. Record CRUD

### Basic Operations

```typescript
interface Message {
  id: string;
  text: string;
  room: string;
  author: string;
  expires_at: string;
  created: string;
}

// Create
const record = await pb.collection('messages').create<Message>({
  text: 'Hello campfire!',
  room: 'ROOM_ID',
  author: pb.authStore.record?.id,
  expires_at: new Date(Date.now() + 3600000).toISOString(),
});

// Read one
const record = await pb.collection('messages').getOne<Message>('RECORD_ID');

// Read list (paginated)
const result = await pb.collection('messages').getList<Message>(1, 50, {
  filter: 'room = "ROOM_ID"',
  sort: '-created',
  expand: 'author',
});
// result.items, result.totalItems, result.totalPages

// Read all (auto-paginates)
const all = await pb.collection('messages').getFullList<Message>({
  filter: 'room = "ROOM_ID"',
  sort: '-created',
});

// Update
const updated = await pb.collection('messages').update<Message>('RECORD_ID', {
  text: 'Edited message',
});

// Delete
await pb.collection('messages').delete('RECORD_ID');
```

### Auto-Cancellation

The SDK automatically cancels duplicate pending requests to the same endpoint. To disable for a specific call:

```typescript
const result = await pb.collection('messages').getList(1, 50, {
  requestKey: null, // Disable auto-cancel for this request
});
```

---

## 4. Real-time Subscriptions (SSE)

### Architecture

PocketBase's real-time system uses **Server-Sent Events (SSE)**, NOT WebSocket. The client connects to `/api/realtime` via `EventSource`. This is important:

| Feature | SSE (PocketBase) | WebSocket |
|---------|-------------------|-----------|
| Direction | Server → Client only | Bidirectional |
| Protocol | HTTP/1.1 or HTTP/2 | Custom upgrade |
| Auto-reconnect | Built-in (browser + SDK) | Must implement |
| Proxy compatible | Yes (standard HTTP) | Requires proxy config |
| Battery impact | Lower (no keepalive frames) | Higher |

### Subscribe to Record Changes

```typescript
// Subscribe to all changes in a collection
const unsubscribe = await pb.collection('messages').subscribe('*', (data) => {
  // data.action: 'create' | 'update' | 'delete'
  // data.record: the full record object
  console.log(data.action, data.record);
});

// Subscribe to a specific record
const unsubscribe = await pb.collection('messages').subscribe('RECORD_ID', (data) => {
  console.log('Record changed:', data.action, data.record);
});

// Unsubscribe
unsubscribe(); // or:
await pb.collection('messages').unsubscribe('*');
await pb.collection('messages').unsubscribe('RECORD_ID');
await pb.collection('messages').unsubscribe(); // all subscriptions for this collection
```

### Subscription Data Shape

```typescript
interface RecordSubscription<T> {
  action: string;  // 'create' | 'update' | 'delete'
  record: T;       // Full record data
}
```

### Custom Topic Subscriptions

For non-record real-time messaging (e.g., presence updates, typing indicators):

```typescript
// Subscribe to custom topic
const unsubscribe = await pb.realtime.subscribe('custom-topic', (data) => {
  console.log('Custom event:', data);
});

// Server-side (Go) sends to custom topic:
// app.SubscriptionsBroker().Send("custom-topic", data)
```

### Connection Lifecycle Events

```typescript
// Detect initial connection and reconnections
await pb.realtime.subscribe('PB_CONNECT', (data) => {
  // Fires on first connect AND every reconnect
  // data.clientId — unique client ID for this connection
  console.log('Connected with client ID:', data.clientId);

  // IMPORTANT: Re-fetch data after reconnect to sync missed events
  refreshData();
});

// Detect disconnection
pb.realtime.onDisconnect = () => {
  console.log('Disconnected from PocketBase');
  showOfflineIndicator();
};
```

### Auto-Reconnection

The SDK automatically reconnects with predefined intervals:

```
[200ms, 300ms, 500ms, 1000ms, 1200ms, 1500ms, 2000ms] → then repeats from 200ms
```

- `maxReconnectAttempts`: `Infinity` (never gives up)
- On reconnect, `PB_CONNECT` fires again
- **Critical:** Reconnect does NOT replay missed events. You must re-fetch data after detecting a `PB_CONNECT` event that isn't the initial connection.

---

## 5. React Integration Patterns

### Subscription Hook

```typescript
import { useEffect, useCallback, useRef } from 'react';
import pb from '../lib/pocketbase';
import type { RecordSubscription } from 'pocketbase';

/**
 * Subscribe to a PocketBase collection with automatic cleanup.
 * Re-subscribes when roomId changes.
 */
function useRealtimeMessages(roomId: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const isInitialConnect = useRef(true);

  // Initial fetch
  useEffect(() => {
    if (!roomId) return;

    pb.collection('messages')
      .getFullList<Message>({
        filter: `room = "${roomId}"`,
        sort: '-created',
      })
      .then(setMessages);
  }, [roomId]);

  // Real-time subscription
  useEffect(() => {
    if (!roomId) return;

    const unsubscribe = pb.collection('messages').subscribe<Message>(
      '*',
      (data: RecordSubscription<Message>) => {
        if (data.record.room !== roomId) return;

        switch (data.action) {
          case 'create':
            setMessages(prev => [data.record, ...prev]);
            break;
          case 'update':
            setMessages(prev =>
              prev.map(m => m.id === data.record.id ? data.record : m)
            );
            break;
          case 'delete':
            setMessages(prev =>
              prev.filter(m => m.id !== data.record.id)
            );
            break;
        }
      }
    );

    return () => { unsubscribe.then(fn => fn()); };
  }, [roomId]);

  // Reconnect handler — re-fetch after reconnect
  useEffect(() => {
    isInitialConnect.current = true;

    const unsubscribe = pb.realtime.subscribe('PB_CONNECT', () => {
      if (isInitialConnect.current) {
        isInitialConnect.current = false;
        return; // Skip initial connect
      }
      // Re-fetch on reconnect to catch missed events
      pb.collection('messages')
        .getFullList<Message>({
          filter: `room = "${roomId}"`,
          sort: '-created',
        })
        .then(setMessages);
    });

    return () => { unsubscribe.then(fn => fn()); };
  }, [roomId]);

  return messages;
}
```

### Optimistic Updates

```typescript
async function sendMessage(text: string, roomId: string) {
  const optimisticId = `temp-${Date.now()}`;
  const optimistic: Message = {
    id: optimisticId,
    text,
    room: roomId,
    author: pb.authStore.record!.id,
    expires_at: new Date(Date.now() + 3600000).toISOString(),
    created: new Date().toISOString(),
  };

  // Immediately add to UI
  setMessages(prev => [optimistic, ...prev]);

  try {
    const real = await pb.collection('messages').create<Message>({
      text,
      room: roomId,
      expires_at: optimistic.expires_at,
    });
    // Replace optimistic with real record
    setMessages(prev =>
      prev.map(m => m.id === optimisticId ? real : m)
    );
  } catch (err) {
    // Revert on failure
    setMessages(prev => prev.filter(m => m.id !== optimisticId));
    showError('Message failed to send');
  }
}
```

### Auth Context Provider

```typescript
import { createContext, useContext, useEffect, useState } from 'react';
import pb from '../lib/pocketbase';

interface AuthContextType {
  user: RecordModel | null;
  isValid: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState(pb.authStore.record);
  const [isValid, setIsValid] = useState(pb.authStore.isValid);

  useEffect(() => {
    return pb.authStore.onChange((token, record) => {
      setUser(record);
      setIsValid(pb.authStore.isValid);
    });
  }, []);

  const login = async (email: string, password: string) => {
    await pb.collection('users').authWithPassword(email, password);
  };

  const logout = () => {
    pb.authStore.clear();
  };

  return (
    <AuthContext.Provider value={{ user, isValid, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be inside AuthProvider');
  return ctx;
};
```

---

## 6. Hearth-Specific Patterns

### Campfire (Fading Chat) Integration

```typescript
// Messages have an expires_at field — CSS handles visual decay
// The subscription handler must also handle server-side GC deletes
pb.collection('messages').subscribe<Message>('*', (data) => {
  if (data.action === 'delete') {
    // Server cron GC'd an expired message — remove from UI
    // CSS animation may have already made it invisible
    removeMessage(data.record.id);
  }
});
```

### Presence (Custom Topic)

```typescript
// Subscribe to presence updates (sent via Go SubscriptionsBroker)
pb.realtime.subscribe('presence', (data) => {
  // data: { userId, status, position: { x, y } }
  updatePresence(data);
});

// Heartbeat (send every 30s via a regular API call or custom hook)
setInterval(() => {
  pb.send('/api/hearth/heartbeat', {
    method: 'POST',
    body: { position: { x: 0.5, y: 0.3 } },
  }).catch(() => { /* offline, will reconnect */ });
}, 30000);
```

### Token Refresh Strategy

```typescript
// Refresh auth token periodically (before expiry)
useEffect(() => {
  const interval = setInterval(async () => {
    if (pb.authStore.isValid) {
      try {
        await pb.collection('users').authRefresh();
      } catch {
        // Token expired or invalid — force re-login
        pb.authStore.clear();
      }
    }
  }, 10 * 60 * 1000); // Every 10 minutes

  return () => clearInterval(interval);
}, []);
```

---

## 7. Gotchas & Edge Cases

### SSE vs WebSocket
- PocketBase uses SSE (`EventSource`), not WebSocket. This is transparent to the developer via the SDK, but important for network debugging (look for `GET /api/realtime` in dev tools, not a WS upgrade).
- SSE is unidirectional (server→client). Client→server communication uses regular HTTP POST requests.

### Reconnect Does NOT Replay Events
- After a disconnect/reconnect cycle, the `PB_CONNECT` event fires but **missed events are lost**.
- You MUST re-fetch data after detecting a non-initial `PB_CONNECT` to synchronize state.
- Pattern: Track whether it's the first connect with a `useRef`.

### Auto-Cancellation Pitfalls
- The SDK auto-cancels duplicate pending requests. If you fire two `getList` calls for the same collection quickly, the first is cancelled.
- Use `requestKey: null` to disable, or use unique `requestKey` values.

### Filter Syntax
- PocketBase filters use a custom syntax, NOT SQL: `filter: 'room = "ROOM_ID" && created > "2026-01-01"'`
- String values must be double-quoted inside the filter string.
- Relations use the relation field name directly: `filter: 'author.name = "Alice"'`

### Subscription Cleanup
- Always unsubscribe in `useEffect` return (cleanup function).
- `subscribe()` returns a `Promise<() => void>`, so cleanup requires `unsubscribe.then(fn => fn())`.
- Calling `pb.collection('x').unsubscribe()` without arguments clears ALL subscriptions for that collection — be careful in shared contexts.

### Auth Token in Real-time
- The SSE connection automatically includes the auth token.
- If the user logs out (`authStore.clear()`), existing subscriptions continue until the SSE connection is re-established.
- For immediate disconnection on logout, call `pb.realtime.disconnect()` after `authStore.clear()`.

---

## 8. API Quick Reference

| Operation | Code |
|-----------|------|
| Init client | `new PocketBase(url)` |
| Login | `pb.collection('users').authWithPassword(email, pass)` |
| Logout | `pb.authStore.clear()` |
| Auth change listener | `pb.authStore.onChange(cb)` → returns unsubscribe |
| Create record | `pb.collection('x').create(data)` |
| Read one | `pb.collection('x').getOne(id)` |
| Read list | `pb.collection('x').getList(page, perPage, options)` |
| Read all | `pb.collection('x').getFullList(options)` |
| Update | `pb.collection('x').update(id, data)` |
| Delete | `pb.collection('x').delete(id)` |
| Subscribe (collection) | `pb.collection('x').subscribe('*', cb)` |
| Subscribe (record) | `pb.collection('x').subscribe(id, cb)` |
| Subscribe (custom topic) | `pb.realtime.subscribe(topic, cb)` |
| Unsubscribe | returned function, or `pb.collection('x').unsubscribe()` |
| Connection event | `pb.realtime.subscribe('PB_CONNECT', cb)` |
| Disconnection hook | `pb.realtime.onDisconnect = cb` |
| Custom API call | `pb.send('/api/custom', { method: 'POST', body: data })` |

---

## Notes for Builder

1. **Create a single `pb` instance** in `src/lib/pocketbase.ts` and import it everywhere. Never create per-component instances.
2. **Use `AsyncAuthStore`** if you later need encrypted token storage or React Native support.
3. **Always handle `PB_CONNECT` reconnect** — this is how you avoid stale UI after network interruptions.
4. **SSE works naturally through Caddy** — no special proxy configuration needed (it's standard HTTP).
5. **Filter strings** are PocketBase-specific syntax. Build a helper utility for type-safe filter construction.
6. **For Campfire:** The `expires_at` field drives both CSS decay animation AND server-side cron GC. Subscribe to `delete` events to clean up records that the server GC removes.
7. **For Presence:** Use custom topic subscriptions (`pb.realtime.subscribe('presence', cb)`) paired with a Go-side `SubscriptionsBroker().Send()` hook.
