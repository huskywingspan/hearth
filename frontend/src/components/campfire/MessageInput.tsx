import { useState, useCallback, useRef, type FormEvent } from 'react';
import { Button } from '@/components/ui/Button';
import { TYPING_DEBOUNCE_MS } from '@/lib/constants';

interface Props {
  onSend: (text: string) => void;
  onTypingStart?: () => void;
  onTypingStop?: () => void;
  disabled?: boolean;
}

/**
 * Message compose input — warm styling, pillow send button.
 * Debounced typing indicator emission.
 */
export function MessageInput({
  onSend,
  onTypingStart,
  onTypingStop,
  disabled = false,
}: Props) {
  const [text, setText] = useState('');
  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      const trimmed = text.trim();
      if (!trimmed) return;
      onSend(trimmed);
      setText('');
      // Clear typing indicator
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      onTypingStop?.();
    },
    [text, onSend, onTypingStop]
  );

  const handleChange = useCallback(
    (value: string) => {
      setText(value);

      // Typing indicator debounce
      if (value.trim()) {
        onTypingStart?.();
        if (typingTimeoutRef.current) {
          clearTimeout(typingTimeoutRef.current);
        }
        typingTimeoutRef.current = setTimeout(() => {
          onTypingStop?.();
        }, TYPING_DEBOUNCE_MS);
      } else {
        onTypingStop?.();
      }
    },
    [onTypingStart, onTypingStop]
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      // Enter sends, Shift+Enter for newline (not yet multiline)
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        const trimmed = text.trim();
        if (!trimmed) return;
        onSend(trimmed);
        setText('');
        if (typingTimeoutRef.current) {
          clearTimeout(typingTimeoutRef.current);
        }
        onTypingStop?.();
      }
    },
    [text, onSend, onTypingStop]
  );

  return (
    <form
      onSubmit={handleSubmit}
      className="flex items-center gap-3 p-4 border-t border-[var(--color-bg-elevated)]"
    >
      <input
        type="text"
        value={text}
        onChange={(e) => handleChange(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Say something..."
        disabled={disabled}
        className={[
          'flex-1 bg-[var(--color-bg-input)]',
          'text-[var(--color-text-primary)]',
          'placeholder:text-[var(--color-text-muted)]',
          'rounded-[var(--radius-pill)]',
          'border-none px-5 py-2.5',
          'transition-shadow duration-[var(--duration-normal)]',
          'focus:outline-none focus:shadow-[var(--shadow-glow)]',
        ].join(' ')}
      />
      <Button
        type="submit"
        variant="primary"
        size="md"
        disabled={disabled || !text.trim()}
      >
        Send
      </Button>
    </form>
  );
}
