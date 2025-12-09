import styles from './YesNo.module.css';

interface YesNoProps {
  value: boolean;
  /** When true, "Yes" is bad (red) and "No" is good (green). Default: false */
  inverted?: boolean;
  className?: string;
}

export function YesNo({ value, inverted = false, className }: YesNoProps) {
  const isPositive = inverted ? !value : value;
  const colorClass = isPositive ? styles.positive : styles.negative;

  return (
    <span className={`${styles.yesNo} ${colorClass} ${className || ''}`}>
      {value ? 'Yes' : 'No'}
    </span>
  );
}
