import { useEffect, useState, useCallback } from 'react';
import pb from '@/lib/pocketbase';
import { useDenMessages } from '@/hooks/useDenMessages';
import { usePresence } from '@/hooks/usePresence';
import { DenMessageList } from './DenMessageList';
import { MessageInput } from '@/components/campfire/MessageInput';
import { MumblingIndicator } from '@/components/campfire/MumblingIndicator';
import { PresenceBar } from '@/components/campfire/PresenceBar';

interface Props {
  roomId: string;
  roomName?: string;
}

/**
 * Full Den room container — permanent messages, no fade.
 * Ensures membership on mount (same as CampfireRoom).
 */
export function DenRoom({ roomId, roomName }: Props) {
  const [joined, setJoined] = useState(false);

  const ensureMembership = useCallback(async () => {
    try {
      await pb.collection('room_members').create({
        room: roomId,
        user: pb.authStore.record?.id,
        role: 'member',
      });
    } catch {
      // Unique constraint = already a member
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

  return <DenRoomInner roomId={roomId} roomName={roomName} />;
}

/** Inner component: only mounts after membership is ensured. */
function DenRoomInner({ roomId, roomName }: Props) {
  const { messages, isLoading, sendMessage } = useDenMessages(roomId);
  const { presentUsers } = usePresence(roomId);
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
        <DenMessageList messages={messages} />
      )}

      <MumblingIndicator typingUsers={typingUsers} />
      <MessageInput onSend={sendMessage} />
    </div>
  );
}
