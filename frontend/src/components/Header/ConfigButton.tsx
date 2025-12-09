import { Icon } from '../common/Icon';
import styles from './Header.module.css';

interface ConfigButtonProps {
  onClick: () => void;
}

export function ConfigButton({ onClick }: ConfigButtonProps) {
  return (
    <button
      className={styles.btnConfig}
      onClick={onClick}
      title="Configure render settings"
      aria-label="Configure render settings"
    >
      <Icon name="sliders" size={18} />
    </button>
  );
}
