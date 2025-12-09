import { useState } from 'react';
import styles from './ArrayDiffValue.module.css';

interface ArrayDiffValueProps {
  value: string[];
  compareValue?: string[];
  listClassName?: string;
  showLineNumbers?: boolean;
  maxItems?: number;
}

export function ArrayDiffValue({ value, compareValue, listClassName, showLineNumbers = false, maxItems }: ArrayDiffValueProps) {
  const [expanded, setExpanded] = useState(false);

  if (value.length === 0 && (!compareValue || compareValue.length === 0)) {
    return <span className={styles.empty}>empty</span>;
  }

  // Only show line numbers when there's more than 1 item
  const shouldShowNumbers = showLineNumbers && value.length > 1;
  const containerClass = [listClassName, shouldShowNumbers ? styles.numbered : ''].filter(Boolean).join(' ');

  // Calculate truncation
  const totalItems = value.length;
  const hasViewMore = maxItems && totalItems > maxItems;
  const shouldTruncate = hasViewMore && !expanded;
  const displayItems = shouldTruncate ? value.slice(0, maxItems) : value;
  const remainingCount = totalItems - (maxItems || 0);

  // Calculate diff counts for summary
  const compareSet = compareValue ? new Set(compareValue) : null;
  const valueSet = new Set(value);
  const addedCount = compareSet ? value.filter(v => !compareSet.has(v)).length : 0;
  const removedCount = compareValue ? compareValue.filter(v => !valueSet.has(v)).length : 0;

  // Summary line shown when view more is available
  const summaryLine = hasViewMore ? (
    <div className={styles.summary}>
      <span>{totalItems} items</span>
      {addedCount > 0 && <span className={styles.summaryAdded}>+{addedCount} added</span>}
      {removedCount > 0 && <span className={styles.summaryRemoved}>-{removedCount} removed</span>}
    </div>
  ) : null;

  if (!compareValue) {
    // No comparison, just render normally
    if (value.length === 0) {
      return <span className={styles.empty}>empty</span>;
    }
    return (
      <div>
        {summaryLine}
        <div className={containerClass}>
          {displayItems.map((item, index) => (
            <span key={index}>{item}</span>
          ))}
          {shouldTruncate && (
            <button type="button" className={styles.viewMore} onClick={() => setExpanded(true)}>
              view more ({remainingCount} more)
            </button>
          )}
        </div>
      </div>
    );
  }

  // Items only in compareValue (missing from this panel)
  const missing = compareValue.filter(v => !valueSet.has(v));

  return (
    <div>
      {summaryLine}
      <div className={containerClass}>
        {displayItems.map((item, i) => (
          <span key={i}>
            <span className={!compareSet!.has(item) ? styles.added : ''}>{item}</span>
          </span>
        ))}
        {shouldTruncate && (
          <button type="button" className={styles.viewMore} onClick={() => setExpanded(true)}>
            view more ({remainingCount} more)
          </button>
        )}
        {!shouldTruncate && missing.map((item, i) => (
          <span key={`missing-${i}`}>
            <span className={styles.missing}>{item}</span>
          </span>
        ))}
      </div>
    </div>
  );
}

// Helper to check if arrays differ
export function arraysEqual(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  const setB = new Set(b);
  return a.every(item => setB.has(item));
}
