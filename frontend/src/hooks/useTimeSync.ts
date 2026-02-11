import { useEffect, useRef, useCallback } from 'react';

/**
 * Calculates offset between client clock and server clock
 * using the Date header from PocketBase API responses.
 *
 * serverTime ≈ clientTime + offset
 *
 * Accuracy: ~500ms (sufficient for visual fading).
 */
export function useTimeSync() {
  const offsetMs = useRef(0);

  useEffect(() => {
    async function sync() {
      const before = Date.now();
      const response = await fetch(
        (import.meta.env.VITE_PB_URL || '') + '/api/health'
      );
      const after = Date.now();
      const rtt = after - before;

      const serverDateStr = response.headers.get('Date');
      if (!serverDateStr) return;

      const serverTime = new Date(serverDateStr).getTime();
      // Estimate: server header sent at the midpoint of RTT
      const estimatedClientTimeAtServer = before + rtt / 2;
      offsetMs.current = serverTime - estimatedClientTimeAtServer;
    }

    sync();
    // Re-sync every 5 minutes
    const interval = setInterval(sync, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  /** Get the current server time (client clock + measured offset). */
  const getServerNow = useCallback(() => {
    return Date.now() + offsetMs.current;
  }, []);

  return { getServerNow, offsetMs: offsetMs.current };
}
