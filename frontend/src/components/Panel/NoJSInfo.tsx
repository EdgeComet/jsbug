import { Icon } from '../common/Icon';
import styles from './Panel.module.css';

export function NoJSInfo() {
  return (
    <div className={styles.noJsInfo}>
      <Icon name="info" size={16} />
      <span>Network, Timeline, and Console data are only available for JS-rendered pages.</span>
    </div>
  );
}
