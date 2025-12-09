import { useState } from 'react';
import type { NetworkData } from '../../types/network';
import type { TimelineData } from '../../types/content';
import type { ConsoleEntry } from '../../types/console';
import { NetworkTab } from './NetworkTab';
import { TimelineTab } from './TimelineTab';
import { ConsoleTab } from './ConsoleTab';
import { HTMLTab } from './HTMLTab';
import styles from './ResultTabs.module.css';

type TabId = 'network' | 'timeline' | 'console' | 'html';

interface ResultTabsProps {
  networkData: NetworkData;
  timelineData: TimelineData;
  consoleData: ConsoleEntry[];
  htmlData: string;
  onOpenHTMLModal: () => void;
}

export function ResultTabs({ networkData, timelineData, consoleData, htmlData, onOpenHTMLModal }: ResultTabsProps) {
  const [activeTab, setActiveTab] = useState<TabId>('network');

  const tabs: { id: TabId; label: string; count?: number; warn?: boolean }[] = [
    { id: 'network', label: 'Network', count: networkData.requests.length },
    { id: 'timeline', label: 'Timeline' },
    { id: 'console', label: 'Console', count: consoleData.length, warn: consoleData.some(e => e.level === 'error' || e.level === 'warn') },
    { id: 'html', label: 'HTML' },
  ];

  return (
    <div className={styles.resultTabs}>
      {/* Tab buttons */}
      <div className={styles.tabsHeader}>
        {tabs.map((tab) => (
          <button
            key={tab.id}
            className={`${styles.tabBtn} ${activeTab === tab.id ? styles.tabBtnActive : ''}`}
            onClick={() => setActiveTab(tab.id)}
            data-tab={tab.id}
          >
            {tab.label}
            {tab.count !== undefined && (
              <span className={`${styles.tabCount} ${tab.warn ? styles.tabCountWarn : ''}`}>
                {tab.count}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className={styles.tabContentWrapper}>
        {activeTab === 'network' && (
          <div className={styles.tabContent} data-tab="network">
            <NetworkTab data={networkData} />
          </div>
        )}
        {activeTab === 'timeline' && (
          <div className={styles.tabContent} data-tab="timeline">
            <TimelineTab data={timelineData} />
          </div>
        )}
        {activeTab === 'console' && (
          <div className={styles.tabContent} data-tab="console">
            <ConsoleTab data={consoleData} />
          </div>
        )}
        {activeTab === 'html' && (
          <div className={styles.tabContent} data-tab="html">
            <HTMLTab html={htmlData} onOpenModal={onOpenHTMLModal} />
          </div>
        )}
      </div>
    </div>
  );
}
