import { useDmMessages } from '@/hooks/useDmMessages';
import { DmMessageList } from './DmMessageList';
import { MessageInput } from '@/components/campfire/MessageInput';

interface Props {
  dmId: string;
  otherUserName: string;
}

/**
 * Full DM conversation container — permanent messages, no presence bar.
 * Simpler than CampfireRoom/DenRoom (no membership, 1:1 only).
 */
export function DmRoom({ dmId, otherUserName }: Props) {
  const { messages, isLoading, sendMessage } = useDmMessages(dmId);

  return (
    <div className="flex flex-col h-full animate-slide-in">
      {/* Header — other user's name */}
      <div className="flex items-center px-4 py-3 border-b border-[var(--color-bg-elevated)]">
        <h2 className="font-display text-lg text-[var(--color-text-primary)]">
          {otherUserName}
        </h2>
      </div>

      {isLoading ? (
        <div className="flex-1 flex items-center justify-center">
          <div
            className="w-8 h-8 rounded-full animate-warm-pulse"
            style={{
              background:
                'radial-gradient(circle, var(--color-accent-amber), transparent 70%)',
            }}
          />
        </div>
      ) : (
        <DmMessageList messages={messages} />
      )}

      <MessageInput onSend={sendMessage} />
    </div>
  );
}
