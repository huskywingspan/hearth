/**
 * Warm pulsing spinner — not a cold spinning wheel.
 * Ember glow that breathes.
 */
export function Spinner({ className = '' }: { className?: string }) {
  return (
    <div className={`flex items-center justify-center ${className}`}>
      <div
        className="w-8 h-8 rounded-full animate-warm-pulse"
        style={{
          background: `radial-gradient(circle, var(--color-accent-amber), transparent 70%)`,
        }}
      />
    </div>
  );
}
