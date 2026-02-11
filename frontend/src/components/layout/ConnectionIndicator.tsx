import { useConnectionState } from '@/hooks/useReconnect';

/**
 * Top banner when SSE connection is lost (K-016).
 * Burnt clay color for visibility. Floats in.
 */
export function ConnectionIndicator() {
  const { isConnected, isReconnecting } = useConnectionState();

  if (isConnected) return null;

  return (
    <div className="fixed top-0 inset-x-0 bg-[var(--color-alert-clay)] text-[var(--color-text-primary)] text-center py-1.5 text-sm animate-float-in z-50">
      {isReconnecting ? 'Reconnecting...' : 'Offline'}
    </div>
  );
}
