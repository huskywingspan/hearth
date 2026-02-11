import { useParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { Shell } from '@/components/layout/Shell';
import { CampfireRoom } from '@/components/campfire/CampfireRoom';
import { Spinner } from '@/components/ui/Spinner';
import pb from '@/lib/pocketbase';

interface Room {
  id: string;
  name: string;
  slug: string;
}

export default function RoomPage() {
  const { roomId } = useParams<{ roomId: string }>();
  const [room, setRoom] = useState<Room | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!roomId) return;

    pb.collection('rooms')
      .getOne<Room>(roomId)
      .then(setRoom)
      .catch(() => setError('Room not found'));
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
      <CampfireRoom roomId={roomId} roomName={room.name} />
    </Shell>
  );
}
