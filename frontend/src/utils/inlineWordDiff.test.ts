import { describe, it, expect } from 'vitest';
import { computeInlineWordDiff, WordDiff } from './inlineWordDiff';

// Helper to extract just words (no whitespace) for easier assertions
function getWords(diffs: WordDiff[]): WordDiff[] {
  return diffs.filter(d => d.text.trim() !== '');
}

describe('computeInlineWordDiff', () => {
  it('should return empty result for empty inputs', () => {
    const result = computeInlineWordDiff('', '');
    const words = getWords(result);
    expect(words).toEqual([]);
  });

  it('should mark all words unchanged for identical text', () => {
    const result = computeInlineWordDiff('hello world', 'hello world');
    const words = getWords(result);

    expect(words).toHaveLength(2);
    expect(words[0]).toEqual({ text: 'hello', type: 'unchanged' });
    expect(words[1]).toEqual({ text: 'world', type: 'unchanged' });
  });

  it('should detect added words', () => {
    const result = computeInlineWordDiff('hello', 'hello world');
    const words = getWords(result);

    expect(words).toHaveLength(2);
    expect(words[0]).toEqual({ text: 'hello', type: 'unchanged' });
    expect(words[1]).toEqual({ text: 'world', type: 'added' });
  });

  it('should detect removed words', () => {
    const result = computeInlineWordDiff('hello world', 'hello');
    const words = getWords(result);

    expect(words).toHaveLength(2);
    expect(words[0]).toEqual({ text: 'hello', type: 'unchanged' });
    expect(words[1]).toEqual({ text: 'world', type: 'removed' });
  });

  it('should detect modified words (replacement)', () => {
    const result = computeInlineWordDiff('hello world', 'hello universe');
    const words = getWords(result);

    expect(words).toHaveLength(3);
    expect(words[0]).toEqual({ text: 'hello', type: 'unchanged' });
    expect(words[1]).toEqual({ text: 'world', type: 'removed' });
    expect(words[2]).toEqual({ text: 'universe', type: 'added' });
  });

  it('should handle reordered words optimally', () => {
    // "a b c d" vs "a c d b" - should keep a, c, d unchanged
    const result = computeInlineWordDiff('a b c d', 'a c d b');
    const words = getWords(result);

    const unchanged = words.filter(w => w.type === 'unchanged');
    const removed = words.filter(w => w.type === 'removed');
    const added = words.filter(w => w.type === 'added');

    expect(unchanged).toHaveLength(3); // a, c, d
    expect(removed).toHaveLength(1);   // b
    expect(added).toHaveLength(1);     // b
  });

  it('should handle complete reversal', () => {
    const result = computeInlineWordDiff('a b c', 'c b a');
    const words = getWords(result);

    // Should find at least 1 unchanged word
    const unchanged = words.filter(w => w.type === 'unchanged');
    expect(unchanged.length).toBeGreaterThanOrEqual(1);
  });

  it('should handle swapped adjacent words', () => {
    const result = computeInlineWordDiff('hello world', 'world hello');
    const words = getWords(result);

    // At least one word should be unchanged
    const unchanged = words.filter(w => w.type === 'unchanged');
    expect(unchanged.length).toBeGreaterThanOrEqual(1);
  });

  it('should preserve whitespace in output', () => {
    const result = computeInlineWordDiff('hello world', 'hello world');

    // Should include the space between words
    expect(result).toHaveLength(3);
    expect(result[1].text).toBe(' ');
    expect(result[1].type).toBe('unchanged');
  });

  it('should handle multiple spaces', () => {
    const result = computeInlineWordDiff('hello  world', 'hello  world');

    expect(result[1].text).toBe('  ');
    expect(result[1].type).toBe('unchanged');
  });

  it('should handle completely different text', () => {
    const result = computeInlineWordDiff('foo bar', 'baz qux');
    const words = getWords(result);

    const removed = words.filter(w => w.type === 'removed');
    const added = words.filter(w => w.type === 'added');

    expect(removed).toHaveLength(2);
    expect(added).toHaveLength(2);
  });

  it('should handle one side empty', () => {
    const resultLeft = computeInlineWordDiff('hello world', '');
    const wordsLeft = getWords(resultLeft);
    expect(wordsLeft.every(w => w.type === 'removed')).toBe(true);

    const resultRight = computeInlineWordDiff('', 'hello world');
    const wordsRight = getWords(resultRight);
    expect(wordsRight.every(w => w.type === 'added')).toBe(true);
  });

  it('should handle duplicate words correctly', () => {
    const result = computeInlineWordDiff('a b a', 'a a b');
    const words = getWords(result);

    // Should find optimal alignment with 2 unchanged
    const unchanged = words.filter(w => w.type === 'unchanged');
    expect(unchanged.length).toBeGreaterThanOrEqual(2);
  });

  it('should handle non-English text (Unicode)', () => {
    // Russian
    const resultRu = computeInlineWordDiff('Привет мир', 'Привет прекрасный мир');
    const wordsRu = getWords(resultRu);
    expect(wordsRu.find(w => w.text === 'Привет')?.type).toBe('unchanged');
    expect(wordsRu.find(w => w.text === 'мир')?.type).toBe('unchanged');
    expect(wordsRu.find(w => w.text === 'прекрасный')?.type).toBe('added');

    // Chinese
    const resultZh = computeInlineWordDiff('你好 世界', '你好 美丽 世界');
    const wordsZh = getWords(resultZh);
    expect(wordsZh.find(w => w.text === '你好')?.type).toBe('unchanged');
    expect(wordsZh.find(w => w.text === '世界')?.type).toBe('unchanged');
    expect(wordsZh.find(w => w.text === '美丽')?.type).toBe('added');

    // Arabic
    const resultAr = computeInlineWordDiff('مرحبا عالم', 'مرحبا الجميل عالم');
    const wordsAr = getWords(resultAr);
    expect(wordsAr.find(w => w.text === 'مرحبا')?.type).toBe('unchanged');
    expect(wordsAr.find(w => w.text === 'عالم')?.type).toBe('unchanged');
    expect(wordsAr.find(w => w.text === 'الجميل')?.type).toBe('added');
  });

  it('should match words with different punctuation as unchanged', () => {
    // Words with different trailing punctuation should match
    const result1 = computeInlineWordDiff('Hello, world!', 'Hello: world.');
    const words1 = getWords(result1);
    // Both "Hello" and "world" should be unchanged despite punctuation
    const unchanged1 = words1.filter(w => w.type === 'unchanged');
    expect(unchanged1.length).toBe(2);

    // Same word with vs without punctuation
    const result2 = computeInlineWordDiff('Wixel, is great', 'Wixel is great');
    const words2 = getWords(result2);
    expect(words2.find(w => w.text.includes('Wixel'))?.type).toBe('unchanged');
    expect(words2.find(w => w.text === 'is')?.type).toBe('unchanged');
    expect(words2.find(w => w.text === 'great')?.type).toBe('unchanged');

    // Case insensitivity with punctuation
    const result3 = computeInlineWordDiff('HELLO world', 'hello WORLD');
    const words3 = getWords(result3);
    const unchanged3 = words3.filter(w => w.type === 'unchanged');
    expect(unchanged3.length).toBe(2);
  });
});
