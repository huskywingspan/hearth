import { useEffect, useRef } from 'react';
import { MessageBubble } from './MessageBubble';
import type { Message } from '@/hooks/useMessages';

interface Props {
  messages: Message[];
  onMessageGone: (id: string) => void;
}

/**
 * Scrollable message list with auto-scroll anchoring.
 * Auto-scrolls to bottom on new messages (if user hasn't scrolled up).
 */
export function MessageList({ messages, onMessageGone }: Props) {
  const listRef = useRef<HTMLDivElement>(null);
  const shouldAutoScroll = useRef(true);

  // Auto-scroll to bottom on new messages
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
      {messages.length === 0 ? (
        <div className="flex items-center justify-center h-full text-[var(--color-text-muted)] text-sm">
          The campfire is quiet. Say something...
        </div>
      ) : (
        messages.map((msg) => (
          <MessageBubble
            key={msg.id}
            message={msg}
            onGone={onMessageGone}
          />
        ))
      )}
    </div>
  );
}
