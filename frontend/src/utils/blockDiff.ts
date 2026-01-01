export type BlockType = 'heading' | 'paragraph' | 'list' | 'blockquote' | 'section' | 'other';

export interface DiffBlock {
  type: 'unchanged' | 'modified' | 'added' | 'removed';
  blockType: BlockType;
  content: string;
  otherContent?: string; // For modified blocks
}

export interface DiffResult {
  leftBlocks: DiffBlock[];
  rightBlocks: DiffBlock[];
}

// Detect block type from content
function detectBlockType(block: string): BlockType {
  const trimmed = block.trim();
  if (trimmed.startsWith('#')) return 'heading';
  if (trimmed.startsWith('- ') || /^\d+\./.test(trimmed)) return 'list';
  if (trimmed.startsWith('>')) return 'blockquote';
  if (trimmed.startsWith('[') && trimmed.includes(']')) return 'section';
  return 'paragraph';
}

// Jaccard similarity based on words
function jaccardSimilarity(text1: string, text2: string): number {
  // Strip punctuation (Unicode-aware: keeps letters, numbers, whitespace from any language)
  const normalize = (text: string) =>
    text.toLowerCase().replace(/[^\p{L}\p{N}\s]/gu, '');

  const words1 = new Set(normalize(text1).split(/\s+/).filter(w => w));
  const words2 = new Set(normalize(text2).split(/\s+/).filter(w => w));

  if (words1.size === 0 && words2.size === 0) return 1;

  let intersection = 0;
  words1.forEach(word => {
    if (words2.has(word)) intersection++;
  });

  const union = new Set([...words1, ...words2]).size;
  return union === 0 ? 1 : intersection / union;
}

// Main diff function
export function computeBlockDiff(leftMarkdown: string, rightMarkdown: string): DiffResult {
  // Split into blocks
  const leftBlocks = leftMarkdown.split(/\n\n+/).filter(b => b.trim());
  const rightBlocks = rightMarkdown.split(/\n\n+/).filter(b => b.trim());

  const leftResult: DiffBlock[] = [];
  const rightResult: DiffBlock[] = [];

  // Track which right blocks have been matched
  const matchedRight = new Set<number>();

  // Match left blocks to right blocks
  for (let i = 0; i < leftBlocks.length; i++) {
    const leftBlock = leftBlocks[i];
    const leftType = detectBlockType(leftBlock);

    let bestMatch = -1;
    let bestScore = 0;

    // Find best matching right block
    for (let j = 0; j < rightBlocks.length; j++) {
      if (matchedRight.has(j)) continue;

      const rightBlock = rightBlocks[j];
      const rightType = detectBlockType(rightBlock);

      // Types must match for comparison
      if (leftType !== rightType) continue;

      const score = jaccardSimilarity(leftBlock, rightBlock);

      // Use position proximity as tie-breaker
      const positionBonus = 1 - Math.abs(i - j) / Math.max(leftBlocks.length, rightBlocks.length) * 0.1;
      const adjustedScore = score + positionBonus * 0.01;

      if (adjustedScore > bestScore) {
        bestScore = adjustedScore;
        bestMatch = j;
      }
    }

    if (bestMatch >= 0 && bestScore >= 0.6) {
      const rightBlock = rightBlocks[bestMatch];

      // Check length ratio - don't pair blocks with vastly different sizes
      // This prevents poor inline diffs when one block has much more content
      const leftLen = leftBlock.length;
      const rightLen = rightBlock.length;
      const lengthRatio = Math.max(leftLen, rightLen) / Math.min(leftLen, rightLen);

      if (lengthRatio > 3) {
        // Blocks are too different in size - treat as separate add/remove
        leftResult.push({ type: 'added', blockType: leftType, content: leftBlock });
      } else {
        matchedRight.add(bestMatch);

        // Count actual words (not just unique) to detect quantity differences
        const leftWordCount = leftBlock.split(/\s+/).filter(w => w).length;
        const rightWordCount = rightBlock.split(/\s+/).filter(w => w).length;
        const wordCountRatio = Math.max(leftWordCount, rightWordCount) / Math.max(1, Math.min(leftWordCount, rightWordCount));

        // If word counts differ significantly (>20%) or similarity is <0.9, mark as modified
        if (bestScore >= 0.9 && wordCountRatio <= 1.2) {
          // Unchanged - high similarity AND similar word counts
          leftResult.push({ type: 'unchanged', blockType: leftType, content: leftBlock });
          rightResult.push({ type: 'unchanged', blockType: leftType, content: rightBlock });
        } else {
          // Modified - either vocabulary or word count differs
          leftResult.push({ type: 'modified', blockType: leftType, content: leftBlock, otherContent: rightBlock });
          rightResult.push({ type: 'modified', blockType: leftType, content: rightBlock, otherContent: leftBlock });
        }
      }
    } else {
      // Added (exists in left/JS but not right/no-JS)
      leftResult.push({ type: 'added', blockType: leftType, content: leftBlock });
    }
  }

  // Add unmatched right blocks as removed
  for (let j = 0; j < rightBlocks.length; j++) {
    if (!matchedRight.has(j)) {
      const rightBlock = rightBlocks[j];
      rightResult.push({ type: 'removed', blockType: detectBlockType(rightBlock), content: rightBlock });
    }
  }

  return { leftBlocks: leftResult, rightBlocks: rightResult };
}
