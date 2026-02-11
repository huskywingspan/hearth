interface Props {
  typingUsers: string[];
}

/**
 * "Mumbling" typing indicator (K-015).
 * Blurred waveform / abstract scribbles indicating rhythm — not "User is typing..."
 * Per master plan: warmth over mechanical status text.
 */
export function MumblingIndicator({ typingUsers }: Props) {
  if (typingUsers.length === 0) return null;

  return (
    <div className="flex items-center gap-2 px-4 py-2 text-[var(--color-text-muted)] text-sm">
      <div className="flex gap-0.5 items-end">
        {/* Three blurred "scribble" bars that undulate */}
        <span className="mumble-bar" style={{ animationDelay: '0ms' }} />
        <span className="mumble-bar" style={{ animationDelay: '150ms' }} />
        <span className="mumble-bar" style={{ animationDelay: '300ms' }} />
      </div>
      <span className="blur-[1px] select-none">
        {typingUsers.length === 1
          ? `${typingUsers[0]} is thinking...`
          : `${typingUsers.length} people murmuring...`}
      </span>
    </div>
  );
}
