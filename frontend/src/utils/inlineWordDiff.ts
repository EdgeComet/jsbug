import { diffWords } from 'diff';

export interface WordDiff {
  text: string;
  type: 'unchanged' | 'added' | 'removed';
}

export function computeInlineWordDiff(oldText: string, newText: string): WordDiff[] {
  const changes = diffWords(oldText, newText);

  return changes.map(change => ({
    text: change.value,
    type: change.added ? 'added' : change.removed ? 'removed' : 'unchanged'
  }));
}
