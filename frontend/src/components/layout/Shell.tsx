import { type ReactNode } from 'react';
import { RoomList } from './RoomList';
import { ConnectionIndicator } from './ConnectionIndicator';

interface Props {
  children: ReactNode;
}

/**
 * App shell — sidebar + main content area.
 * Mobile: single column (sidebar hidden). Tablet+: sidebar visible.
 * K-023: Mobile-first responsive layout.
 */
export function Shell({ children }: Props) {
  return (
    <>
      <ConnectionIndicator />
      <div className="flex h-screen bg-[var(--color-bg-primary)]">
        {/* Room sidebar — hidden on mobile, visible on md+ */}
        <aside className="hidden md:flex md:w-64 lg:w-72 flex-col shrink-0">
          <RoomList />
        </aside>

        {/* Main content */}
        <main className="flex-1 flex flex-col min-w-0">
          {children}
        </main>
      </div>
    </>
  );
}
