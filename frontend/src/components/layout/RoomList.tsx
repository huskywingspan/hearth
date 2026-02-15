import { useEffect, useState, useCallback } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import pb from '@/lib/pocketbase';
import { useAuth } from '@/hooks/useAuth';

interface Room {
  id: string;
  name: string;
  slug: string;
  description: string;
  type: 'den' | 'campfire';
}

interface DirectMessage {
  id: string;
  participant_a: string;
  participant_b: string;
  expand?: {
    participant_a?: { display_name: string };
    participant_b?: { display_name: string };
  };
}

/**
 * Room navigation sidebar — ADR-007 three-section layout.
 * Sections: Dens (permanent) | Campfires (ephemeral) | Messages (DMs)
 */
export function RoomList() {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [dms, setDms] = useState<DirectMessage[]>([]);
  const [isCreating, setIsCreating] = useState(false);
  const navigate = useNavigate();
  const { user } = useAuth();

  const userRole = (user as Record<string, unknown>)?.['role'] as string | undefined;
  const canCreateDen = userRole === 'homeowner' || userRole === 'keyholder';

  const fetchRooms = useCallback(async () => {
    try {
      const result = await pb.collection('rooms').getFullList<Room>({
        sort: 'name',
      });
      setRooms(result);
    } catch {
      // Will retry on reconnect
    }
  }, []);

  const fetchDms = useCallback(async () => {
    if (!pb.authStore.record?.id) return;
    try {
      const result = await pb
        .collection('direct_messages')
        .getFullList<DirectMessage>({
          expand: 'participant_a,participant_b',
        });
      setDms(result);
    } catch {
      // DMs collection may not exist yet on fresh installs
    }
  }, []);

  useEffect(() => {
    fetchRooms();
    fetchDms();
  }, [fetchRooms, fetchDms]);

  const dens = rooms.filter((r) => r.type === 'den');
  const campfires = rooms.filter((r) => r.type === 'campfire' || !r.type);

  const handleCreateRoom = async (type: 'den' | 'campfire') => {
    const label = type === 'den' ? 'den' : 'campfire';
    const name = prompt(`Name your ${label}:`);
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
        type,
        default_ttl: type === 'den' ? 0 : 3600,
        max_participants: type === 'den' ? 25 : 10,
        livekit_room_name: `hearth-${slug}-${Date.now()}`,
      });

      await fetchRooms();
      navigate(`/room/${room.id}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : `Failed to create ${label}`);
    } finally {
      setIsCreating(false);
    }
  };

  const handleNewDm = async () => {
    const email = prompt('Enter the user\'s email or display name:');
    if (!email?.trim()) return;

    try {
      // Search for user by display_name or email
      const users = await pb.collection('users').getFullList({
        filter: `display_name ~ "${email.trim()}" || email = "${email.trim()}"`,
      });

      if (users.length === 0) {
        alert('User not found');
        return;
      }

      const other = users[0]!;
      if (other.id === pb.authStore.record?.id) {
        alert("You can't message yourself");
        return;
      }

      // Canonical ordering — sort IDs for unique index
      const myId = pb.authStore.record?.id ?? '';
      const [a, b] = [myId, other.id].sort();

      // Find or create DM
      let dmId: string;
      try {
        const existing = await pb
          .collection('direct_messages')
          .getFirstListItem(`participant_a = "${a}" && participant_b = "${b}"`);
        dmId = existing.id;
      } catch {
        const created = await pb.collection('direct_messages').create({
          participant_a: a,
          participant_b: b,
        });
        dmId = created.id;
      }

      await fetchDms();
      navigate(`/dm/${dmId}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to start conversation');
    }
  };

  /** Get the other participant's display name for a DM */
  const getDmName = (dm: DirectMessage): string => {
    const myId = pb.authStore.record?.id;
    if (dm.participant_a === myId) {
      return dm.expand?.participant_b?.display_name || 'User';
    }
    return dm.expand?.participant_a?.display_name || 'User';
  };

  const roomLink = (room: Room) => (
    <NavLink
      key={room.id}
      to={`/room/${room.id}`}
      className={({ isActive }) =>
        [
          'flex items-center gap-2 px-3 py-1.5 rounded-[var(--radius-md)] text-sm',
          'transition-colors duration-[var(--duration-fast)]',
          isActive
            ? 'bg-[var(--color-bg-elevated)] text-[var(--color-accent-amber)]'
            : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-primary)] hover:text-[var(--color-text-primary)]',
        ].join(' ')
      }
    >
      <span className="text-xs opacity-60">{room.type === 'den' ? '🏠' : '🔥'}</span>
      {room.name}
    </NavLink>
  );

  return (
    <nav className="flex flex-col h-full bg-[var(--color-bg-secondary)]">
      <div className="px-4 py-5">
        <h1 className="font-display text-xl text-[var(--color-accent-amber)]">
          Hearth
        </h1>
      </div>

      <div className="flex-1 overflow-y-auto px-2 space-y-4">
        {/* Dens — permanent rooms */}
        <section>
          <h2 className="px-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
            Dens
          </h2>
          <div className="space-y-0.5">
            {dens.length === 0 ? (
              <p className="px-3 py-1 text-xs text-[var(--color-text-muted)]">
                No dens yet
              </p>
            ) : (
              dens.map(roomLink)
            )}
            {canCreateDen && (
              <button
                onClick={() => handleCreateRoom('den')}
                disabled={isCreating}
                className="w-full px-3 py-1.5 rounded-[var(--radius-md)] text-xs
                  text-[var(--color-text-muted)] hover:text-[var(--color-accent-amber)]
                  hover:bg-[var(--color-bg-primary)] transition-colors duration-[var(--duration-fast)]
                  disabled:opacity-50 text-left"
              >
                + New Den
              </button>
            )}
          </div>
        </section>

        {/* Campfires — ephemeral rooms */}
        <section>
          <h2 className="px-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
            Campfires
          </h2>
          <div className="space-y-0.5">
            {campfires.length === 0 ? (
              <p className="px-3 py-1 text-xs text-[var(--color-text-muted)]">
                No campfires yet
              </p>
            ) : (
              campfires.map(roomLink)
            )}
            <button
              onClick={() => handleCreateRoom('campfire')}
              disabled={isCreating}
              className="w-full px-3 py-1.5 rounded-[var(--radius-md)] text-xs
                text-[var(--color-text-muted)] hover:text-[var(--color-accent-amber)]
                hover:bg-[var(--color-bg-primary)] transition-colors duration-[var(--duration-fast)]
                disabled:opacity-50 text-left"
            >
              + New Campfire
            </button>
          </div>
        </section>

        {/* Messages — DM conversations */}
        <section>
          <h2 className="px-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
            Messages
          </h2>
          <div className="space-y-0.5">
            {dms.length === 0 ? (
              <p className="px-3 py-1 text-xs text-[var(--color-text-muted)]">
                No conversations yet
              </p>
            ) : (
              dms.map((dm) => (
                <NavLink
                  key={dm.id}
                  to={`/dm/${dm.id}`}
                  className={({ isActive }) =>
                    [
                      'block px-3 py-1.5 rounded-[var(--radius-md)] text-sm',
                      'transition-colors duration-[var(--duration-fast)]',
                      isActive
                        ? 'bg-[var(--color-bg-elevated)] text-[var(--color-accent-amber)]'
                        : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-primary)] hover:text-[var(--color-text-primary)]',
                    ].join(' ')
                  }
                >
                  {getDmName(dm)}
                </NavLink>
              ))
            )}
            <button
              onClick={handleNewDm}
              className="w-full px-3 py-1.5 rounded-[var(--radius-md)] text-xs
                text-[var(--color-text-muted)] hover:text-[var(--color-accent-amber)]
                hover:bg-[var(--color-bg-primary)] transition-colors duration-[var(--duration-fast)]
                text-left"
            >
              + New Message
            </button>
          </div>
        </section>
      </div>

      {/* User section */}
      <div className="p-4 border-t border-[var(--color-bg-elevated)]">
        <div className="flex items-center justify-between">
          <span className="text-sm text-[var(--color-text-secondary)] truncate">
            {userRole === 'homeowner' && '🏠 '}
            {userRole === 'keyholder' && '🔑 '}
            {(user as Record<string, unknown>)?.['display_name'] as string || 'User'}
          </span>
          <button
            onClick={() => pb.authStore.clear()}
            className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-alert-clay)] transition-colors duration-[var(--duration-fast)]"
          >
            Sign out
          </button>
        </div>
      </div>
    </nav>
  );
}
