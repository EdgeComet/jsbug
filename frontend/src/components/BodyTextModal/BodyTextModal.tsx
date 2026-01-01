import { useState, useEffect, useMemo, useRef, useCallback } from 'react';
import { highlightText } from '../../utils/highlightText';
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
  bodyText: string;
  bodyMarkdown?: string;
  wordCount: number;
  compareBodyText?: string;      // no-JS body text
  compareBodyMarkdown?: string;  // no-JS body markdown
  isLoading?: boolean;
}

interface BlockPosition {
  index: number;
  top: number;
  height: number;
}

export function BodyTextModal({ isOpen, onClose, bodyText, bodyMarkdown, wordCount, compareBodyText, compareBodyMarkdown, isLoading = false }: BodyTextModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [currentMatchIndex, setCurrentMatchIndex] = useState(0);
  const [compareMode, setCompareMode] = useState(false);
  const bodyRef = useRef<HTMLDivElement>(null);
  const leftPanelRef = useRef<HTMLDivElement>(null);
  const rightPanelRef = useRef<HTMLDivElement>(null);
  const isScrolling = useRef(false);
  const [leftBlockPositions, setLeftBlockPositions] = useState<BlockPosition[]>([]);
  const [rightBlockPositions, setRightBlockPositions] = useState<BlockPosition[]>([]);

  // Measure block positions after render
  useEffect(() => {
    if (!compareMode) return;

    const measureBlocks = (container: HTMLElement | null): BlockPosition[] => {
      if (!container) return [];
      const blocks = container.querySelectorAll('[data-block-index]');
      return Array.from(blocks).map((block, i) => ({
        index: i,
        top: (block as HTMLElement).offsetTop,
        height: (block as HTMLElement).offsetHeight,
      }));
    };

    // Use setTimeout to ensure DOM is updated
    const timer = setTimeout(() => {
      setLeftBlockPositions(measureBlocks(leftPanelRef.current));
      setRightBlockPositions(measureBlocks(rightPanelRef.current));
    }, 100);

    return () => clearTimeout(timer);
  }, [compareMode, bodyMarkdown, compareBodyMarkdown]);

  // Improved scroll sync with block alignment
  const handleScroll = useCallback((source: 'left' | 'right') => {
    if (isScrolling.current) return;
    isScrolling.current = true;

    const sourceRef = source === 'left' ? leftPanelRef : rightPanelRef;
    const targetRef = source === 'left' ? rightPanelRef : leftPanelRef;
    const sourcePositions = source === 'left' ? leftBlockPositions : rightBlockPositions;
    const targetPositions = source === 'left' ? rightBlockPositions : leftBlockPositions;

    if (!sourceRef.current || !targetRef.current || sourcePositions.length === 0) {
      // Fall back to percentage-based sync
      if (sourceRef.current && targetRef.current) {
        const scrollPercentage = sourceRef.current.scrollTop /
          Math.max(1, sourceRef.current.scrollHeight - sourceRef.current.clientHeight);
        targetRef.current.scrollTop = scrollPercentage *
          (targetRef.current.scrollHeight - targetRef.current.clientHeight);
      }
      requestAnimationFrame(() => {
        isScrolling.current = false;
      });
      return;
    }

    const scrollTop = sourceRef.current.scrollTop;
    const viewportCenter = scrollTop + sourceRef.current.clientHeight / 2;

    // Find which block is in the center of viewport
    let activeBlockIndex = 0;
    for (let i = 0; i < sourcePositions.length; i++) {
      if (sourcePositions[i].top + sourcePositions[i].height / 2 > viewportCenter) {
        break;
      }
      activeBlockIndex = i;
    }

    // Scroll target to align matching block
    if (targetPositions[activeBlockIndex]) {
      const targetScrollTop = targetPositions[activeBlockIndex].top -
        sourceRef.current.clientHeight / 2 +
        targetPositions[activeBlockIndex].height / 2;

      targetRef.current.scrollTop = Math.max(0, targetScrollTop);
    }

    requestAnimationFrame(() => {
      isScrolling.current = false;
    });
  }, [leftBlockPositions, rightBlockPositions]);

  const hasCompareData = !!compareBodyMarkdown || !!compareBodyText;

  useEffect(() => {
    if (isOpen) {
      setSearchTerm('');
      setCurrentMatchIndex(0);
      setCompareMode(false);
      // Focus search input when modal opens
      requestAnimationFrame(() => {
        searchInputRef.current?.focus();
      });
    }
  }, [isOpen]);

  useEffect(() => {
    setCurrentMatchIndex(0);
  }, [searchTerm]);

  const hasMarkdown = bodyMarkdown && bodyMarkdown.trim().length > 0;
  const hasCompareMarkdown = compareBodyMarkdown && compareBodyMarkdown.trim().length > 0;

  // Content size calculations
  const contentLength = (bodyMarkdown || bodyText).length;
  const isLargeContent = contentLength > 100000; // 100KB
  const isContentEmpty = !bodyText && !bodyMarkdown;

  const matchCount = useMemo(() => {
    if (!searchTerm.trim()) return 0;
    const textToSearch = hasMarkdown ? bodyMarkdown : bodyText;
    const regex = new RegExp(searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
    const matches = textToSearch.match(regex);
    return matches ? matches.length : 0;
  }, [bodyText, bodyMarkdown, hasMarkdown, searchTerm]);

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
      size={compareMode ? "wide" : "lg"}
      headerExtra={headerExtra}
      searchInputRef={searchInputRef}
    >
      <div className={filterStyles.modalFilters}>
        <div role="search" aria-label="Search within content">
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
        {hasCompareData && (
          <button
            className={`${styles.toggleButton} ${compareMode ? styles.active : ''}`}
            onClick={() => setCompareMode(!compareMode)}
            aria-pressed={compareMode}
          >
            Compare {compareMode ? 'ON' : 'OFF'}
          </button>
        )}
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
            <div className={styles.columnHeader}>JS Rendered</div>
            <div
              ref={leftPanelRef}
              className={styles.columnContent}
              onScroll={() => handleScroll('left')}
              tabIndex={0}
              aria-label="JS Rendered content"
            >
              {!hasMarkdown && bodyText && (
                <div className={styles.fallbackNotice}>
                  Structured view unavailable. Showing plain text.
                </div>
              )}
              {hasMarkdown && hasCompareMarkdown ? (
                <DiffMarkdownContent
                  leftContent={bodyMarkdown || ''}
                  rightContent={compareBodyMarkdown || ''}
                  searchTerm={searchTerm}
                  side="left"
                />
              ) : hasMarkdown ? (
                <MarkdownContent
                  content={bodyMarkdown || ''}
                  searchTerm={searchTerm}
                  activeMatchIndex={currentMatchIndex}
                />
              ) : (
                <div className={styles.textContent}>
                  {highlightText(bodyText, searchTerm, highlightStyles.highlight)}
                </div>
              )}
            </div>
          </div>
          <div className={styles.compareDivider} aria-hidden="true" />
          <div className={styles.compareColumn}>
            <div className={styles.columnHeader}>No JS</div>
            <div
              ref={rightPanelRef}
              className={styles.columnContent}
              onScroll={() => handleScroll('right')}
              tabIndex={0}
              aria-label="No JS content"
            >
              {!hasCompareMarkdown && compareBodyText && (
                <div className={styles.fallbackNotice}>
                  Structured view unavailable. Showing plain text.
                </div>
              )}
              {hasMarkdown && hasCompareMarkdown ? (
                <DiffMarkdownContent
                  leftContent={bodyMarkdown || ''}
                  rightContent={compareBodyMarkdown || ''}
                  searchTerm={searchTerm}
                  side="right"
                />
              ) : hasCompareMarkdown ? (
                <MarkdownContent
                  content={compareBodyMarkdown}
                  searchTerm={searchTerm}
                />
              ) : (
                <div className={styles.textContent}>
                  {highlightText(compareBodyText || '', searchTerm, highlightStyles.highlight)}
                </div>
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
          {!hasMarkdown && bodyText && (
            <div className={styles.fallbackNotice}>
              Structured view unavailable. Showing plain text.
            </div>
          )}
          {hasMarkdown ? (
            <MarkdownContent
              content={bodyMarkdown || ''}
              searchTerm={searchTerm}
              activeMatchIndex={currentMatchIndex}
            />
          ) : (
            <div className={styles.textContent}>
              {highlightText(bodyText, searchTerm, highlightStyles.highlight)}
            </div>
          )}
        </div>
      )}
    </Modal>
  );
}
