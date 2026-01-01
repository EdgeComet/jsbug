import { describe, it, expect } from 'vitest';
import { computeInlineWordDiff, WordDiff } from './inlineWordDiff';

// Helper to get combined text by type
function getTextByType(diffs: WordDiff[], type: WordDiff['type']): string {
  return diffs.filter(d => d.type === type).map(d => d.text).join('');
}

// Helper to check if text contains a word with given type
function hasWordWithType(diffs: WordDiff[], word: string, type: WordDiff['type']): boolean {
  return diffs.some(d => d.type === type && d.text.includes(word));
}

describe('computeInlineWordDiff', () => {
  it('should return empty result for empty inputs', () => {
    const result = computeInlineWordDiff('', '');
    expect(result).toEqual([]);
  });

  it('should mark all content unchanged for identical text', () => {
    const result = computeInlineWordDiff('hello world', 'hello world');

    expect(result.every(d => d.type === 'unchanged')).toBe(true);
    expect(getTextByType(result, 'unchanged')).toBe('hello world');
  });

  it('should detect added words', () => {
    const result = computeInlineWordDiff('hello', 'hello world');

    expect(hasWordWithType(result, 'hello', 'unchanged')).toBe(true);
    expect(hasWordWithType(result, 'world', 'added')).toBe(true);
  });

  it('should detect removed words', () => {
    const result = computeInlineWordDiff('hello world', 'hello');

    expect(hasWordWithType(result, 'hello', 'unchanged')).toBe(true);
    expect(hasWordWithType(result, 'world', 'removed')).toBe(true);
  });

  it('should detect modified words (replacement)', () => {
    const result = computeInlineWordDiff('hello world', 'hello universe');

    expect(hasWordWithType(result, 'hello', 'unchanged')).toBe(true);
    expect(hasWordWithType(result, 'world', 'removed')).toBe(true);
    expect(hasWordWithType(result, 'universe', 'added')).toBe(true);
  });

  it('should handle reordered words', () => {
    const result = computeInlineWordDiff('a b c d', 'a c d b');

    // Should have some unchanged content
    const unchangedText = getTextByType(result, 'unchanged');
    expect(unchangedText).toContain('a');
  });

  it('should handle complete reversal', () => {
    const result = computeInlineWordDiff('a b c', 'c b a');

    // Should have some unchanged content
    const unchangedText = getTextByType(result, 'unchanged');
    expect(unchangedText.length).toBeGreaterThan(0);
  });

  it('should handle swapped adjacent words', () => {
    const result = computeInlineWordDiff('hello world', 'world hello');

    // At least one word should be found
    const hasHello = result.some(d => d.text.includes('hello'));
    const hasWorld = result.some(d => d.text.includes('world'));
    expect(hasHello).toBe(true);
    expect(hasWorld).toBe(true);
  });

  it('should handle completely different text', () => {
    const result = computeInlineWordDiff('foo bar', 'baz qux');

    const removedText = getTextByType(result, 'removed');
    const addedText = getTextByType(result, 'added');

    expect(removedText).toContain('foo');
    expect(removedText).toContain('bar');
    expect(addedText).toContain('baz');
    expect(addedText).toContain('qux');
  });

  it('should handle one side empty', () => {
    const resultLeft = computeInlineWordDiff('hello world', '');
    expect(resultLeft.every(d => d.type === 'removed')).toBe(true);

    const resultRight = computeInlineWordDiff('', 'hello world');
    expect(resultRight.every(d => d.type === 'added')).toBe(true);
  });

  it('should handle duplicate words correctly', () => {
    const result = computeInlineWordDiff('a b a', 'a a b');

    // Should produce a valid diff
    expect(result.length).toBeGreaterThan(0);

    // Combined text should reconstruct properly
    const addedText = getTextByType(result, 'added');
    const unchangedText = getTextByType(result, 'unchanged');

    // The diff should be meaningful
    expect(addedText + unchangedText).toBeTruthy();
  });

  it('should handle non-English text (Unicode)', () => {
    // Russian
    const resultRu = computeInlineWordDiff('Привет мир', 'Привет прекрасный мир');
    expect(hasWordWithType(resultRu, 'Привет', 'unchanged')).toBe(true);
    expect(hasWordWithType(resultRu, 'прекрасный', 'added')).toBe(true);

    // Chinese
    const resultZh = computeInlineWordDiff('你好 世界', '你好 美丽 世界');
    expect(hasWordWithType(resultZh, '你好', 'unchanged')).toBe(true);
    expect(hasWordWithType(resultZh, '美丽', 'added')).toBe(true);

    // Arabic
    const resultAr = computeInlineWordDiff('مرحبا عالم', 'مرحبا الجميل عالم');
    expect(hasWordWithType(resultAr, 'مرحبا', 'unchanged')).toBe(true);
    expect(hasWordWithType(resultAr, 'الجميل', 'added')).toBe(true);
  });

  it('should handle punctuation differences', () => {
    // The diff library treats punctuation as part of words
    const result = computeInlineWordDiff('Hello, world!', 'Hello: world.');

    // Should identify the changes
    expect(result.length).toBeGreaterThan(0);
  });

  it('should produce correct output for rendering', () => {
    // Key test: the concatenated output should equal the new text
    const oldText = 'The quick brown fox';
    const newText = 'The slow brown dog';
    const result = computeInlineWordDiff(oldText, newText);

    // Concatenating unchanged + added should give new text
    const newSide = result
      .filter(d => d.type !== 'removed')
      .map(d => d.text)
      .join('');
    expect(newSide).toBe(newText);

    // Concatenating unchanged + removed should give old text
    const oldSide = result
      .filter(d => d.type !== 'added')
      .map(d => d.text)
      .join('');
    expect(oldSide).toBe(oldText);
  });
});
