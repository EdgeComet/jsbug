import { useMemo } from 'react';
import styles from './Panel.module.css';

interface AICardProps {
  tokenCount: number;
  compareTokenCount?: number;
}

export function AICard({ tokenCount, compareTokenCount }: AICardProps) {
  const comparison = useMemo(() => {
    if (compareTokenCount === undefined || compareTokenCount === null) {
      return null;
    }

    if (compareTokenCount === 0) {
      // Don't show percentage when comparing against zero
      return null;
    }

    const percentChange = ((tokenCount - compareTokenCount) / compareTokenCount) * 100;
    return {
      percent: Math.round(percentChange),
      isPositive: percentChange >= 0,
    };
  }, [tokenCount, compareTokenCount]);

  const hasChange = comparison !== null && comparison.percent !== 0;

  return (
    <div className={styles.resultCard}>
      <div className={styles.resultCardHeader}>
        <span className={styles.resultCardTitle}>AI</span>
      </div>
      <div className={styles.resultCardBody}>
        <div className={`${styles.resultRow} ${hasChange ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Content Tokens</span>
          <span className={styles.resultValue}>
            {tokenCount.toLocaleString()}
            {comparison && comparison.percent !== 0 && (
              <span className={styles.wordDiffNumbers}>
                {' '}
                <span className={comparison.isPositive ? styles.wordDiffAdded : styles.wordDiffRemoved}>
                  ({comparison.isPositive ? '+' : ''}{comparison.percent}%)
                </span>
              </span>
            )}
          </span>
        </div>
      </div>
    </div>
  );
}
