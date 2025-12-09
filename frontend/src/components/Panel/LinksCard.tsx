import { useMemo } from 'react';
import type { LinksData, Link } from '../../types/content';
import type { LinkFilterType } from '../LinksModal/LinksModal';
import styles from './Panel.module.css';

interface LinksDiff {
  added: Link[];
  removed: Link[];
}

function computeLinksDiff(links: Link[], compareLinks: Link[]): LinksDiff {
  const currentHrefs = new Set(links.map(l => l.href));
  const compareHrefs = new Set(compareLinks.map(l => l.href));

  const added = links.filter(l => !compareHrefs.has(l.href));
  const removed = compareLinks.filter(l => !currentHrefs.has(l.href));

  return { added, removed };
}

interface LinksCardProps {
  data: LinksData;
  compareData?: LinksData;
  onOpenModal?: (filter: LinkFilterType) => void;
  onOpenDiffModal?: (filter: LinkFilterType, diffType: 'added' | 'removed') => void;
}

export function LinksCard({ data, compareData, onOpenModal, onOpenDiffModal }: LinksCardProps) {
  const { total, internal, external } = useMemo(() => {
    const total = data.links.length;
    const external = data.links.filter(link => link.isExternal).length;
    const internal = total - external;
    return { total, internal, external };
  }, [data.links]);

  const diff = useMemo(() => {
    if (!compareData) return null;
    return computeLinksDiff(data.links, compareData.links);
  }, [data.links, compareData]);

  const diffCounts = useMemo(() => {
    if (!diff) return null;

    const addedInternal = diff.added.filter(l => !l.isExternal).length;
    const addedExternal = diff.added.filter(l => l.isExternal).length;
    const removedInternal = diff.removed.filter(l => !l.isExternal).length;
    const removedExternal = diff.removed.filter(l => l.isExternal).length;

    return {
      totalAdded: diff.added.length,
      totalRemoved: diff.removed.length,
      internalAdded: addedInternal,
      internalRemoved: removedInternal,
      externalAdded: addedExternal,
      externalRemoved: removedExternal,
    };
  }, [diff]);

  const hasDiff = diffCounts && (diffCounts.totalAdded > 0 || diffCounts.totalRemoved > 0);

  return (
    <div className={styles.resultCard}>
      <div className={styles.resultCardHeader}>
        <span className={styles.resultCardTitle}>Links</span>
      </div>
      <div className={`${styles.linksRow} ${hasDiff ? styles.diffHighlightChanged : ''}`}>
        <div className={styles.linksStatWrapper}>
          <button
            className={styles.linksStatClickable}
            onClick={() => onOpenModal?.('all')}
            type="button"
          >
            <span className={styles.linksStatValue}>{total}</span>
            <span className={styles.linksStatLabel}>Total</span>
          </button>
          {hasDiff && (diffCounts.totalAdded > 0 || diffCounts.totalRemoved > 0) && (
            <div className={styles.linksDiffNumbers}>
              {diffCounts.totalAdded > 0 && (
                <button
                  type="button"
                  className={styles.linksDiffAdded}
                  onClick={() => onOpenDiffModal?.('all', 'added')}
                >
                  +{diffCounts.totalAdded}
                </button>
              )}
              {diffCounts.totalRemoved > 0 && (
                <button
                  type="button"
                  className={styles.linksDiffRemoved}
                  onClick={() => onOpenDiffModal?.('all', 'removed')}
                >
                  -{diffCounts.totalRemoved}
                </button>
              )}
            </div>
          )}
        </div>
        <div className={styles.linksStatWrapper}>
          <button
            className={styles.linksStatClickable}
            onClick={() => onOpenModal?.('internal')}
            type="button"
          >
            <span className={styles.linksStatValue}>{internal}</span>
            <span className={styles.linksStatLabel}>Internal</span>
          </button>
          {hasDiff && (diffCounts.internalAdded > 0 || diffCounts.internalRemoved > 0) && (
            <div className={styles.linksDiffNumbers}>
              {diffCounts.internalAdded > 0 && (
                <button
                  type="button"
                  className={styles.linksDiffAdded}
                  onClick={() => onOpenDiffModal?.('internal', 'added')}
                >
                  +{diffCounts.internalAdded}
                </button>
              )}
              {diffCounts.internalRemoved > 0 && (
                <button
                  type="button"
                  className={styles.linksDiffRemoved}
                  onClick={() => onOpenDiffModal?.('internal', 'removed')}
                >
                  -{diffCounts.internalRemoved}
                </button>
              )}
            </div>
          )}
        </div>
        <div className={styles.linksStatWrapper}>
          <button
            className={styles.linksStatClickable}
            onClick={() => onOpenModal?.('external')}
            type="button"
          >
            <span className={styles.linksStatValue}>{external}</span>
            <span className={styles.linksStatLabel}>External</span>
          </button>
          {hasDiff && (diffCounts.externalAdded > 0 || diffCounts.externalRemoved > 0) && (
            <div className={styles.linksDiffNumbers}>
              {diffCounts.externalAdded > 0 && (
                <button
                  type="button"
                  className={styles.linksDiffAdded}
                  onClick={() => onOpenDiffModal?.('external', 'added')}
                >
                  +{diffCounts.externalAdded}
                </button>
              )}
              {diffCounts.externalRemoved > 0 && (
                <button
                  type="button"
                  className={styles.linksDiffRemoved}
                  onClick={() => onOpenDiffModal?.('external', 'removed')}
                >
                  -{diffCounts.externalRemoved}
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
