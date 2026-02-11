interface AvatarProps {
  name: string;
  src?: string;
  size?: 'sm' | 'md' | 'lg';
  active?: boolean;
  className?: string;
}

const sizeClasses = {
  sm: 'w-7 h-7 text-xs',
  md: 'w-10 h-10 text-sm',
  lg: 'w-14 h-14 text-base',
} as const;

/**
 * Rounded avatar with initials fallback.
 * Ember glow ring when user is active/speaking.
 */
export function Avatar({
  name,
  src,
  size = 'md',
  active = false,
  className = '',
}: AvatarProps) {
  const initials = name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  // Deterministic color from name (warm palette)
  const hue = name.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0) % 40 + 15; // 15-55 (warm)

  return (
    <div
      className={[
        'rounded-full flex items-center justify-center font-medium select-none',
        'transition-shadow duration-[var(--duration-normal)]',
        sizeClasses[size],
        active ? 'ring-2 ring-[var(--color-accent-amber)] shadow-[var(--shadow-glow)]' : '',
        className,
      ].join(' ')}
      style={
        src
          ? undefined
          : { backgroundColor: `hsl(${hue}, 40%, 35%)`, color: `hsl(${hue}, 30%, 80%)` }
      }
      title={name}
    >
      {src ? (
        <img
          src={src}
          alt={name}
          className="w-full h-full rounded-full object-cover"
        />
      ) : (
        initials
      )}
    </div>
  );
}
