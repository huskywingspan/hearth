import { useEffect, useState, useCallback } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import pb from '@/lib/pocketbase';

interface Room {
  id: string;
  name: string;
  slug: string;
  description: string;
}

/**
 * Room navigation sidebar.
 * Fetches all rooms (ADR-006: open-lobby model).
 */
export function RoomList() {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [isCreating, setIsCreating] = useState(false);
  const navigate = useNavigate();

  const fetchRooms = useCallback(async () => {
    try {
      // ADR-006: Any authenticated user can list all rooms
      const result = await pb.collection('rooms').getFullList<Room>({
        sort: 'name',
      });
      setRooms(result);
    } catch {
      // Will retry on reconnect
    }
  }, []);

  useEffect(() => {
    fetchRooms();
  }, [fetchRooms]);

  const handleCreateRoom = async () => {
    const name = prompt('Name your campfire:');
    if (!name?.trim()) return;

    setIsCreating(true);
    try {
      const slug = name
        .trim()
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-|-$/g, '');

      const room = await pb.collection('rooms').create({
        name: name.trim(),
        slug,
        owner: pb.authStore.record?.id,
        default_ttl: 3600,
        max_participants: 10,
        livekit_room_name: `hearth-${slug}-${Date.now()}`,
      });

      // Backend hook auto-creates owner membership (S3-010: no duplicate create)
      await fetchRooms();
      navigate(`/room/${room.id}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create room');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <nav className="flex flex-col h-full bg-[var(--color-bg-secondary)]">
      <div className="px-4 py-5">
        <h1 className="font-display text-xl text-[var(--color-accent-amber)]">
          Hearth
        </h1>
      </div>

      <div className="flex-1 overflow-y-auto px-2 space-y-1">
        {rooms.length === 0 ? (
          <p className="px-3 py-2 text-sm text-[var(--color-text-muted)]">
            No rooms yet
          </p>
        ) : (
          rooms.map((room) => (
            <NavLink
              key={room.id}
              to={`/room/${room.id}`}
              className={({ isActive }) =>
                [
                  'block px-3 py-2 rounded-[var(--radius-md)] text-sm',
                  'transition-colors duration-[var(--duration-fast)]',
                  isActive
                    ? 'bg-[var(--color-bg-elevated)] text-[var(--color-accent-amber)]'
                    : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-primary)] hover:text-[var(--color-text-primary)]',
                ].join(' ')
              }
            >
              {room.name}
            </NavLink>
          ))
        )}

        <button
          onClick={handleCreateRoom}
          disabled={isCreating}
          className="w-full mt-2 px-3 py-2 rounded-[var(--radius-md)] text-sm
            text-[var(--color-text-muted)] hover:text-[var(--color-accent-amber)]
            hover:bg-[var(--color-bg-primary)] transition-colors duration-[var(--duration-fast)]
            border border-dashed border-[var(--color-bg-elevated)] hover:border-[var(--color-accent-amber)]
            disabled:opacity-50"
        >
          {isCreating ? '...' : '+ New campfire'}
        </button>
      </div>

      <div className="p-4 border-t border-[var(--color-bg-elevated)]">
        <button
          onClick={() => pb.authStore.clear()}
          className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-alert-clay)] transition-colors duration-[var(--duration-fast)]"
        >
          Sign out
        </button>
      </div>
    </nav>
  );
}
