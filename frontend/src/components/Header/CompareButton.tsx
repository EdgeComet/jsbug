import styles from './Header.module.css';

interface CompareButtonProps {
  onClick: () => void;
  disabled?: boolean;
}

export function CompareButton({ onClick, disabled }: CompareButtonProps) {
  return (
    <button
      className={styles.btnCompare}
      onClick={onClick}
      disabled={disabled}
    >
      <span className={styles.compareText}>ANALYZE</span>
    </button>
  );
}
