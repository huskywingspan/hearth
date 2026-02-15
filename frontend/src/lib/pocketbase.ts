import PocketBase from 'pocketbase';

/**
 * Resolve the PocketBase server URL.
 * Priority: 1) localStorage (user-selected), 2) env var, 3) same-origin
 */
function getServerUrl(): string {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('hearth_server_url');
    if (saved) return saved;
  }
  return import.meta.env.VITE_PB_URL || '/';
}

// Single global instance — NEVER create per-component.
const pb = new PocketBase(getServerUrl());

/** Update the PocketBase server URL, clear auth, and reload. */
export function setServerUrl(url: string) {
  let normalized = url.trim().replace(/\/+$/, '');
  if (!/^https?:\/\//i.test(normalized)) {
    normalized = 'https://' + normalized;
  }
  localStorage.setItem('hearth_server_url', normalized);
  pb.authStore.clear();
  window.location.reload();
}

/** Clear the saved server URL and return to default. */
export function clearServerUrl() {
  localStorage.removeItem('hearth_server_url');
  pb.authStore.clear();
  window.location.reload();
}

/** Get the current server URL for display. */
export function getCurrentServerUrl(): string {
  return localStorage.getItem('hearth_server_url') || '(local)';
}

/** Check if a custom server URL is configured. */
export function hasCustomServerUrl(): boolean {
  return !!localStorage.getItem('hearth_server_url');
}

export default pb;
