import { Icon } from '../common/Icon';
import styles from './Header.module.css';

export function Logo() {
  return (
    <a href="/" className={styles.headerBrand}>
      <span className={styles.brandIcon}>
        <Icon name="bug" size={25} />
      </span>
      <span className={styles.brandName}>jsbug</span>
    </a>
  );
}
