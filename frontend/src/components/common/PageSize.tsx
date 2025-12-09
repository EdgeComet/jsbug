import styles from './PageSize.module.css';

interface PageSizeProps {
  bytes: number | null;
  compareBytes?: number | null;
  className?: string;
  onClick?: () => void;
}

function getSizeClass(bytes: number): string {
  if (bytes < 204800) return styles.excellent;     // < 200 KB
  if (bytes < 512000) return styles.moderate;      // < 500 KB
  if (bytes < 1048576) return styles.warning;      // < 1 MB
  return styles.critical;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1048576) return `${Math.round(bytes / 1024)}KB`;
  return `${(bytes / 1048576).toFixed(1)}MB`;
}

export function PageSize({ bytes, compareBytes, className, onClick }: PageSizeProps) {
  if (bytes === null) {
    return <span className={`${styles.pageSize} ${className || ''}`}>-</span>;
  }
  const sizeClass = bytes === 0 ? styles.zero : getSizeClass(bytes);
  const isCritical = bytes >= 1048576;
  const formatted = formatBytes(bytes);

  // Show diff badge if this panel's size is larger than the other
  const showDiff = compareBytes !== null && compareBytes !== undefined && bytes > compareBytes;
  const diffBytes = showDiff ? bytes - compareBytes : 0;

  return (
    <span
      className={`${styles.pageSize} ${sizeClass} ${onClick ? styles.clickable : ''} ${className || ''}`}
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={onClick ? (e) => { if (e.key === 'Enter' || e.key === ' ') onClick(); } : undefined}
    >
      <span>{formatted}</span>
      {isCritical && (
        <span className={styles.dangerIcon} title="Critical page size">
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
            <span className={styles.diffBadge}>+{formatBytes(diffBytes)}</span>
        )}
    </span>
  );
}
