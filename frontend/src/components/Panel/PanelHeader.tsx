import { useConfig } from '../../context/ConfigContext';
import { userAgentLabels } from '../../utils/panelLabel';
import styles from './Panel.module.css';

interface PanelHeaderProps {
  side: 'left' | 'right';
}

export function PanelHeader({ side }: PanelHeaderProps) {
  const { config } = useConfig();
  const panelConfig = side === 'left' ? config.left : config.right;
  const isJs = panelConfig.jsEnabled;

  return (
    <div className={`${styles.panelHeader} ${isJs ? styles.panelHeaderJsEnabled : styles.panelHeaderJsDisabled}`}>
      <span className={`${styles.panelLabel} ${isJs ? styles.panelLabelJs : styles.panelLabelNoJs}`}>
        {isJs ? 'JS Rendered' : 'Non JS'}
      </span>
      <span className={styles.panelStatus}>
        {userAgentLabels[panelConfig.userAgent] || panelConfig.userAgent}
        {' • '}
        {panelConfig.timeout}s timeout
      </span>
    </div>
  );
}
