import { useState, useEffect, useMemo, useRef, useCallback } from 'react';
import { Icon } from '../common/Icon';
import { highlightText } from '../../utils/highlightText';
import { Modal } from '../common/Modal';
import { FilterInput } from '../common/FilterInput';
import { MatchCountBadge } from '../common/MatchCountBadge/MatchCountBadge';
import { SearchNavigation } from '../common/SearchNavigation/SearchNavigation';
import { EmptyState } from '../common/EmptyState/EmptyState';
import filterStyles from '../common/Modal/ModalFilters.module.css';
import highlightStyles from '../../styles/highlight.module.css';
import styles from './HTMLModal.module.css';

interface HTMLModalProps {
  isOpen: boolean;
  onClose: () => void;
  html: string;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1048576) return `${Math.round(bytes / 1024)}KB`;
  return `${(bytes / 1048576).toFixed(1)}MB`;
}

export function HTMLModal({ isOpen, onClose, html }: HTMLModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [currentMatchIndex, setCurrentMatchIndex] = useState(0);
  const bodyRef = useRef<HTMLDivElement>(null);

  const pageSize = useMemo(() => new Blob([html]).size, [html]);

  useEffect(() => {
    if (isOpen) {
      setSearchTerm('');
      setCurrentMatchIndex(0);
    }
  }, [isOpen]);

  useEffect(() => {
    setCurrentMatchIndex(0);
  }, [searchTerm]);

  const matchCount = useMemo(() => {
    if (!searchTerm.trim()) return 0;
    const regex = new RegExp(searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
    const matches = html.match(regex);
    return matches ? matches.length : 0;
  }, [html, searchTerm]);

  const goToNextMatch = useCallback(() => {
    if (matchCount > 0) {
      setCurrentMatchIndex((prev) => (prev + 1) % matchCount);
    }
  }, [matchCount]);

  const goToPrevMatch = useCallback(() => {
    if (matchCount > 0) {
      setCurrentMatchIndex((prev) => (prev - 1 + matchCount) % matchCount);
    }
  }, [matchCount]);

  useEffect(() => {
    if (searchTerm.trim() && bodyRef.current) {
      const matches = bodyRef.current.querySelectorAll('mark');
      matches.forEach((m, i) => {
        m.classList.toggle(highlightStyles.highlightActive, i === currentMatchIndex);
      });
      if (matches[currentMatchIndex]) {
        matches[currentMatchIndex].scrollIntoView({ behavior: 'instant', block: 'center' });
      }
    }
  }, [searchTerm, currentMatchIndex]);

  const title = (
    <>
      HTML Source
      <span className={styles.sizeInfo}>({formatBytes(pageSize)})</span>
    </>
  );

  const headerExtra = searchTerm && matchCount > 0 ? (
    <MatchCountBadge count={matchCount} />
  ) : null;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={title}
      size="xl"
      headerExtra={headerExtra}
      searchInputRef={searchInputRef}
    >
      <div className={filterStyles.modalFilters}>
        <FilterInput
          ref={searchInputRef}
          placeholder="Search HTML source..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault();
              if (e.shiftKey) {
                goToPrevMatch();
              } else {
                goToNextMatch();
              }
            }
          }}
        />
        <SearchNavigation
          currentIndex={currentMatchIndex}
          totalMatches={matchCount}
          onPrevious={goToPrevMatch}
          onNext={goToNextMatch}
        />
      </div>

      <div ref={bodyRef} className={styles.modalBody}>
        {html ? (
          <div className={styles.codeContainer}>
            <pre><code>{highlightText(html, searchTerm, highlightStyles.highlight)}</code></pre>
          </div>
        ) : (
          <EmptyState
            icon={<Icon name="search" size={24} />}
            message="No HTML content"
          />
        )}
      </div>
    </Modal>
  );
}
