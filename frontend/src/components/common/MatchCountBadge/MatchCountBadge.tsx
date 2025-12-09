import styles from './MatchCountBadge.module.css';

interface MatchCountBadgeProps {
  count: number;
}

export function MatchCountBadge({ count }: MatchCountBadgeProps) {
  return (
    <span className={styles.matchCount}>
      {count} {count === 1 ? 'match' : 'matches'}
    </span>
  );
}
