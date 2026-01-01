import React from 'react';
import { computeInlineWordDiff } from '../../utils/inlineWordDiff';
import styles from './InlineDiff.module.css';

interface InlineDiffProps {
  oldText: string;
  newText: string;
  showSide: 'old' | 'new';
}

export const InlineDiff: React.FC<InlineDiffProps> = ({ oldText, newText, showSide }) => {
  const diffs = React.useMemo(
    () => computeInlineWordDiff(oldText, newText),
    [oldText, newText]
  );

  return (
    <span>
      {diffs.map((diff, i) => {
        // When showing "new" side, hide removed words
        if (showSide === 'new' && diff.type === 'removed') return null;
        // When showing "old" side, hide added words
        if (showSide === 'old' && diff.type === 'added') return null;

        const className =
          diff.type === 'added' ? styles.added :
          diff.type === 'removed' ? styles.removed :
          '';

        return (
          <span key={i} className={className}>
            {diff.text}
          </span>
        );
      })}
    </span>
  );
};
