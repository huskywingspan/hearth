import { Shell } from '@/components/layout/Shell';
import { Card } from '@/components/ui/Card';
import { useAuth } from '@/hooks/useAuth';

export default function HomePage() {
  const { user } = useAuth();

  return (
    <Shell>
      <div className="flex items-center justify-center h-full p-8">
        <Card className="max-w-md text-center animate-float-in">
          <h1 className="font-display text-2xl mb-4 text-[var(--color-text-primary)]">
            Welcome back{user ? `, ${(user as Record<string, unknown>).display_name ?? 'friend'}` : ''}
          </h1>
          <p className="text-[var(--color-text-secondary)] mb-6">
            Pick a room from the sidebar, or start a new campfire.
          </p>
          <div className="text-[var(--color-text-muted)] text-sm">
            The fire is warm. Pull up a chair.
          </div>
        </Card>
      </div>
    </Shell>
  );
}
