import { useEffect, useRef } from 'react';
import type { Message } from '@/hooks/useMessages';

interface Props {
  messages: Message[];
}

/**
 * Message list for Dens — permanent messages, NO fade animation.
 * Same auto-scroll behavior as campfire MessageList.
 */
export function DenMessageList({ messages }: Props) {
  const listRef = useRef<HTMLDivElement>(null);
  const shouldAutoScroll = useRef(true);

  useEffect(() => {
    const el = listRef.current;
    if (!el || !shouldAutoScroll.current) return;
    el.scrollTop = el.scrollHeight;
  }, [messages.length]);

  const handleScroll = () => {
    const el = listRef.current;
    if (!el) return;
    const threshold = 100;
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
          Pull up a chair. Say something...
        </div>
      ) : (
        messages.map((msg) => {
          const authorName =
            msg.author_name || msg.expand?.author?.display_name || 'Wanderer';

          return (
            <div
              key={msg.id}
              className="animate-float-in"
              data-optimistic={msg.id.startsWith('temp-') || undefined}
            >
              <div className="flex items-start gap-3 p-3 rounded-[var(--radius-lg)]">
                <span className="font-medium text-sm text-[var(--color-accent-amber)] shrink-0">
                  {authorName}
                </span>
                <span className="text-[var(--color-text-primary)] text-base break-words min-w-0">
                  {msg.body}
                </span>
              </div>
            </div>
          );
        })
      )}
    </div>
  );
}
