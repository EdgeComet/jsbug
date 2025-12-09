import { useState, useEffect, useMemo, useRef } from 'react';
import { Modal } from '../common/Modal';
import { FilterInput } from '../common/FilterInput';
import filterStyles from '../common/Modal/ModalFilters.module.css';
import styles from './WordDiffModal.module.css';

interface WordDiffModalProps {
  isOpen: boolean;
  onClose: () => void;
  addedWords: string[];
  removedWords: string[];
  scrollTo?: 'added' | 'removed';
}

export function WordDiffModal({ isOpen, onClose, addedWords, removedWords, scrollTo }: WordDiffModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const removedSectionRef = useRef<HTMLDivElement>(null);
  const [filter, setFilter] = useState('');

  useEffect(() => {
    if (isOpen) {
      setFilter('');
    }
  }, [isOpen]);

  useEffect(() => {
    if (isOpen && scrollTo === 'removed' && removedSectionRef.current) {
      removedSectionRef.current.scrollIntoView({ behavior: 'instant' });
    }
  }, [isOpen, scrollTo]);

  const filteredAdded = useMemo(() => {
    if (!filter) return addedWords;
    const lowerFilter = filter.toLowerCase();
    return addedWords.filter(word => word.toLowerCase().includes(lowerFilter));
  }, [addedWords, filter]);

  const filteredRemoved = useMemo(() => {
    if (!filter) return removedWords;
    const lowerFilter = filter.toLowerCase();
    return removedWords.filter(word => word.toLowerCase().includes(lowerFilter));
  }, [removedWords, filter]);

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Word Differences" size="md" searchInputRef={searchInputRef}>
      <div className={filterStyles.modalFilters}>
        <FilterInput
          ref={searchInputRef}
          placeholder="Filter words..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
      </div>

      <div className={styles.modalContent}>
        <div className={styles.section}>
          <div className={styles.sectionHeader}>
            <span className={styles.sectionTitle}>Added Words</span>
            <span className={`${styles.sectionCount} ${styles.sectionCountAdded}`}>
              +{filteredAdded.length}
            </span>
          </div>
          {filteredAdded.length > 0 ? (
            <div className={styles.wordList}>
              {filteredAdded.map((word, index) => (
                <span key={index} className={`${styles.wordChip} ${styles.wordChipAdded}`}>
                  {word}
                </span>
              ))}
            </div>
          ) : (
            <p className={styles.emptyMessage}>
              {filter ? 'No matching words' : 'No words added'}
            </p>
          )}
        </div>

        <div ref={removedSectionRef} className={styles.section}>
          <div className={styles.sectionHeader}>
            <span className={styles.sectionTitle}>Removed Words</span>
            <span className={`${styles.sectionCount} ${styles.sectionCountRemoved}`}>
              -{filteredRemoved.length}
            </span>
          </div>
          {filteredRemoved.length > 0 ? (
            <div className={styles.wordList}>
              {filteredRemoved.map((word, index) => (
                <span key={index} className={`${styles.wordChip} ${styles.wordChipRemoved}`}>
                  {word}
                </span>
              ))}
            </div>
          ) : (
            <p className={styles.emptyMessage}>
              {filter ? 'No matching words' : 'No words removed'}
            </p>
          )}
        </div>
      </div>
    </Modal>
  );
}
