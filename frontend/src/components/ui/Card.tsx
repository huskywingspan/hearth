import { type HTMLAttributes, type ReactNode } from 'react';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  padded?: boolean;
}

/**
 * Elevated card — warm charcoal, rounded-xl, soft shadow.
 * Uses generous padding per "Subtle Warmth" design system.
 */
export function Card({
  children,
  padded = true,
  className = '',
  ...props
}: CardProps) {
  return (
    <div
      className={[
        'bg-[var(--color-bg-elevated)]',
        'rounded-[var(--radius-xl)]',
        'shadow-[var(--shadow-md)]',
        padded ? 'p-[var(--space-lg)]' : '',
        className,
      ].join(' ')}
      {...props}
    >
      {children}
    </div>
  );
}
