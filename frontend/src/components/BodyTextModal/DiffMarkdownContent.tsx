import React from 'react';
import { computeBlockDiff, DiffBlock } from '../../utils/blockDiff';
import { MarkdownContent } from './MarkdownContent';
import { InlineDiff } from './InlineDiff';
import styles from './DiffMarkdownContent.module.css';

interface DiffMarkdownContentProps {
  leftContent: string;
  rightContent: string;
  searchTerm?: string;
  side: 'left' | 'right';
}

export const DiffMarkdownContent: React.FC<DiffMarkdownContentProps> = ({
  leftContent,
  rightContent,
  searchTerm,
  side,
}) => {
  const { leftBlocks, rightBlocks } = React.useMemo(
    () => computeBlockDiff(leftContent, rightContent),
    [leftContent, rightContent]
  );

  const blocks = side === 'left' ? leftBlocks : rightBlocks;

  return (
    <div className={styles.diffContent}>
      {blocks.map((block, index) => (
        <DiffBlockRenderer
          key={index}
          block={block}
          searchTerm={searchTerm}
          side={side}
          index={index}
        />
      ))}
    </div>
  );
};

const DiffBlockRenderer: React.FC<{
  block: DiffBlock;
  searchTerm?: string;
  side: 'left' | 'right';
  index: number;
}> = ({ block, searchTerm, side, index }) => {
  const getBlockClass = () => {
    switch (block.type) {
      case 'added':
        // Added blocks exist in left (JS) only
        return side === 'left' ? styles.added : styles.placeholder;
      case 'removed':
        // Removed blocks exist in right (no-JS) only
        return side === 'right' ? styles.removed : styles.placeholder;
      case 'modified':
        return styles.modified;
      default:
        return styles.unchanged;
    }
  };

  // For placeholders (gaps), render empty space with matching height hint
  // Added blocks only exist on left side, so show placeholder on right
  // Removed blocks only exist on right side, so show placeholder on left
  if ((block.type === 'added' && side === 'right') ||
      (block.type === 'removed' && side === 'left')) {
    return <div className={styles.placeholder} data-block-index={index}>&nbsp;</div>;
  }

  // For modified blocks, show inline word diff
  if (block.type === 'modified' && block.otherContent) {
    return (
      <div className={`${styles.block} ${styles.modified}`} data-block-index={index}>
        <InlineDiff
          oldText={side === 'right' ? block.content : block.otherContent}
          newText={side === 'left' ? block.content : block.otherContent}
          showSide={side === 'left' ? 'new' : 'old'}
        />
      </div>
    );
  }

  return (
    <div className={`${styles.block} ${getBlockClass()}`} data-block-index={index}>
      <MarkdownContent content={block.content} searchTerm={searchTerm} />
    </div>
  );
};
