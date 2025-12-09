import { useState, useEffect, useMemo, useRef } from 'react';
import { Icon, IconName } from '../common/Icon';
import type { ConsoleEntry } from '../../types/console';
import { highlightText } from '../../utils/highlightText';
import { formatTimestamp } from '../../utils/networkUtils';
import { Modal } from '../common/Modal';
import { FilterInput } from '../common/FilterInput';
import { MatchCountBadge } from '../common/MatchCountBadge/MatchCountBadge';
import { EmptyState } from '../common/EmptyState/EmptyState';
import filterStyles from '../common/Modal/ModalFilters.module.css';
import highlightStyles from '../../styles/highlight.module.css';
import styles from './ConsoleModal.module.css';

interface ConsoleModalProps {
  isOpen: boolean;
  onClose: () => void;
  entries: ConsoleEntry[];
}

const iconNameMap: Record<string, IconName> = {
  log: 'terminal',
  warn: 'alert-triangle',
  error: 'x-circle',
};

const levelClassMap = {
  log: styles.entryLog,
  warn: styles.entryWarn,
  error: styles.entryError,
};

export function ConsoleModal({ isOpen, onClose, entries }: ConsoleModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    if (isOpen) {
      setSearchTerm('');
    }
  }, [isOpen]);

  const { filteredEntries, matchCount } = useMemo(() => {
    if (!searchTerm.trim()) {
      return { filteredEntries: entries, matchCount: 0 };
    }

    const term = searchTerm.toLowerCase();
    const filtered = entries.filter(entry =>
      entry.message.toLowerCase().includes(term)
    );

    return { filteredEntries: filtered, matchCount: filtered.length };
  }, [entries, searchTerm]);

  const headerExtra = searchTerm && matchCount > 0 ? (
    <MatchCountBadge count={matchCount} />
  ) : null;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Console Output (${entries.length.toLocaleString()} entries)`}
      size="xl"
      headerExtra={headerExtra}
      searchInputRef={searchInputRef}
    >
      <div className={filterStyles.modalFilters}>
        <FilterInput
          ref={searchInputRef}
          placeholder="Search console messages..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </div>

      <div className={styles.modalBody}>
        {filteredEntries.length === 0 ? (
          <EmptyState
            icon={<Icon name="search" size={24} />}
            message="No matching entries found"
          />
        ) : (
          filteredEntries.map((entry) => {
            const iconName = iconNameMap[entry.level] ?? 'terminal';
            const levelClass = levelClassMap[entry.level] ?? styles.entryLog;
            return (
              <div key={entry.id} className={`${styles.consoleEntry} ${levelClass}`}>
                <span className={styles.entryIcon}>
                  <Icon name={iconName} size={14} />
                </span>
                <span className={styles.entryMessage}>
                  {highlightText(entry.message, searchTerm, highlightStyles.highlight)}
                </span>
                <span className={styles.entryTime}>{formatTimestamp(entry.time)}</span>
              </div>
            );
          })
        )}
      </div>
    </Modal>
  );
}
