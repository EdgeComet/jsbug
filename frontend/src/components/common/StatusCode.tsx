import styles from './StatusCode.module.css';

interface StatusCodeProps {
  code: number | null;
  className?: string;
}

function getStatusClass(code: number | null): string {
  if (code === null) return styles.networkError;
  if (code >= 200 && code < 300) return styles.success;
  if (code >= 300 && code < 400) return styles.redirect;
  if (code >= 400 && code < 500) return styles.clientError;
  if (code >= 500) return styles.serverError;
  return '';
}

export function StatusCode({ code, className }: StatusCodeProps) {
  const statusClass = getStatusClass(code);
  const displayValue = code !== null ? code.toString() : 'N/A';

  return (
    <span className={`${styles.statusCode} ${statusClass} ${className || ''}`}>
      {displayValue}
    </span>
  );
}
