import { useState, useEffect, forwardRef } from 'react';
import styles from './Header.module.css';

interface URLInputProps {
  value: string;
  onChange: (value: string) => void;
  onValidChange?: (isValid: boolean) => void;
  onSubmit?: () => void;
}

export const URLInput = forwardRef<HTMLInputElement, URLInputProps>(function URLInput({ value, onChange, onValidChange, onSubmit }, ref) {
  const [error, setError] = useState<string | null>(null);
  const [touched, setTouched] = useState(false);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      onSubmit?.();
    }
  };

  useEffect(() => {
    const result = validateUrl(value);
    onValidChange?.(result.valid);
  }, [value, onValidChange]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    onChange(newValue);

    if (touched) {
      const result = validateUrl(newValue);
      setError(result.valid ? null : result.error ?? null);
    }
  };

  const handleBlur = () => {
    setTouched(true);
    const result = validateUrl(value);
    setError(result.valid ? null : result.error ?? null);
  };

  return (
    <div className={styles.headerUrl}>
      <input
        ref={ref}
        type="url"
        className={`${styles.urlInput} ${error ? styles.urlInputError : ''}`}
        placeholder="Enter URL to compare (e.g., https://example.com)"
        value={value}
        onChange={handleChange}
        onBlur={handleBlur}
        onKeyDown={handleKeyDown}
      />
      {error && <div className={styles.urlErrorMessage}>{error}</div>}
    </div>
  );
})

function validateUrl(string: string): { valid: boolean; error?: string } {
  if (!string.trim()) {
    return { valid: false, error: 'URL is required' };
  }

  try {
    const url = new URL(string);

    if (!['http:', 'https:'].includes(url.protocol)) {
      return { valid: false, error: 'URL must use http or https' };
    }

    const host = url.hostname.toLowerCase();
    if (host === 'localhost' || host === '127.0.0.1' || host.startsWith('192.168.') || host === '::1') {
      return { valid: false, error: 'Local URLs are not allowed' };
    }

    return { valid: true };
  } catch {
    return { valid: false, error: 'Please enter a valid URL' };
  }
}
