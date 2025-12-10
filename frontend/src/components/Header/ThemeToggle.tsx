import { Icon } from '../common/Icon/Icon';
import { useTheme } from '../../context/ThemeContext';
import styles from './ThemeToggle.module.css';

export function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();

  const handleToggle = () => {
    setTheme(resolvedTheme === 'light' ? 'dark' : 'light');
  };

  return (
    <button
      className={styles.themeToggle}
      onClick={handleToggle}
      title={`Switch to ${resolvedTheme === 'light' ? 'dark' : 'light'} mode`}
      aria-label={`Switch to ${resolvedTheme === 'light' ? 'dark' : 'light'} mode`}
    >
      <Icon name={resolvedTheme === 'light' ? 'moon' : 'sun'} size={18} />
    </button>
  );
}
