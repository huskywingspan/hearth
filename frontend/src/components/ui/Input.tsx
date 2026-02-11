import { type InputHTMLAttributes, forwardRef } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

/**
 * Warm input — deep background, rounded, ember glow on focus.
 * No border by default; focus uses shadow-glow.
 */
export const Input = forwardRef<HTMLInputElement, InputProps>(
  function Input({ label, error, className = '', id, ...props }, ref) {
    const inputId = id ?? label?.toLowerCase().replace(/\s+/g, '-');

    return (
      <div className="flex flex-col gap-1">
        {label && (
          <label
            htmlFor={inputId}
            className="text-sm font-medium text-[var(--color-text-secondary)]"
          >
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          className={[
            'bg-[var(--color-bg-input)]',
            'text-[var(--color-text-primary)]',
            'placeholder:text-[var(--color-text-muted)]',
            'rounded-[var(--radius-lg)]',
            'border-none',
            'px-4 py-2.5',
            'transition-shadow duration-[var(--duration-normal)]',
            'focus:outline-none focus:shadow-[var(--shadow-glow)]',
            error ? 'ring-2 ring-[var(--color-alert-clay)]' : '',
            className,
          ].join(' ')}
          {...props}
        />
        {error && (
          <span className="text-xs text-[var(--color-alert-clay)]">{error}</span>
        )}
      </div>
    );
  }
);
