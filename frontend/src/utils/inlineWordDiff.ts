export interface WordDiff {
  text: string;
  type: 'unchanged' | 'added' | 'removed';
}

// Normalize token for comparison: lowercase, strip punctuation
// Keeps Unicode letters and numbers from any language
function normalizeToken(token: string): string {
  return token.toLowerCase().replace(/[^\p{L}\p{N}]/gu, '');
}

// Find index of a normalized token in an array starting from offset
function findNormalizedIndex(tokens: string[], normalized: string[], target: string, startIndex: number): number {
  for (let k = startIndex; k < normalized.length; k++) {
    if (normalized[k] === target) return k;
  }
  return -1;
}

export function computeInlineWordDiff(oldText: string, newText: string): WordDiff[] {
  const oldWords = oldText.split(/(\s+)/);
  const newWords = newText.split(/(\s+)/);

  // Pre-compute normalized versions for comparison
  const oldNorm = oldWords.map(normalizeToken);
  const newNorm = newWords.map(normalizeToken);

  // Greedy two-pointer diff with look-ahead heuristic
  const result: WordDiff[] = [];

  let i = 0, j = 0;
  while (i < oldWords.length || j < newWords.length) {
    if (i >= oldWords.length) {
      result.push({ text: newWords[j], type: 'added' });
      j++;
    } else if (j >= newWords.length) {
      result.push({ text: oldWords[i], type: 'removed' });
      i++;
    } else if (oldNorm[i] === newNorm[j]) {
      // Tokens match (ignoring punctuation differences)
      // Use the new text for display (preserves updated punctuation)
      result.push({ text: newWords[j], type: 'unchanged' });
      i++;
      j++;
    } else {
      // Check if normalized word appears later
      const oldInNew = findNormalizedIndex(newWords, newNorm, oldNorm[i], j);
      const newInOld = findNormalizedIndex(oldWords, oldNorm, newNorm[j], i);

      // Check if tokens are content words vs whitespace/punctuation only
      const oldIsContent = oldNorm[i].length > 0;
      const newIsContent = newNorm[j].length > 0;

      if (!oldIsContent && !newIsContent) {
        // Both are just punctuation/whitespace - treat as unchanged
        result.push({ text: newWords[j], type: 'unchanged' });
        i++;
        j++;
      } else if (oldInNew === -1 && !oldIsContent) {
        // Old is just punctuation not found - skip it
        i++;
      } else if (newInOld === -1 && !newIsContent) {
        // New is just punctuation not found - add it unchanged
        result.push({ text: newWords[j], type: 'unchanged' });
        j++;
      } else if (oldInNew === -1) {
        result.push({ text: oldWords[i], type: 'removed' });
        i++;
      } else if (newInOld === -1) {
        result.push({ text: newWords[j], type: 'added' });
        j++;
      } else if (oldIsContent && !newIsContent) {
        // Old word is content, new is whitespace/punct - add new, keep looking for old
        result.push({ text: newWords[j], type: 'added' });
        j++;
      } else if (!oldIsContent && newIsContent) {
        // Old is whitespace/punct, new is content - remove old, keep looking for new
        result.push({ text: oldWords[i], type: 'removed' });
        i++;
      } else if (oldInNew - j < newInOld - i) {
        result.push({ text: newWords[j], type: 'added' });
        j++;
      } else {
        result.push({ text: oldWords[i], type: 'removed' });
        i++;
      }
    }
  }

  return result;
}
