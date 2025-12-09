import { useState, useEffect, useMemo, useRef } from 'react';
import type { Link } from '../../types/content';
import { UrlLink } from '../common/UrlLink';
import { Modal } from '../common/Modal';
import { FilterInput } from '../common/FilterInput';
import filterStyles from '../common/Modal/ModalFilters.module.css';
import tableStyles from '../../styles/dataTable.module.css';
import styles from './LinksModal.module.css';

export type LinkFilterType = 'all' | 'internal' | 'external';
export type LinkDiffType = 'all' | 'added' | 'removed';

interface LinksModalProps {
  isOpen: boolean;
  onClose: () => void;
  links: Link[];
  compareLinks?: Link[];
  initialFilter: LinkFilterType;
  initialDiffType?: LinkDiffType;
}

export function LinksModal({ isOpen, onClose, links, compareLinks, initialFilter, initialDiffType = 'all' }: LinksModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [hrefFilter, setHrefFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState<LinkFilterType>(initialFilter);
  const [diffFilter, setDiffFilter] = useState<LinkDiffType>(initialDiffType);

  // Compute added/removed links
  const { addedLinks, removedLinks, addedHrefs, removedHrefs } = useMemo(() => {
    if (!compareLinks) {
      return { addedLinks: [], removedLinks: [], addedHrefs: new Set<string>(), removedHrefs: new Set<string>() };
    }
    const currentHrefs = new Set(links.map(l => l.href));
    const compareHrefs = new Set(compareLinks.map(l => l.href));

    const added = links.filter(l => !compareHrefs.has(l.href));
    const removed = compareLinks.filter(l => !currentHrefs.has(l.href));

    return {
      addedLinks: added,
      removedLinks: removed,
      addedHrefs: new Set(added.map(l => l.href)),
      removedHrefs: new Set(removed.map(l => l.href)),
    };
  }, [links, compareLinks]);

  const hasChanges = compareLinks && (addedLinks.length > 0 || removedLinks.length > 0);

  useEffect(() => {
    if (isOpen) {
      setHrefFilter('');
      setTypeFilter(initialFilter);
      setDiffFilter(initialDiffType);
    }
  }, [isOpen, initialFilter, initialDiffType]);

  // Build the list of links to display based on diff filter
  const displayLinks = useMemo(() => {
    if (diffFilter === 'added') {
      return addedLinks;
    } else if (diffFilter === 'removed') {
      return removedLinks;
    }
    if (compareLinks) {
      const allLinks = [...links];
      removedLinks.forEach(link => {
        if (!links.some(l => l.href === link.href)) {
          allLinks.push(link);
        }
      });
      return allLinks;
    }
    return links;
  }, [links, compareLinks, addedLinks, removedLinks, diffFilter]);

  const filteredLinks = useMemo(() => {
    return displayLinks.filter(link => {
      if (typeFilter === 'internal' && link.isExternal) return false;
      if (typeFilter === 'external' && !link.isExternal) return false;
      if (hrefFilter) {
        const searchTerm = hrefFilter.toLowerCase();
        const matchesHref = link.href.toLowerCase().includes(searchTerm);
        const matchesText = link.text.toLowerCase().includes(searchTerm);
        if (!matchesHref && !matchesText) {
          return false;
        }
      }
      return true;
    });
  }, [displayLinks, typeFilter, hrefFilter]);

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Links (${filteredLinks.length} of ${links.length})`}
      size="lg"
      searchInputRef={searchInputRef}
    >
      <div className={filterStyles.modalFilters}>
        <FilterInput
          ref={searchInputRef}
          placeholder="Filter by href or text..."
          value={hrefFilter}
          onChange={(e) => setHrefFilter(e.target.value)}
        />
        <select
          className={filterStyles.filterSelect}
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value as LinkFilterType)}
        >
          <option value="all">All Links</option>
          <option value="internal">Internal Only</option>
          <option value="external">External Only</option>
        </select>
        {hasChanges && (
          <select
            className={filterStyles.filterSelect}
            value={diffFilter}
            onChange={(e) => setDiffFilter(e.target.value as LinkDiffType)}
          >
            <option value="all">All</option>
            <option value="added">Added (+{addedLinks.length})</option>
            <option value="removed">Removed (-{removedLinks.length})</option>
          </select>
        )}
      </div>

      <div className={styles.modalBody}>
        <table className={tableStyles.dataTable}>
          <thead>
            <tr>
              <th className={styles.hrefCol}>Href</th>
              <th className={styles.textCol}>Text</th>
            </tr>
          </thead>
          <tbody>
            {filteredLinks.map((link, index) => {
              const isAdded = addedHrefs.has(link.href);
              const isRemoved = removedHrefs.has(link.href);
              const rowClass = isAdded ? tableStyles.rowAdded : isRemoved ? tableStyles.rowRemoved : '';
              return (
                <tr key={index} className={rowClass}>
                  <td className={`${styles.hrefCol} ${tableStyles.cellBreakAll}`}>
                    <UrlLink url={link.href} />
                  </td>
                  <td className={`${styles.textCol} ${tableStyles.cellBreakWord}`}>
                    {link.isImageLink ? <span className={styles.imageTag}>(image)</span> : link.text || <span className={tableStyles.cellMuted}>(empty)</span>}
                  </td>
                </tr>
              );
            })}
            {filteredLinks.length === 0 && (
              <tr>
                <td colSpan={2} className={tableStyles.emptyMessage}>
                  No links match your filters
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </Modal>
  );
}
