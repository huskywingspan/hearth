import { useEffect, useState } from 'react';
import { NavLink } from 'react-router-dom';
import pb from '@/lib/pocketbase';

interface Room {
  id: string;
  name: string;
  slug: string;
  description: string;
}

/**
 * Room navigation sidebar.
 * Fetches rooms the user is a member of.
 */
export function RoomList() {
  const [rooms, setRooms] = useState<Room[]>([]);

  useEffect(() => {
    async function fetchRooms() {
      try {
        // Fetch rooms the current user is a member of
        const userId = pb.authStore.record?.id;
        if (!userId) return;

        const memberships = await pb
          .collection('room_members')
          .getFullList({
            filter: `user = "${userId}"`,
            expand: 'room',
          });

        const roomList = memberships
          .map((m) => {
            const expanded = m.expand as { room?: Room } | undefined;
            return expanded?.room;
          })
          .filter((r): r is Room => !!r);

        setRooms(roomList);
      } catch {
        // Will retry on reconnect
      }
    }

    fetchRooms();
  }, []);

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
