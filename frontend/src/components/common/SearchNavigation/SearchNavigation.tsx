import { Icon } from '../Icon';
import styles from './SearchNavigation.module.css';

interface SearchNavigationProps {
  currentIndex: number;
  totalMatches: number;
  onPrevious: () => void;
  onNext: () => void;
}

export function SearchNavigation({
  currentIndex,
  totalMatches,
  onPrevious,
  onNext,
}: SearchNavigationProps) {
  if (totalMatches === 0) return null;

  return (
    <div className={styles.navButtons}>
      <span className={styles.matchInfo}>
        {currentIndex + 1} / {totalMatches}
      </span>
      <button className={styles.navButton} onClick={onPrevious} aria-label="Previous match">
        <Icon name="chevron-up" size={16} />
      </button>
      <button className={styles.navButton} onClick={onNext} aria-label="Next match">
        <Icon name="chevron-down" size={16} />
      </button>
    </div>
  );
}
