import { useParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { Shell } from '@/components/layout/Shell';
import { CampfireRoom } from '@/components/campfire/CampfireRoom';
import { DenRoom } from '@/components/den/DenRoom';
import { Spinner } from '@/components/ui/Spinner';
import pb from '@/lib/pocketbase';

interface Room {
  id: string;
  name: string;
  slug: string;
  type: 'den' | 'campfire';
}

export default function RoomPage() {
  const { roomId } = useParams<{ roomId: string }>();
  const [room, setRoom] = useState<Room | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!roomId) return;

    pb.collection('rooms')
      .getOne<Room>(roomId, { requestKey: null })
      .then(setRoom)
      .catch((err) => {
        console.error('Room fetch failed:', err);
        setError(err?.status === 404 ? 'Room not found' : `Failed to load room: ${err?.message || err}`);
      });
  }, [roomId]);

  if (error) {
    return (
      <Shell>
        <div className="flex items-center justify-center h-full text-[var(--color-alert-clay)]">
          {error}
        </div>
      </Shell>
    );
  }

  if (!room || !roomId) {
    return (
      <Shell>
        <div className="flex items-center justify-center h-full">
          <Spinner />
        </div>
      </Shell>
    );
  }

  return (
    <Shell>
      {room.type === 'den' ? (
        <DenRoom key={roomId} roomId={roomId} roomName={room.name} />
      ) : (
        <CampfireRoom key={roomId} roomId={roomId} roomName={room.name} />
      )}
    </Shell>
  );
}
