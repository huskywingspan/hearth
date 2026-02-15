interface Props {
  role?: string;
  className?: string;
}

/**
 * Role badge — visual indicator for Homeowner/Keyholder.
 * Members get no badge (they're the default).
 */
export function RoleBadge({ role, className = '' }: Props) {
  if (role === 'homeowner') {
    return (
      <span title="Homeowner" className={className} aria-label="Homeowner">
        🏠
      </span>
    );
  }
  if (role === 'keyholder') {
    return (
      <span title="Keyholder" className={className} aria-label="Keyholder">
        🔑
      </span>
    );
  }
  return null;
}
