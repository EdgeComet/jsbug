import styles from './DiffBadge.module.css';

interface DiffBadgeProps {
  value: number;
  unit: string;
  className?: string;
}

export function DiffBadge({ value, unit, className }: DiffBadgeProps) {
  const isAdded = value >= 0;
  const prefix = isAdded ? '+' : '';
  const typeClass = isAdded ? styles.added : styles.removed;

  return (
    <span className={`${styles.badge} ${typeClass} ${className || ''}`}>
      {prefix}{value}{unit}
    </span>
  );
}
