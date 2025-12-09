import { useState, useMemo } from 'react';
import { Icon } from '../common/Icon';
import type { ConsoleEntry } from '../../types/console';
import { ConsoleModal } from './ConsoleModal';
import styles from './ResultTabs.module.css';

interface ConsoleTabProps {
  data: ConsoleEntry[];
}

export function ConsoleTab({ data }: ConsoleTabProps) {
  const [modalOpen, setModalOpen] = useState(false);

  const stats = useMemo(() => ({
    total: data.length,
    errors: data.filter(e => e.level === 'error').length,
    warnings: data.filter(e => e.level === 'warn').length,
  }), [data]);

  return (
    <>
      <button className={styles.consoleSummary} onClick={() => setModalOpen(true)}>
        <span className={styles.consoleStat}>
            Open Console:
            &nbsp;{stats.total} {stats.total === 1 ? 'line' : 'lines'}
        </span>
        {stats.errors > 0 && (
          <span className={`${styles.consoleStat} ${styles.consoleStatError}`}>
            <Icon name="x-circle" size={14} />
            {stats.errors} {stats.errors === 1 ? 'error' : 'errors'}
          </span>
        )}
        {stats.warnings > 0 && (
          <span className={`${styles.consoleStat} ${styles.consoleStatWarn}`}>
            <Icon name="alert-triangle" size={14} />
            {stats.warnings} {stats.warnings === 1 ? 'warning' : 'warnings'}
          </span>
        )}
      </button>
      <ConsoleModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        entries={data}
      />
    </>
  );
}
