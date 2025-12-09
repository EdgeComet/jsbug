import type { IndexationData } from '../../types/content';
import { TextValue } from '../common/TextValue';
import { UrlLink } from '../common/UrlLink';
import { YesNo } from '../common/YesNo';
import { HrefLangDiffValue, hrefLangsEqual } from '../common/HrefLangDiffValue';
import { isCanonical } from '../../utils/indexationUtils';
import styles from './Panel.module.css';

interface IndexationCardProps {
  data: IndexationData;
  compareData?: IndexationData;
  robotsAllowed?: boolean;
  robotsLoading?: boolean;
  currentUrl?: string;
}

export function IndexationCard({ data, compareData, robotsAllowed, robotsLoading, currentUrl }: IndexationCardProps) {
  // Calculate effective indexability considering robots.txt
  // Formula: isIndexable = robots_allowed AND meta_indexable (via data.isIndexable)
  const robotsKnown = robotsAllowed !== undefined;
  const effectiveIndexable = robotsKnown ? (robotsAllowed && data.isIndexable) : undefined;

  return (
    <div className={styles.resultCard}>
      <div className={styles.resultCardHeader}>
        <span className={styles.resultCardTitle}>Indexation</span>
      </div>
      <div className={styles.resultCardBody}>
        <div className={`${styles.resultRow} ${styles.resultRowMultiline} ${compareData && effectiveIndexable !== undefined && effectiveIndexable !== compareData.isIndexable ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Indexable</span>
          <div className={styles.resultValueMultiline}>
            <span className={styles.resultValue}>
              {effectiveIndexable === undefined ? (
                <TextValue value="Pending..." />
              ) : (
                <YesNo value={effectiveIndexable} />
              )}
            </span>
            <div className={styles.indexableFactors}>
              <div className={styles.indexableFactor}>
                <span className={styles.indexableFactorLabel}>Meta Index</span>
                <YesNo value={data.metaIndexable} />
                {compareData && data.metaIndexable !== compareData.metaIndexable && (
                  <span className={styles.indexableFactorChangedBadge}>Changed</span>
                )}
              </div>
              <div className={styles.indexableFactor}>
                <span className={styles.indexableFactorLabel}>Robots.txt</span>
                {robotsLoading ? (
                  <TextValue value="..." />
                ) : robotsAllowed !== undefined ? (
                  <>
                    <YesNo value={robotsAllowed} />
                    {compareData && compareData.robotsAllowed !== undefined && robotsAllowed !== compareData.robotsAllowed && (
                      <span className={styles.indexableFactorChangedBadge}>Changed</span>
                    )}
                  </>
                ) : (
                  <TextValue value="-" />
                )}
              </div>
              <div className={styles.indexableFactor}>
                <span className={styles.indexableFactorLabel}>Canonical</span>
                <YesNo value={isCanonical(data.canonicalUrl, currentUrl)} />
                {compareData && isCanonical(data.canonicalUrl, currentUrl) !== isCanonical(compareData.canonicalUrl, currentUrl) && (
                  <span className={styles.indexableFactorChangedBadge}>Changed</span>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className={`${styles.resultRow} ${compareData && data.metaFollow !== compareData.metaFollow ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Meta Follow</span>
          <span className={styles.resultValue}>
            <YesNo value={data.metaFollow} />
          </span>
        </div>

        <div className={`${styles.resultRow} ${compareData && data.canonicalUrl !== compareData.canonicalUrl ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Canonical URL</span>
          <span className={styles.resultValue}>
            {data.canonicalUrl ? (
              <UrlLink url={data.canonicalUrl} />
            ) : (
              <TextValue value="" />
            )}
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
