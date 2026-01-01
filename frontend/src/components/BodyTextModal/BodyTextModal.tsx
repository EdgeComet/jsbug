import { useState, useEffect, useMemo, useRef, useCallback } from 'react';
import { computeBlockDiff } from '../../utils/blockDiff';
import { Modal } from '../common/Modal';
import { FilterInput } from '../common/FilterInput';
import { MatchCountBadge } from '../common/MatchCountBadge/MatchCountBadge';
import { SearchNavigation } from '../common/SearchNavigation/SearchNavigation';
import { MarkdownContent } from './MarkdownContent';
import { DiffMarkdownContent } from './DiffMarkdownContent';
import filterStyles from '../common/Modal/ModalFilters.module.css';
import highlightStyles from '../../styles/highlight.module.css';
import styles from './BodyTextModal.module.css';

interface BodyTextModalProps {
  isOpen: boolean;
  onClose: () => void;
  bodyMarkdown: string;
  wordCount: number;
  compareBodyMarkdown?: string;  // other panel's body markdown
  isLoading?: boolean;
  defaultCompareMode?: boolean;
  side?: 'left' | 'right';       // which panel opened the modal
}

export function BodyTextModal({ isOpen, onClose, bodyMarkdown, wordCount, compareBodyMarkdown, isLoading = false, defaultCompareMode = true, side = 'left' }: BodyTextModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [currentMatchIndex, setCurrentMatchIndex] = useState(0);
  const [compareMode, setCompareMode] = useState(defaultCompareMode);
  const [currentChangeIndex, setCurrentChangeIndex] = useState(0);
  const bodyRef = useRef<HTMLDivElement>(null);
  const leftPanelRef = useRef<HTMLDivElement>(null);
  const rightPanelRef = useRef<HTMLDivElement>(null);
  const isScrolling = useRef(false);


  // Improved scroll sync with block alignment
  const handleScroll = useCallback((source: 'left' | 'right') => {
    if (isScrolling.current) return;
    isScrolling.current = true;

    const sourceRef = source === 'left' ? leftPanelRef : rightPanelRef;
    const targetRef = source === 'left' ? rightPanelRef : leftPanelRef;

    if (!sourceRef.current || !targetRef.current) {
      requestAnimationFrame(() => {
        isScrolling.current = false;
      });
      return;
    }

    const sourceHeight = sourceRef.current.scrollHeight;
    const targetHeight = targetRef.current.scrollHeight;
    const heightDiff = Math.abs(sourceHeight - targetHeight);
    const maxHeight = Math.max(sourceHeight, targetHeight);

    // Use 1:1 scroll sync when heights are similar (within 5% or 100px)
    const heightsAreSimilar = heightDiff < 100 || (heightDiff / maxHeight) < 0.05;

    if (heightsAreSimilar) {
      targetRef.current.scrollTop = sourceRef.current.scrollTop;
      requestAnimationFrame(() => {
        isScrolling.current = false;
      });
      return;
    }

    // For significantly different heights, use percentage-based sync
    const scrollPercentage = sourceRef.current.scrollTop /
      Math.max(1, sourceRef.current.scrollHeight - sourceRef.current.clientHeight);
    targetRef.current.scrollTop = scrollPercentage *
      (targetRef.current.scrollHeight - targetRef.current.clientHeight);

    requestAnimationFrame(() => {
      isScrolling.current = false;
    });
  }, []);

  const hasCompareData = !!compareBodyMarkdown;

  useEffect(() => {
    if (isOpen) {
      setSearchTerm('');
      setCurrentMatchIndex(0);
      setCompareMode(defaultCompareMode);
      // Focus search input when modal opens
      requestAnimationFrame(() => {
        searchInputRef.current?.focus();
      });
    }
  }, [isOpen, defaultCompareMode]);

  useEffect(() => {
    setCurrentMatchIndex(0);
  }, [searchTerm]);

  // Swap content based on which panel opened the modal
  // Left column should always show left panel content, right column shows right panel content
  const leftColumnMarkdown = side === 'left' ? bodyMarkdown : (compareBodyMarkdown || '');
  const rightColumnMarkdown = side === 'left' ? (compareBodyMarkdown || '') : bodyMarkdown;

  // Compute changed block indices for diff navigation (uses swapped content)
  const changedBlockIndices = useMemo(() => {
    if (!leftColumnMarkdown || !rightColumnMarkdown) return [];
    const { leftBlocks } = computeBlockDiff(leftColumnMarkdown, rightColumnMarkdown);
    return leftBlocks
      .map((block, i) => block.type !== 'unchanged' ? i : -1)
      .filter(i => i !== -1);
  }, [leftColumnMarkdown, rightColumnMarkdown]);

  const changeCount = changedBlockIndices.length;

  // Content size calculations
  const contentLength = bodyMarkdown.length;
  const isLargeContent = contentLength > 100000; // 100KB
  const isContentEmpty = !bodyMarkdown.trim();

  const matchCount = useMemo(() => {
    if (!searchTerm.trim()) return 0;
    const textToSearch = compareMode ? leftColumnMarkdown : bodyMarkdown;
    const regex = new RegExp(searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
    const matches = textToSearch.match(regex);
    return matches ? matches.length : 0;
  }, [bodyMarkdown, searchTerm, compareMode, leftColumnMarkdown]);

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

  const goToNextChange = useCallback(() => {
    if (changeCount > 0) {
      setCurrentChangeIndex((prev) => (prev + 1) % changeCount);
    }
  }, [changeCount]);

  const goToPrevChange = useCallback(() => {
    if (changeCount > 0) {
      setCurrentChangeIndex((prev) => (prev - 1 + changeCount) % changeCount);
    }
  }, [changeCount]);

  // Scroll to current change and highlight it
  useEffect(() => {
    if (!compareMode || changeCount === 0) return;

    const blockIndex = changedBlockIndices[currentChangeIndex];

    // Remove highlight from all blocks in both panels
    [leftPanelRef, rightPanelRef].forEach(ref => {
      ref.current?.querySelectorAll(`.${styles.activeChange}`).forEach(el => {
        el.classList.remove(styles.activeChange);
      });
    });

    // Add highlight and scroll to current block in both panels
    [leftPanelRef, rightPanelRef].forEach(ref => {
      const blockElement = ref.current?.querySelector(
        `[data-block-index="${blockIndex}"]`
      );
      if (blockElement) {
        blockElement.classList.add(styles.activeChange);
      }
    });

    // Scroll left panel (right will follow via scroll sync)
    const leftBlock = leftPanelRef.current?.querySelector(
      `[data-block-index="${blockIndex}"]`
    );
    if (leftBlock) {
      leftBlock.scrollIntoView({ behavior: 'instant', block: 'center' });
    }
  }, [currentChangeIndex, compareMode, changeCount, changedBlockIndices]);

  // Reset change index when content changes
  useEffect(() => {
    setCurrentChangeIndex(0);
  }, [bodyMarkdown, compareBodyMarkdown]);

  useEffect(() => {
    if (!searchTerm.trim()) return;

    // In compare mode, use left panel; otherwise use bodyRef
    const container = compareMode ? leftPanelRef.current : bodyRef.current;
    if (!container) return;

    const matches = container.querySelectorAll('mark');
    matches.forEach((m, i) => {
      m.classList.toggle(highlightStyles.highlightActive, i === currentMatchIndex);
    });
    if (matches[currentMatchIndex]) {
      matches[currentMatchIndex].scrollIntoView({ behavior: 'instant', block: 'center' });
    }
  }, [searchTerm, currentMatchIndex, compareMode]);

  const headerExtra = (
    <>
      {searchTerm && matchCount > 0 && <MatchCountBadge count={matchCount} />}
      {hasCompareData && (
        <button
          className={`${styles.toggleButton} ${compareMode ? styles.active : ''}`}
          onClick={() => setCompareMode(!compareMode)}
          aria-pressed={compareMode}
        >
          Compare {compareMode ? 'ON' : 'OFF'}
        </button>
      )}
      {compareMode && changeCount > 0 && (
        <div className={styles.changeNavigation}>
          <button
            className={styles.changeNavButton}
            onClick={goToPrevChange}
            aria-label="Previous change"
            title="Previous change"
          >
            &lt;
          </button>
          <span className={styles.changeCount}>
            {currentChangeIndex + 1} / {changeCount}
          </span>
          <button
            className={styles.changeNavButton}
            onClick={goToNextChange}
            aria-label="Next change"
            title="Next change"
          >
            &gt;
          </button>
        </div>
      )}
    </>
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Body Text (${wordCount.toLocaleString()} words)`}
      size={compareMode ? "wide" : "lg"}
      headerExtra={headerExtra}
      searchInputRef={searchInputRef}
    >
      <div className={filterStyles.modalFilters}>
        <div role="search" aria-label="Search within content" className={styles.searchWrapper}>
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
            aria-label="Search text content"
            data-search-input
          />
        </div>
        <SearchNavigation
          currentIndex={currentMatchIndex}
          totalMatches={matchCount}
          onPrevious={goToPrevMatch}
          onNext={goToNextMatch}
        />
        <div className={styles.keyboardHints} aria-hidden="true">
          <span><kbd>Enter</kbd> Next match</span>
          <span><kbd>Shift+Enter</kbd> Previous match</span>
          <span><kbd>Esc</kbd> Close</span>
        </div>
      </div>

      {isLoading ? (
        <div className={styles.loadingState}>
          <div className={styles.spinner} />
          <span>Loading content...</span>
        </div>
      ) : isContentEmpty ? (
        <div className={styles.emptyState}>
          No text content found on this page.
        </div>
      ) : compareMode ? (
        <div
          className={styles.compareContainer}
          role="region"
          aria-label="Content comparison"
        >
          {isLargeContent && (
            <div className={styles.sizeWarning}>
              Large document ({Math.round(contentLength / 1024)}KB). Scrolling may be slow.
            </div>
          )}
          <div className={styles.compareColumn}>
            <div className={styles.columnHeader}>Left Panel</div>
            <div
              ref={leftPanelRef}
              className={styles.columnContent}
              onScroll={() => handleScroll('left')}
              tabIndex={0}
              aria-label="Left panel content"
            >
              {leftColumnMarkdown && rightColumnMarkdown ? (
                <DiffMarkdownContent
                  leftContent={leftColumnMarkdown}
                  rightContent={rightColumnMarkdown}
                  searchTerm={searchTerm}
                  side="left"
                />
              ) : (
                <MarkdownContent
                  content={leftColumnMarkdown}
                  searchTerm={searchTerm}
                  activeMatchIndex={currentMatchIndex}
                />
              )}
            </div>
          </div>
          <div className={styles.compareDivider} aria-hidden="true" />
          <div className={styles.compareColumn}>
            <div className={styles.columnHeader}>Right Panel</div>
            <div
              ref={rightPanelRef}
              className={styles.columnContent}
              onScroll={() => handleScroll('right')}
              tabIndex={0}
              aria-label="Right panel content"
            >
              {leftColumnMarkdown && rightColumnMarkdown ? (
                <DiffMarkdownContent
                  leftContent={leftColumnMarkdown}
                  rightContent={rightColumnMarkdown}
                  searchTerm={searchTerm}
                  side="right"
                />
              ) : (
                <MarkdownContent
                  content={rightColumnMarkdown}
                  searchTerm={searchTerm}
                />
              )}
            </div>
          </div>
        </div>
      ) : (
        <div
          ref={bodyRef}
          className={styles.modalBody}
          role="region"
          aria-label="Page content"
          tabIndex={0}
        >
          {isLargeContent && (
            <div className={styles.sizeWarning}>
              Large document ({Math.round(contentLength / 1024)}KB). Scrolling may be slow.
            </div>
          )}
          <MarkdownContent
            content={bodyMarkdown}
            searchTerm={searchTerm}
            activeMatchIndex={currentMatchIndex}
          />
        </div>
      )}
    </Modal>
  );
}
