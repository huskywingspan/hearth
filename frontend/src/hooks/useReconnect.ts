import { useEffect, useRef, useState, useCallback } from 'react';
import pb from '@/lib/pocketbase';

/**
 * Handles SSE reconnect state synchronization.
 * On every PB_CONNECT after the first, re-fetches data to catch missed events.
 *
 * From R-004:
 * - PB_CONNECT fires on EVERY connection (including the first)
 * - Reconnect does NOT replay missed events
 * - SDK auto-reconnects with backoff: [200,300,500,1000,1200,1500,2000]ms
 */
export function useReconnect(onResync: () => void) {
  const isFirstConnect = useRef(true);

  useEffect(() => {
    const unsubscribe = pb.realtime.subscribe('PB_CONNECT', async () => {
      if (isFirstConnect.current) {
        isFirstConnect.current = false;
        return; // First connect — no resync needed
      }

      // Non-initial connect: we missed events during disconnection.
      // Validate auth is still good
      try {
        await pb.collection('users').authRefresh();
      } catch {
        pb.authStore.clear();
        return;
      }

      // Re-fetch current state
      onResync();
    });

    return () => {
      unsubscribe.then((fn) => fn());
    };
  }, [onResync]);
}

/**
 * Tracks connection state (connected / reconnecting / offline).
 * Uses PB_CONNECT and onDisconnect callbacks from R-004.
 */
export function useConnectionState() {
  const [isConnected, setIsConnected] = useState(true);
  const [isReconnecting, setIsReconnecting] = useState(false);

  useEffect(() => {
    pb.realtime.onDisconnect = () => {
      setIsConnected(false);
      setIsReconnecting(true);
    };

    const unsubscribe = pb.realtime.subscribe('PB_CONNECT', () => {
      setIsConnected(true);
      setIsReconnecting(false);
    });

    return () => {
      unsubscribe.then((fn) => fn());
    };
  }, []);

  const forceResync = useCallback(() => {
    setIsReconnecting(true);
  }, []);

  return { isConnected, isReconnecting, forceResync };
}
