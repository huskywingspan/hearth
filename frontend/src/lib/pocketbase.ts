import PocketBase from 'pocketbase';

// Single global instance — NEVER create per-component.
// In dev, Vite proxy handles /api/* → localhost:8090.
// In production, Caddy proxies everything under the same origin.
const pb = new PocketBase(import.meta.env.VITE_PB_URL || '/');

export default pb;
