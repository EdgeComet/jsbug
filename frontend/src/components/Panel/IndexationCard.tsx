import type { IndexationData } from '../../types/content';
import { TextValue } from '../common/TextValue';
import { UrlLink } from '../common/UrlLink';
import { YesNo } from '../common/YesNo';
import { HrefLangDiffValue, hrefLangsEqual } from '../common/HrefLangDiffValue';
import styles from './Panel.module.css';

interface IndexationCardProps {
  data: IndexationData;
  compareData?: IndexationData;
}

export function IndexationCard({ data, compareData }: IndexationCardProps) {
  return (
    <div className={styles.resultCard}>
      <div className={styles.resultCardHeader}>
        <span className={styles.resultCardTitle}>Indexation</span>
      </div>
      <div className={styles.resultCardBody}>
        <div className={`${styles.resultRow} ${compareData && data.isIndexable !== compareData.isIndexable ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Indexable</span>
          <span className={styles.resultValue}>
            <YesNo value={data.isIndexable} />
            {!data.isIndexable && <span className={styles.indexabilityReason}> ({data.indexabilityReason})</span>}
          </span>
        </div>

        <div className={`${styles.resultRow} ${compareData && data.metaIndexable !== compareData.metaIndexable ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Meta Indexable</span>
          <span className={styles.resultValue}>
            <YesNo value={data.metaIndexable} />
          </span>
        </div>

        <div className={`${styles.resultRow} ${compareData && data.metaFollow !== compareData.metaFollow ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Meta Follow</span>
          <span className={styles.resultValue}>
            <YesNo value={data.metaFollow} />
          </span>
        </div>

        <div className={styles.resultRow}>
          <span className={styles.resultLabel}>Canonical URL</span>
          <span className={styles.resultValue}>
            {data.canonicalUrl ? <UrlLink url={data.canonicalUrl} /> : <TextValue value="" />}
          </span>
        </div>

        <div className={`${styles.resultRow} ${compareData && !hrefLangsEqual(data.hrefLangs, compareData.hrefLangs) ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Href Langs</span>
          <span className={styles.resultValue}>
            <HrefLangDiffValue value={data.hrefLangs} compareValue={compareData?.hrefLangs} />
          </span>
        </div>
      </div>
    </div>
  );
}
