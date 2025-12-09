import { useState, useEffect, useMemo, useRef } from 'react';
import type { Image } from '../../types/content';
import { UrlLink } from '../common/UrlLink';
import { Modal } from '../common/Modal';
import { FilterInput } from '../common/FilterInput';
import filterStyles from '../common/Modal/ModalFilters.module.css';
import tableStyles from '../../styles/dataTable.module.css';
import styles from './ImagesModal.module.css';

export type ImageFilterType = 'all' | 'internal' | 'external';
export type ImageDiffType = 'all' | 'added' | 'removed';

interface ImagesModalProps {
  isOpen: boolean;
  onClose: () => void;
  images: Image[];
  compareImages?: Image[];
  initialFilter: ImageFilterType;
  initialDiffType?: ImageDiffType;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '-';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function ImagesModal({ isOpen, onClose, images, compareImages, initialFilter, initialDiffType = 'all' }: ImagesModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [srcFilter, setSrcFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState<ImageFilterType>(initialFilter);
  const [diffFilter, setDiffFilter] = useState<ImageDiffType>(initialDiffType);

  // Compute added/removed images
  const { addedImages, removedImages, addedSrcs, removedSrcs } = useMemo(() => {
    if (!compareImages) {
      return { addedImages: [], removedImages: [], addedSrcs: new Set<string>(), removedSrcs: new Set<string>() };
    }
    const currentSrcs = new Set(images.map(img => img.src));
    const compareSrcs = new Set(compareImages.map(img => img.src));

    const added = images.filter(img => !compareSrcs.has(img.src));
    const removed = compareImages.filter(img => !currentSrcs.has(img.src));

    return {
      addedImages: added,
      removedImages: removed,
      addedSrcs: new Set(added.map(img => img.src)),
      removedSrcs: new Set(removed.map(img => img.src)),
    };
  }, [images, compareImages]);

  const hasChanges = compareImages && (addedImages.length > 0 || removedImages.length > 0);

  useEffect(() => {
    if (isOpen) {
      setSrcFilter('');
      setTypeFilter(initialFilter);
      setDiffFilter(initialDiffType);
    }
  }, [isOpen, initialFilter, initialDiffType]);

  // Build the list of images to display based on diff filter
  const displayImages = useMemo(() => {
    if (diffFilter === 'added') {
      return addedImages;
    } else if (diffFilter === 'removed') {
      return removedImages;
    }
    if (compareImages) {
      const allImages = [...images];
      removedImages.forEach(img => {
        if (!images.some(i => i.src === img.src)) {
          allImages.push(img);
        }
      });
      return allImages;
    }
    return images;
  }, [images, compareImages, addedImages, removedImages, diffFilter]);

  const filteredImages = useMemo(() => {
    return displayImages.filter(img => {
      if (typeFilter === 'internal' && img.isExternal) return false;
      if (typeFilter === 'external' && !img.isExternal) return false;
      if (srcFilter && !img.src.toLowerCase().includes(srcFilter.toLowerCase())) {
        return false;
      }
      return true;
    });
  }, [displayImages, typeFilter, srcFilter]);

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Images (${filteredImages.length} of ${images.length})`}
      size="xl"
      searchInputRef={searchInputRef}
    >
      <div className={filterStyles.modalFilters}>
        <FilterInput
          ref={searchInputRef}
          placeholder="Filter by src..."
          value={srcFilter}
          onChange={(e) => setSrcFilter(e.target.value)}
        />
        <select
          className={filterStyles.filterSelect}
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value as ImageFilterType)}
        >
          <option value="all">All Images</option>
          <option value="internal">Internal Only</option>
          <option value="external">External Only</option>
        </select>
        {hasChanges && (
          <select
            className={filterStyles.filterSelect}
            value={diffFilter}
            onChange={(e) => setDiffFilter(e.target.value as ImageDiffType)}
          >
            <option value="all">All</option>
            <option value="added">Added (+{addedImages.length})</option>
            <option value="removed">Removed (-{removedImages.length})</option>
          </select>
        )}
      </div>

      <div className={styles.modalBody}>
        <table className={tableStyles.dataTable}>
          <thead>
            <tr>
              <th className={styles.srcCol}>Src</th>
              <th className={styles.altCol}>Alt</th>
              <th className={styles.sizeCol}>Size</th>
            </tr>
          </thead>
          <tbody>
            {filteredImages.map((img, index) => {
              const isAdded = addedSrcs.has(img.src);
              const isRemoved = removedSrcs.has(img.src);
              const rowClass = isAdded ? tableStyles.rowAdded : isRemoved ? tableStyles.rowRemoved : '';
              return (
                <tr key={index} className={rowClass}>
                  <td className={`${styles.srcCol} ${tableStyles.cellBreakAll}`}>
                    <UrlLink url={img.src} />
                  </td>
                  <td className={`${styles.altCol} ${tableStyles.cellBreakWord}`}>
                    {img.isInLink && <span className={styles.inLinkTag}>(in link)</span>}
                    {img.alt || <span className={tableStyles.cellMuted}>(empty)</span>}
                  </td>
                  <td className={styles.sizeCol}>
                    {formatBytes(img.size)}
                  </td>
                </tr>
              );
            })}
            {filteredImages.length === 0 && (
              <tr>
                <td colSpan={3} className={tableStyles.emptyMessage}>
                  No images match your filters
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </Modal>
  );
}
