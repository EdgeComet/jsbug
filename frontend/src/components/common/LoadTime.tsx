import styles from './LoadTime.module.css';

interface LoadTimeProps {
  seconds: number | null;
  compareSeconds?: number | null;
  className?: string;
}

function getSpeedClass(seconds: number): string {
  if (seconds < 0.7) return styles.excellent;
  if (seconds < 1) return styles.good;
  if (seconds < 2) return styles.moderate;
  if (seconds < 4) return styles.slow;
  return styles.critical;
}

function formatTime(seconds: number): string {
  if (seconds < 1) return `${Math.round(seconds * 1000)}ms`;
  return `${seconds.toFixed(2).replace(/\.?0+$/, '')}s`;
}

export function LoadTime({ seconds, compareSeconds, className }: LoadTimeProps) {
  if (seconds === null) {
    return <span className={`${styles.loadTime} ${className || ''}`}>-</span>;
  }
  const speedClass = seconds === 0 ? styles.zero : getSpeedClass(seconds);
  const isCritical = seconds >= 4;
  const formatted = formatTime(seconds);

  // Show diff badge if this panel's load time is larger than the other
  const showDiff = compareSeconds !== null && compareSeconds !== undefined && seconds > compareSeconds;
  const diffSeconds = showDiff ? seconds - compareSeconds : 0;

  return (
    <span className={`${styles.loadTime} ${speedClass} ${className || ''}`}>
      <span>{formatted}</span>
      {isCritical && (
        <span className={styles.dangerIcon} title="Critical load time">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z"
              clipRule="evenodd"
            />
          </svg>
        </span>
      )}
      {showDiff && (
        <span className={styles.diffBadge}>+{formatTime(diffSeconds)}</span>
      )}
    </span>
  );
}
