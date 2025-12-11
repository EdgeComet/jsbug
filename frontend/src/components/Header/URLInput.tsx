import { useState, useEffect, forwardRef } from 'react';
import { validateHttpUrl } from '../../utils/urlValidation';
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
    const result = validateHttpUrl(value);
    onValidChange?.(result.valid);
  }, [value, onValidChange]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    onChange(newValue);

    if (touched) {
      const result = validateHttpUrl(newValue);
      setError(result.valid ? null : result.error ?? null);
    }
  };

  const handleBlur = () => {
    setTouched(true);
    const result = validateHttpUrl(value);
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
