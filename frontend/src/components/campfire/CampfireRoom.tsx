import { useEffect, useState, useCallback } from 'react';
import pb from '@/lib/pocketbase';
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
 * Ensures membership on mount (ADR-006: join-on-entry model).
 */
export function CampfireRoom({ roomId, roomName }: Props) {
  const [joined, setJoined] = useState(false);

  // Ensure the current user is a member of this room before loading anything.
  // Silently catches duplicate-key errors (already a member).
  const ensureMembership = useCallback(async () => {
    try {
      await pb.collection('room_members').create({
        room: roomId,
        user: pb.authStore.record?.id,
        role: 'member',
      });
    } catch {
      // Unique constraint violation = already a member — safe to ignore
    }
    setJoined(true);
  }, [roomId]);

  useEffect(() => {
    ensureMembership();
  }, [ensureMembership]);

  if (!joined) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div
          className="w-8 h-8 rounded-full animate-warm-pulse"
          style={{
            background:
              'radial-gradient(circle, var(--color-accent-amber), transparent 70%)',
          }}
        />
      </div>
    );
  }

  return <CampfireRoomInner roomId={roomId} roomName={roomName} />;
}

/** Inner component: only mounts after membership is ensured. */
function CampfireRoomInner({ roomId, roomName }: Props) {
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
