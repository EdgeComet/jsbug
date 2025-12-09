import styles from './LoadingOverlay.module.css';

interface LoadingOverlayProps {
  isVisible: boolean;
  status?: string;
}

export function LoadingOverlay({ isVisible, status }: LoadingOverlayProps) {
  if (!isVisible) return null;

  return (
    <div className={styles.loadingOverlay}>
      <div className={styles.loadingContent}>
        <div className={styles.loadingSpinner}></div>
        <span className={styles.loadingText}>Rendering page...</span>
        <div className={styles.loadingProgress}>
          <div className={styles.progressBar}></div>
        </div>
        {status && <span className={styles.loadingStatus}>{status}</span>}
      </div>
    </div>
  );
}
