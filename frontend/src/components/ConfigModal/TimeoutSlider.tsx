import styles from './ConfigModal.module.css';

interface TimeoutSliderProps {
  value: number;
  onChange: (value: number) => void;
  disabled?: boolean;
}

export function TimeoutSlider({ value, onChange, disabled }: TimeoutSliderProps) {
  return (
    <div className={`${styles.timeoutControl} ${disabled ? styles.disabled : ''}`}>
      <input
        type="range"
        className={styles.timeoutSlider}
        min={1}
        max={30}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        disabled={disabled}
      />
      <span className={styles.timeoutValue}>{value}s</span>
    </div>
  );
}
