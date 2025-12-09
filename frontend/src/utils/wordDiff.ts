export interface WordDiff {
  added: string[];    // words in current but not in compare
  removed: string[];  // words in compare but not in current
}

interface TokenizedWord {
  normalized: string;
  original: string;
}

/**
 * Tokenize text into words, preserving original form alongside normalized version
 */
function tokenize(text: string): TokenizedWord[] {
  return text
    .split(/[\s.,;:!?'"()[\]{}↗&]+/)
    .filter(word => word.length > 0)
    .map(word => ({
      original: word,
      normalized: word.toLowerCase().replace(/[^\w\u00C0-\u024F]/g, '')
    }))
    .filter(w => w.normalized.length > 0);
}

/**
 * Compute word differences between two body texts
 * Returns original word forms (not normalized) so they can be found when searching
 */
export function computeWordDiff(current: string, compare: string): WordDiff {
  const currentTokens = tokenize(current);
  const compareTokens = tokenize(compare);

  // Build maps: normalized -> first original occurrence
  const currentMap = new Map<string, string>();
  for (const t of currentTokens) {
    if (!currentMap.has(t.normalized)) {
      currentMap.set(t.normalized, t.original);
    }
  }

  const compareMap = new Map<string, string>();
  for (const t of compareTokens) {
    if (!compareMap.has(t.normalized)) {
      compareMap.set(t.normalized, t.original);
    }
  }

  const compareSet = new Set(compareTokens.map(t => t.normalized));
  const currentSet = new Set(currentTokens.map(t => t.normalized));

  const added: string[] = [];
  const removed: string[] = [];

  // Words in current but not in compare (use original form)
  for (const [normalized, original] of currentMap) {
    if (!compareSet.has(normalized)) {
      added.push(original);
    }
  }

  // Words in compare but not in current (use original form)
  for (const [normalized, original] of compareMap) {
    if (!currentSet.has(normalized)) {
      removed.push(original);
    }
  }

  // Sort alphabetically (case-insensitive) for consistent display
  added.sort((a, b) => a.toLowerCase().localeCompare(b.toLowerCase()));
  removed.sort((a, b) => a.toLowerCase().localeCompare(b.toLowerCase()));

  return { added, removed };
}
