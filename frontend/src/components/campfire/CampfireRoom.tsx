import { useMessages } from '@/hooks/useMessages';
import { usePresence } from '@/hooks/usePresence';
import { MessageList } from './MessageList';
import { MessageInput } from './MessageInput';
import { MumblingIndicator } from './MumblingIndicator';
import { PresenceBar } from './PresenceBar';

interface Props {
  roomId: string;
  roomName?: string;
}

/**
 * Full Campfire room container — presence bar, message list, input.
 * Composes all campfire sub-components.
 */
export function CampfireRoom({ roomId, roomName }: Props) {
  const { messages, isLoading, sendMessage, removeMessage } =
    useMessages(roomId);
  const { presentUsers } = usePresence(roomId);

  // TODO: Typing indicator state (future: custom topic subscription)
  const typingUsers: string[] = [];

  return (
    <div className="flex flex-col h-full animate-slide-in">
      <PresenceBar users={presentUsers} roomName={roomName} />

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
        <MessageList messages={messages} onMessageGone={removeMessage} />
      )}

      <MumblingIndicator typingUsers={typingUsers} />
      <MessageInput onSend={sendMessage} />
    </div>
  );
}
