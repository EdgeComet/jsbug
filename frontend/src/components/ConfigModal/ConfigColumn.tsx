import type { PanelConfig, UserAgent, WaitEvent } from '../../types/config';
import { JSToggle } from './JSToggle';
import { TimeoutSlider } from './TimeoutSlider';
import { BlockingCheckboxes } from './BlockingCheckboxes';
import styles from './ConfigModal.module.css';

interface ConfigColumnProps {
  side: 'left' | 'right';
  config: PanelConfig;
  onChange: (config: PanelConfig) => void;
  customUaError?: boolean;
}

const userAgentGroups = [
  {
    label: 'Browsers',
    options: [
      { value: 'chrome-mobile', label: 'Chrome Mobile' },
      { value: 'chrome', label: 'Chrome Desktop' },
    ],
  },
  {
    label: 'Googlebots',
    options: [
      { value: 'googlebot', label: 'Googlebot' },
      { value: 'googlebot-mobile', label: 'Googlebot Mobile' },
    ],
  },
  {
    label: 'Bing',
    options: [
      { value: 'bingbot', label: 'Bingbot' },
    ],
  },
  {
    label: 'AI Bots',
    options: [
      { value: 'claudebot', label: 'ClaudeBot' },
      { value: 'claude-user', label: 'Claude-User' },
      { value: 'gptbot', label: 'GPTBot' },
      { value: 'chatgpt-user', label: 'ChatGPT-User' },
    ],
  },
  {
    label: 'Other',
    options: [
      { value: 'custom', label: 'Custom...' },
    ],
  },
];

const isGooglebot = (ua: string) => ua === 'googlebot' || ua === 'googlebot-mobile';

const waitEventOptions = [
  { value: 'DOMContentLoaded', label: 'DOM Content Loaded' },
  { value: 'load', label: 'Load' },
  { value: 'networkIdle', label: 'Network Idle' },
  { value: 'networkAlmostIdle', label: 'Network Almost Idle' },
];

export function ConfigColumn({ side, config, onChange, customUaError }: ConfigColumnProps) {
  const isLeft = side === 'left';

  return (
    <div className={styles.configColumn}>
      <div className={`${styles.configColumnHeader} ${config.jsEnabled ? styles.configColumnJsEnabled : styles.configColumnJsDisabled}`}>
        <span className={styles.columnTitle}>{isLeft ? 'Left Panel' : 'Right Panel'}</span>
      </div>

      <div className={styles.configColumnBody}>
        {/* JavaScript Toggle */}
        <div className={styles.configRow}>
          <span className={styles.configLabel}>JavaScript</span>
          <JSToggle
            enabled={config.jsEnabled}
            onChange={(jsEnabled) => onChange({ ...config, jsEnabled })}
          />
        </div>

        {/* User Agent Select */}
        <div className={styles.configRow}>
          <span className={styles.configLabel}>User Agent</span>
          <select
            className={styles.configSelect}
            value={config.userAgent}
            onChange={(e) => onChange({ ...config, userAgent: e.target.value as UserAgent })}
          >
            {userAgentGroups.map((group) => (
              <optgroup key={group.label} label={group.label}>
                {group.options.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </optgroup>
            ))}
          </select>
          {isGooglebot(config.userAgent) && (
            <div className={styles.googlebotWarning}>
              Googlebots user-agents often get blocked by websites as fake Googlebots
            </div>
          )}
          {config.userAgent === 'custom' && (
            <input
              type="text"
              className={`${styles.configInput} ${customUaError ? styles.configInputError : ''}`}
              value={config.customUserAgent || ''}
              onChange={(e) => onChange({ ...config, customUserAgent: e.target.value })}
              placeholder="Enter custom user agent..."
            />
          )}
        </div>

        {/* Timeout Slider */}
        <div className={styles.configRow}>
          <span className={styles.configLabel}>Timeout</span>
          <TimeoutSlider
            value={config.timeout}
            onChange={(timeout) => onChange({ ...config, timeout })}
          />
        </div>

        {/* Wait For Select */}
        <div className={`${styles.configRow} ${!config.jsEnabled ? styles.configRowDisabled : ''}`}>
          <span className={styles.configLabel}>Wait For</span>
          <select
            className={styles.configSelect}
            value={config.waitFor}
            onChange={(e) => onChange({ ...config, waitFor: e.target.value as WaitEvent })}
            disabled={!config.jsEnabled}
          >
            {waitEventOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* Blocking Checkboxes */}
        <div className={`${styles.configRow} ${!config.jsEnabled ? styles.configRowDisabled : ''}`}>
          <span className={styles.configLabel}>Block</span>
          <BlockingCheckboxes
            blocking={config.blocking}
            onChange={(blocking) => onChange({ ...config, blocking })}
            disabled={!config.jsEnabled}
          />
        </div>
      </div>
    </div>
  );
}
