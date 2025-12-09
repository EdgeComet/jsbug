import { useState, useEffect, useMemo, useRef, useCallback } from 'react';
import { highlightText } from '../../utils/highlightText';
import { Modal } from '../common/Modal';
import { FilterInput } from '../common/FilterInput';
import { MatchCountBadge } from '../common/MatchCountBadge/MatchCountBadge';
import { SearchNavigation } from '../common/SearchNavigation/SearchNavigation';
import filterStyles from '../common/Modal/ModalFilters.module.css';
import highlightStyles from '../../styles/highlight.module.css';
import styles from './BodyTextModal.module.css';

interface BodyTextModalProps {
  isOpen: boolean;
  onClose: () => void;
  bodyText: string;
  wordCount: number;
}

export function BodyTextModal({ isOpen, onClose, bodyText, wordCount }: BodyTextModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [currentMatchIndex, setCurrentMatchIndex] = useState(0);
  const bodyRef = useRef<HTMLDivElement>(null);

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
    const matches = bodyText.match(regex);
    return matches ? matches.length : 0;
  }, [bodyText, searchTerm]);

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

  const headerExtra = searchTerm && matchCount > 0 ? (
    <MatchCountBadge count={matchCount} />
  ) : null;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Body Text (${wordCount.toLocaleString()} words)`}
      size="lg"
      headerExtra={headerExtra}
      searchInputRef={searchInputRef}
    >
      <div className={filterStyles.modalFilters}>
        <FilterInput
          ref={searchInputRef}
          placeholder="Search text..."
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
        <div className={styles.textContent}>
          {highlightText(bodyText, searchTerm, highlightStyles.highlight)}
        </div>
      </div>
    </Modal>
  );
}
