import { type ButtonHTMLAttributes, type ReactNode } from 'react';

type Variant = 'primary' | 'secondary' | 'ghost';
type Size = 'sm' | 'md' | 'lg';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  children: ReactNode;
}

/**
 * "Pillow" button — rounded-pill, squash on press, warm amber primary.
 * Per master plan: Disney "squash & stretch" principle.
 */
export function Button({
  variant = 'primary',
  size = 'md',
  className = '',
  children,
  ...props
}: ButtonProps) {
  const base = [
    'inline-flex items-center justify-center',
    'font-medium',
    'rounded-[9999px]',
    'transition-all duration-[var(--duration-normal)]',
    'active:scale-95',
    'focus-visible:outline-2 focus-visible:outline-offset-2',
    'focus-visible:outline-[var(--color-accent-amber)]',
    'disabled:opacity-50 disabled:cursor-not-allowed',
  ].join(' ');

  const variants: Record<Variant, string> = {
    primary: [
      'bg-[var(--color-accent-amber)] text-[var(--color-bg-primary)]',
      'hover:bg-[var(--color-accent-amber-hover)]',
      'shadow-[var(--shadow-md)] hover:shadow-[var(--shadow-lg)]',
    ].join(' '),
    secondary: [
      'bg-[var(--color-bg-elevated)] text-[var(--color-text-primary)]',
      'border border-[var(--color-bg-elevated)]',
      'hover:bg-[var(--color-bg-secondary)]',
    ].join(' '),
    ghost: [
      'text-[var(--color-text-secondary)]',
      'hover:text-[var(--color-text-primary)]',
      'hover:bg-[var(--color-bg-secondary)]',
    ].join(' '),
  };

  const sizes: Record<Size, string> = {
    sm: 'text-sm px-3 py-1.5',
    md: 'text-base px-5 py-2.5',
    lg: 'text-lg px-7 py-3',
  };

  return (
    <button
      className={`${base} ${variants[variant]} ${sizes[size]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
