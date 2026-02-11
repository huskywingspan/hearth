import type { PresenceEntry } from '@/hooks/usePresence';
import { Avatar } from '@/components/ui/Avatar';

interface Props {
  users: PresenceEntry[];
  roomName?: string;
}

/**
 * Who's-here bar at the top of a Campfire room.
 * Stacked avatars, count, overflow indicator.
 */
export function PresenceBar({ users, roomName }: Props) {
  return (
    <div className="flex items-center gap-3 px-4 py-3 border-b border-[var(--color-bg-elevated)]">
      {roomName && (
        <h2 className="font-display text-lg text-[var(--color-text-primary)] mr-auto">
          {roomName}
        </h2>
      )}
      <span className="text-[var(--color-text-muted)] text-sm">
        {users.length} {users.length === 1 ? 'person' : 'people'} here
      </span>
      <div className="flex -space-x-2">
        {users.slice(0, 8).map((u) => (
          <Avatar
            key={u.user_id}
            name={u.display_name}
            className="ring-2 ring-[var(--color-bg-primary)]"
            size="sm"
          />
        ))}
        {users.length > 8 && (
          <span className="text-[var(--color-text-muted)] text-xs ml-2 self-center">
            +{users.length - 8}
          </span>
        )}
      </div>
    </div>
  );
}
