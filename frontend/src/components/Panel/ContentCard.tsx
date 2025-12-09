import { useMemo } from 'react';
import type { ContentData } from '../../types/content';
import { TextValue } from '../common/TextValue';
import { ArrayDiffValue, arraysEqual } from '../common/ArrayDiffValue';
import { computeWordDiff } from '../../utils/wordDiff';
import { getSchemaTypeCounts, schemaTypesEqual, getSchemaTypesList } from '../../utils/schemaUtils';
import styles from './Panel.module.css';
import arrayStyles from '../common/ArrayDiffValue.module.css';

interface ContentCardProps {
  data: ContentData;
  compareData?: ContentData;
  onOpenBodyTextModal?: () => void;
  onOpenWordDiffModal?: (scrollTo: 'added' | 'removed') => void;
}

export function ContentCard({ data, compareData, onOpenBodyTextModal, onOpenWordDiffModal }: ContentCardProps) {
  const wordDiff = useMemo(() => {
    if (!compareData) return null;
    return computeWordDiff(data.bodyText, compareData.bodyText);
  }, [data.bodyText, compareData?.bodyText]);

  const hasWordDiff = wordDiff && (wordDiff.added.length > 0 || wordDiff.removed.length > 0);

  return (
    <div className={styles.resultCard}>
      <div className={styles.resultCardHeader}>
        <span className={styles.resultCardTitle}>Content</span>
      </div>
      <div className={styles.resultCardBody}>
        <div className={`${styles.resultRow} ${compareData && data.title !== compareData.title ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Title</span>
          <span className={styles.resultValue}>
            <TextValue value={data.title} />
          </span>
        </div>

        <div className={`${styles.resultRow} ${compareData && data.metaDescription !== compareData.metaDescription ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Meta Description</span>
          <span className={styles.resultValue}>
            <TextValue value={data.metaDescription} />
          </span>
        </div>

        <div className={`${styles.resultRow} ${compareData && !arraysEqual(data.h1, compareData.h1) ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>H1</span>
          <span className={styles.resultValue}>
            <ArrayDiffValue value={data.h1} compareValue={compareData?.h1} listClassName={styles.headingsList} showLineNumbers maxItems={10} />
          </span>
        </div>

        <div className={`${styles.resultRow} ${compareData && !arraysEqual(data.h2, compareData.h2) ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>H2</span>
          <span className={styles.resultValue}>
            <ArrayDiffValue value={data.h2} compareValue={compareData?.h2} listClassName={styles.headingsList} showLineNumbers maxItems={10} />
          </span>
        </div>

        <div className={`${styles.resultRow} ${compareData && !arraysEqual(data.h3, compareData.h3) ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>H3</span>
          <span className={styles.resultValue}>
            <ArrayDiffValue value={data.h3} compareValue={compareData?.h3} listClassName={styles.headingsList} showLineNumbers maxItems={10} />
          </span>
        </div>

        <div className={`${styles.resultRow} ${hasWordDiff ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Body Words</span>
          <span className={styles.resultValue}>
            <button
              type="button"
              className={styles.resultValueClickable}
              onClick={() => onOpenBodyTextModal?.()}
            >
              {data.bodyWords.toLocaleString()}
            </button>
            {hasWordDiff && (
              <span className={styles.wordDiffNumbers}>
                <button
                  type="button"
                  className={styles.wordDiffAdded}
                  onClick={() => onOpenWordDiffModal?.('added')}
                >
                  +{wordDiff.added.length}
                </button>
                <button
                  type="button"
                  className={styles.wordDiffRemoved}
                  onClick={() => onOpenWordDiffModal?.('removed')}
                >
                  -{wordDiff.removed.length}
                </button>
              </span>
            )}
          </span>
        </div>

        <div className={styles.resultRow}>
          <span className={styles.resultLabel}>Text/HTML Ratio</span>
          <span className={styles.resultValue}>{(data.textHtmlRatio * 100).toFixed(0)}%</span>
        </div>

        {data.structuredData && data.structuredData.length > 0 && (() => {
          const counts = getSchemaTypeCounts(data.structuredData);
          const compareCounts = compareData?.structuredData
            ? getSchemaTypeCounts(compareData.structuredData)
            : undefined;
          // Only show diff styling when compareData exists (comparing panels)
          const types = getSchemaTypesList(counts, compareData ? compareCounts : undefined);
          const hasDiff = compareCounts && !schemaTypesEqual(counts, compareCounts);

          return (
            <div className={`${styles.resultRow} ${hasDiff ? styles.diffHighlightChanged : ''}`}>
              <span className={styles.resultLabel}>Schema</span>
              <span className={styles.resultValue}>
                {types.map((t, i) => (
                  <span key={i}>
                    {i > 0 && ', '}
                    <span className={
                      t.status === 'added' ? arrayStyles.added :
                      t.status === 'removed' ? arrayStyles.missing : ''
                    }>
                      {t.count > 1 ? `${t.type} (${t.count})` : t.type}
                    </span>
                  </span>
                ))}
                {counts.invalid > 0 && (
                  <span>, invalid ({counts.invalid})</span>
                )}
              </span>
            </div>
          );
        })()}
      </div>
    </div>
  );
}
