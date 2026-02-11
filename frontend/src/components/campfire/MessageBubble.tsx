import { useCallback, type CSSProperties } from 'react';
import { useTimeSync } from '@/hooks/useTimeSync';
import type { Message } from '@/hooks/useMessages';

interface Props {
  message: Message;
  onGone: (id: string) => void;
}

/**
 * Individual fading message in the Campfire.
 *
 * Architecture (from R-008):
 * - ONE CSS animation per message via custom properties
 * - Negative animation-delay starts mid-fade for old messages on page load
 * - animationend fires → remove from React state → DOM cleanup
 * - content-visibility: auto handled by .campfire-message class
 * - NO will-change (browser auto-promotes animated elements)
 */
export function MessageBubble({ message, onGone }: Props) {
  const { getServerNow } = useTimeSync();

  const isOptimistic = message.id.startsWith('temp-');
  const serverNow = getServerNow();

  // Calculate animation timing
  const createdMs = new Date(message.created).getTime();
  const expiresMs = new Date(message.expires_at).getTime();
  const fadeDuration = (expiresMs - createdMs) / 1000; // Total TTL in seconds
  const ageSeconds = (serverNow - createdMs) / 1000;
  const ageOffset = -ageSeconds; // Negative = start mid-fade

  // Already expired — don't render
  if (ageSeconds >= fadeDuration) return null;

  const handleAnimationEnd = useCallback(
    (e: React.AnimationEvent) => {
      // Only handle the fade animation, not the float-in entrance
      if (e.animationName === 'campfire-fade') {
        onGone(message.id);
      }
    },
    [message.id, onGone]
  );

  const style: CSSProperties = {
    '--fade-duration': `${fadeDuration}s`,
    '--age-offset': `${ageOffset}s`,
  } as CSSProperties;

  const authorName =
    message.author_name || message.expand?.author?.display_name || 'Wanderer';

  return (
    <div
      className="campfire-message animate-float-in"
      style={style}
      data-optimistic={isOptimistic || undefined}
      onAnimationEnd={handleAnimationEnd}
    >
      <div className="flex items-start gap-3 p-3 rounded-[var(--radius-lg)]">
        <span className="font-medium text-sm text-[var(--color-accent-amber)] shrink-0">
          {authorName}
        </span>
        <span className="text-[var(--color-text-primary)] text-base break-words min-w-0">
          {message.body}
        </span>
      </div>
    </div>
  );
}
