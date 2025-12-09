import { useMemo } from 'react';
import { Icon } from '../common/Icon';
import type { ImagesData, Image } from '../../types/content';
import type { ImageFilterType } from '../ImagesModal/ImagesModal';
import styles from './Panel.module.css';

interface ImagesDiff {
  added: Image[];
  removed: Image[];
}

function computeImagesDiff(images: Image[], compareImages: Image[]): ImagesDiff {
  const currentSrcs = new Set(images.map(img => img.src));
  const compareSrcs = new Set(compareImages.map(img => img.src));

  const added = images.filter(img => !compareSrcs.has(img.src));
  const removed = compareImages.filter(img => !currentSrcs.has(img.src));

  return { added, removed };
}

interface ImagesCardProps {
  data: ImagesData;
  compareData?: ImagesData;
  onOpenModal?: (filter: ImageFilterType) => void;
  onOpenDiffModal?: (filter: ImageFilterType, diffType: 'added' | 'removed') => void;
}

export function ImagesCard({ data, compareData, onOpenModal, onOpenDiffModal }: ImagesCardProps) {
  const { total, internal, external } = useMemo(() => {
    const total = data.images.length;
    const external = data.images.filter(img => img.isExternal).length;
    const internal = total - external;
    return { total, internal, external };
  }, [data.images]);

  const diff = useMemo(() => {
    if (!compareData) return null;
    return computeImagesDiff(data.images, compareData.images);
  }, [data.images, compareData]);

  const diffCounts = useMemo(() => {
    if (!diff) return null;

    const addedInternal = diff.added.filter(img => !img.isExternal).length;
    const addedExternal = diff.added.filter(img => img.isExternal).length;
    const removedInternal = diff.removed.filter(img => !img.isExternal).length;
    const removedExternal = diff.removed.filter(img => img.isExternal).length;

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
        <span className={styles.resultCardTitle}>
          <Icon name="image" size={14} />
          Images
        </span>
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
