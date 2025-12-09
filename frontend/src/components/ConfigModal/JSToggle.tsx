import styles from './ConfigModal.module.css';

interface JSToggleProps {
  enabled: boolean;
  onChange: (enabled: boolean) => void;
}

export function JSToggle({ enabled, onChange }: JSToggleProps) {
  return (
    <div className={styles.segmentedControl} role="group">
      <button
        type="button"
        className={`${styles.segment} ${enabled ? styles.segmentOn : ''}`}
        onClick={() => onChange(true)}
        aria-pressed={enabled}
      >
        ON
      </button>
      <button
        type="button"
        className={`${styles.segment} ${!enabled ? styles.segmentOff : ''}`}
        onClick={() => onChange(false)}
        aria-pressed={!enabled}
      >
        OFF
      </button>
    </div>
  );
}
