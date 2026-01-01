import { describe, it, expect } from 'vitest';
import { computeBlockDiff } from './blockDiff';

describe('computeBlockDiff', () => {
  it('should return empty results for empty inputs', () => {
    const result = computeBlockDiff('', '');
    expect(result.leftBlocks).toEqual([]);
    expect(result.rightBlocks).toEqual([]);
  });

  it('should detect unchanged blocks with identical content', () => {
    const markdown = '# Heading\n\nSome paragraph text here.';
    const result = computeBlockDiff(markdown, markdown);

    expect(result.leftBlocks).toHaveLength(2);
    expect(result.rightBlocks).toHaveLength(2);
    expect(result.leftBlocks[0].type).toBe('unchanged');
    expect(result.leftBlocks[1].type).toBe('unchanged');
  });

  it('should detect added blocks (only in left)', () => {
    const left = '# Heading\n\nNew paragraph added by JS.';
    const right = '# Heading';
    const result = computeBlockDiff(left, right);

    expect(result.leftBlocks).toHaveLength(2);
    expect(result.leftBlocks[0].type).toBe('unchanged');
    expect(result.leftBlocks[1].type).toBe('added');
    expect(result.leftBlocks[1].content).toBe('New paragraph added by JS.');
  });

  it('should detect removed blocks (only in right)', () => {
    const left = '# Heading';
    const right = '# Heading\n\nParagraph only in no-JS version.';
    const result = computeBlockDiff(left, right);

    expect(result.leftBlocks).toHaveLength(1);
    expect(result.rightBlocks).toHaveLength(2);
    expect(result.rightBlocks[1].type).toBe('removed');
    expect(result.rightBlocks[1].content).toBe('Paragraph only in no-JS version.');
  });

  it('should detect modified blocks with similar content', () => {
    const left = '# Heading\n\nThis is the updated paragraph with new words.';
    const right = '# Heading\n\nThis is the original paragraph with old words.';
    const result = computeBlockDiff(left, right);

    expect(result.leftBlocks).toHaveLength(2);
    expect(result.leftBlocks[0].type).toBe('unchanged');
    expect(result.leftBlocks[1].type).toBe('modified');
    expect(result.leftBlocks[1].otherContent).toBe('This is the original paragraph with old words.');
  });

  it('should correctly detect block types', () => {
    const markdown = `# Heading

Regular paragraph.

- List item 1
- List item 2

> A blockquote here

[Section] Some section content`;

    const result = computeBlockDiff(markdown, markdown);

    expect(result.leftBlocks[0].blockType).toBe('heading');
    expect(result.leftBlocks[1].blockType).toBe('paragraph');
    expect(result.leftBlocks[2].blockType).toBe('list');
    expect(result.leftBlocks[3].blockType).toBe('blockquote');
    expect(result.leftBlocks[4].blockType).toBe('section');
  });

  it('should detect numbered lists with any number', () => {
    const markdown = `1. First item

2. Second item

10. Tenth item

99. Ninety-ninth item`;

    const result = computeBlockDiff(markdown, markdown);

    expect(result.leftBlocks[0].blockType).toBe('list');
    expect(result.leftBlocks[1].blockType).toBe('list');
    expect(result.leftBlocks[2].blockType).toBe('list');
    expect(result.leftBlocks[3].blockType).toBe('list');
  });

  it('should handle multiple paragraphs with varying similarity', () => {
    const left = `# Title

First paragraph unchanged.

Second paragraph with minor edits here.

Third paragraph completely new.`;

    const right = `# Title

First paragraph unchanged.

Second paragraph with minor changes here.`;

    const result = computeBlockDiff(left, right);

    expect(result.leftBlocks[0].type).toBe('unchanged'); // Title
    expect(result.leftBlocks[1].type).toBe('unchanged'); // First para
    expect(result.leftBlocks[2].type).toBe('modified');  // Second para
    expect(result.leftBlocks[3].type).toBe('added');     // Third para
  });

  it('should not match blocks of different types', () => {
    const left = '# This is a heading';
    const right = 'This is a heading';
    const result = computeBlockDiff(left, right);

    // Heading should be added, paragraph should be removed
    expect(result.leftBlocks[0].type).toBe('added');
    expect(result.leftBlocks[0].blockType).toBe('heading');
    expect(result.rightBlocks[0].type).toBe('removed');
    expect(result.rightBlocks[0].blockType).toBe('paragraph');
  });

  it('should handle blocks with only whitespace differences as unchanged', () => {
    const left = '# Heading\n\nParagraph with text.';
    const right = '# Heading\n\n\n\nParagraph with text.';
    const result = computeBlockDiff(left, right);

    expect(result.leftBlocks).toHaveLength(2);
    expect(result.rightBlocks).toHaveLength(2);
    expect(result.leftBlocks[0].type).toBe('unchanged');
    expect(result.leftBlocks[1].type).toBe('unchanged');
  });

  it('should treat punctuation-only differences as unchanged', () => {
    const left = '# Hello world!\n\nRead more...';
    const right = '# Hello world\n\nRead more';
    const result = computeBlockDiff(left, right);

    expect(result.leftBlocks).toHaveLength(2);
    expect(result.rightBlocks).toHaveLength(2);
    expect(result.leftBlocks[0].type).toBe('unchanged');
    expect(result.leftBlocks[1].type).toBe('unchanged');
  });

  it('should handle punctuation differences in non-English text', () => {
    // Russian
    const leftRu = 'Привет мир!';
    const rightRu = 'Привет мир';
    const resultRu = computeBlockDiff(leftRu, rightRu);
    expect(resultRu.leftBlocks[0].type).toBe('unchanged');

    // Chinese
    const leftZh = '你好世界！';
    const rightZh = '你好世界';
    const resultZh = computeBlockDiff(leftZh, rightZh);
    expect(resultZh.leftBlocks[0].type).toBe('unchanged');

    // Arabic
    const leftAr = 'مرحبا بالعالم!';
    const rightAr = 'مرحبا بالعالم';
    const resultAr = computeBlockDiff(leftAr, rightAr);
    expect(resultAr.leftBlocks[0].type).toBe('unchanged');
  });

  it('should handle reordered blocks with distinct content', () => {
    // Blocks A, B, C reordered to B, A, C
    const left = `# First Heading

This is the first paragraph with unique content.

# Second Heading

This is the second paragraph with different content.

# Third Heading

This is the third paragraph.`;

    const right = `# Second Heading

This is the second paragraph with different content.

# First Heading

This is the first paragraph with unique content.

# Third Heading

This is the third paragraph.`;

    const result = computeBlockDiff(left, right);

    // All blocks should be matched as unchanged (reordering detected)
    const unchangedLeft = result.leftBlocks.filter(b => b.type === 'unchanged');
    const unchangedRight = result.rightBlocks.filter(b => b.type === 'unchanged');

    // Should find at least 4 unchanged blocks (headings + paragraphs)
    expect(unchangedLeft.length).toBeGreaterThanOrEqual(4);
    expect(unchangedRight.length).toBeGreaterThanOrEqual(4);
  });

  it('should handle swapped adjacent blocks', () => {
    const left = `# Alpha

Content for alpha.

# Beta

Content for beta.`;

    const right = `# Beta

Content for beta.

# Alpha

Content for alpha.`;

    const result = computeBlockDiff(left, right);

    // All 4 blocks should be unchanged
    const unchangedLeft = result.leftBlocks.filter(b => b.type === 'unchanged');
    expect(unchangedLeft.length).toBe(4);
  });

  it('should not pair blocks with vastly different lengths', () => {
    // Left block is much longer than right block
    const left = `This is a very long paragraph with lots of content about many different topics including web development, JavaScript frameworks, React components, TypeScript types, and various other programming concepts that make this text significantly longer than the other side.`;

    const right = `Short text.`;

    const result = computeBlockDiff(left, right);

    // Should NOT be paired as modified (length ratio > 3)
    // Instead should be separate added/removed
    const modifiedLeft = result.leftBlocks.filter(b => b.type === 'modified');
    expect(modifiedLeft.length).toBe(0);

    // Left block should be added, right block should be removed
    expect(result.leftBlocks[0].type).toBe('added');
    expect(result.rightBlocks[0].type).toBe('removed');
  });
});
