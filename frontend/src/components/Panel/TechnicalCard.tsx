import type { TechnicalData } from '../../types/content';
import { LoadTime } from '../common/LoadTime';
import { PageSize } from '../common/PageSize';
import { StatusCode } from '../common/StatusCode';
import styles from './Panel.module.css';

interface TechnicalCardProps {
  data: TechnicalData;
  compareData?: TechnicalData;
  onOpenHTMLModal?: () => void;
  userAgent?: string;
  onRetryWithBrowserUA?: () => void;
}

export function TechnicalCard({ data, compareData, onOpenHTMLModal, userAgent, onRetryWithBrowserUA }: TechnicalCardProps) {
  const hasError = data.errorMessage !== undefined;
  const isSuccess = data.statusCode === 200;
  const isRedirect =
    data.statusCode !== null &&
    data.statusCode >= 300 &&
    data.statusCode < 400;
  const isGooglebot = userAgent === 'googlebot' || userAgent === 'googlebot-mobile';
  const isClientError = data.statusCode !== null && data.statusCode >= 400 && data.statusCode < 500;
  const showGooglebotWarning = isClientError && isGooglebot;

  return (
    <div
      className={`${styles.resultCard} ${hasError ? styles.resultCardError : ''}`}
    >
      <div className={styles.resultCardHeader}>
        <span className={styles.resultCardTitle}>Technical</span>
      </div>
      <div className={styles.resultCardBody}>
        <div className={`${styles.resultRow} ${!isSuccess ? styles.diffHighlightChanged : ''}`}>
          <span className={styles.resultLabel}>Status Code</span>
          <span className={styles.resultValue}>
            <StatusCode code={data.statusCode} />
          </span>
        </div>

        {showGooglebotWarning && (
          <div className={styles.googlebotWarning}>
            Requests with Googlebot user agents often get blocked.
            {onRetryWithBrowserUA && (
              <button className={styles.retryButton} onClick={onRetryWithBrowserUA}>
                Retry with browser UA
              </button>
            )}
          </div>
        )}

        {data.errorMessage && (
          <div className={`${styles.resultRow} ${styles.errorRow}`}>
            <span className={styles.resultLabel}>Error</span>
            <span className={`${styles.resultValue} ${styles.errorMessage}`}>
              {data.errorMessage}
            </span>
          </div>
        )}

        {isRedirect && data.redirectUrl && (
          <div className={styles.resultRow}>
            <span className={styles.resultLabel}>Redirects To</span>
            <span className={styles.resultValue}>{data.redirectUrl}</span>
          </div>
        )}

        <div className={styles.resultRow}>
          <span className={styles.resultLabel}>Page Size</span>
          <span className={styles.resultValue}>
            <PageSize bytes={data.pageSize} compareBytes={compareData?.pageSize} onClick={onOpenHTMLModal} />
          </span>
        </div>

        <div className={styles.resultRow}>
          <span className={styles.resultLabel}>Load Time</span>
          <span className={styles.resultValue}>
            <LoadTime seconds={data.loadTime} compareSeconds={compareData?.loadTime} />
          </span>
        </div>
      </div>
    </div>
  );
}
