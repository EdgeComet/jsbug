import { useMemo } from 'react';
import { Icon } from '../common/Icon';
import { Button } from '../common/Button';
import styles from './ResultTabs.module.css';

interface HTMLTabProps {
  html: string;
  onOpenModal: () => void;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1048576) return `${Math.round(bytes / 1024)}KB`;
  return `${(bytes / 1048576).toFixed(1)}MB`;
}

function getSizeClass(bytes: number): string {
  if (bytes < 204800) return styles.htmlSizeExcellent;     // < 200 KB
  if (bytes < 512000) return styles.htmlSizeModerate;     // < 500 KB
  if (bytes < 1048576) return styles.htmlSizeWarning;     // < 1 MB
  return styles.htmlSizeCritical;
}

export function HTMLTab({ html, onOpenModal }: HTMLTabProps) {
  const pageSize = useMemo(() => new Blob([html]).size, [html]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(html);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  const handleDownload = () => {
    const blob = new Blob([html], { type: 'text/html' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'rendered.html';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <div className={styles.htmlPreview}>
      <div className={styles.htmlToolbar}>
        <button
          className={`${styles.htmlSizeButton} ${getSizeClass(pageSize)}`}
          onClick={onOpenModal}
        >
          <Icon name="file-code" size={14} />
          <span>Open source {formatBytes(pageSize)}</span>
        </button>
        <Button variant="ghost" size="sm" onClick={handleCopy}>
          <Icon name="clipboard" size={14} />
          Copy
        </Button>
        <Button variant="ghost" size="sm" onClick={handleDownload}>
          <Icon name="download" size={14} />
          Download
        </Button>
      </div>
    </div>
  );
}
