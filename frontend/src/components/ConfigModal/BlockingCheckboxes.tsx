import styles from './ConfigModal.module.css';

interface BlockingState {
  imagesMedia: boolean;
  trackingScripts: boolean;
}

interface BlockingCheckboxesProps {
  blocking: BlockingState;
  onChange: (blocking: BlockingState) => void;
  disabled?: boolean;
}

export function BlockingCheckboxes({ blocking, onChange, disabled }: BlockingCheckboxesProps) {
  const handleChange = (key: keyof BlockingState) => {
    // trackingScripts is always checked and can't be changed
    if (key === 'trackingScripts') return;

    onChange({
      ...blocking,
      [key]: !blocking[key],
    });
  };

  const items = [
    { key: 'imagesMedia' as const, label: 'Images & Media', locked: false },
    { key: 'trackingScripts' as const, label: 'Tracking scripts', locked: true },
  ];

  return (
    <div className={`${styles.checkboxGroup} ${styles.checkboxInline} ${disabled ? styles.disabled : ''}`}>
      {items.map((item) => (
        <label key={item.key} className={`${styles.checkboxItem} ${item.locked ? styles.checkboxLocked : ''}`}>
          <input
            type="checkbox"
            checked={blocking[item.key]}
            onChange={() => handleChange(item.key)}
            disabled={disabled || item.locked}
          />
          <span className={styles.checkboxLabel}>{item.label}</span>
        </label>
      ))}
    </div>
  );
}
