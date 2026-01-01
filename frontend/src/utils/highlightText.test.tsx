import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { highlightText } from './highlightText';

describe('highlightText', () => {
  it('should return plain text when search term is empty', () => {
    const result = highlightText('Hello world', '', 'highlight');
    expect(result).toBe('Hello world');
  });

  it('should return plain text when search term is whitespace', () => {
    const result = highlightText('Hello world', '   ', 'highlight');
    expect(result).toBe('Hello world');
  });

  it('should highlight matching text', () => {
    const result = highlightText('Hello world', 'world', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.textContent).toBe('Hello world');
    expect(container.querySelectorAll('mark')).toHaveLength(1);
    expect(container.querySelector('mark')?.textContent).toBe('world');
  });

  it('should be case-insensitive', () => {
    const result = highlightText('Hello World', 'WORLD', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelectorAll('mark')).toHaveLength(1);
    expect(container.querySelector('mark')?.textContent).toBe('World');
  });

  it('should highlight multiple occurrences', () => {
    const result = highlightText('foo bar foo baz foo', 'foo', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelectorAll('mark')).toHaveLength(3);
  });

  it('should handle no matches', () => {
    const result = highlightText('Hello world', 'xyz', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.textContent).toBe('Hello world');
    expect(container.querySelectorAll('mark')).toHaveLength(0);
  });

  it('should escape regex special characters in search term', () => {
    const result = highlightText('test (foo) bar', '(foo)', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelectorAll('mark')).toHaveLength(1);
    expect(container.querySelector('mark')?.textContent).toBe('(foo)');
  });

  it('should handle search term with brackets', () => {
    const result = highlightText('array[0] and array[1]', '[0]', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelectorAll('mark')).toHaveLength(1);
    expect(container.querySelector('mark')?.textContent).toBe('[0]');
  });

  it('should handle search term with dots', () => {
    const result = highlightText('config.json and data.json', '.json', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelectorAll('mark')).toHaveLength(2);
  });

  it('should handle search term at start of text', () => {
    const result = highlightText('Hello world', 'Hello', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelector('mark')?.textContent).toBe('Hello');
    expect(container.textContent).toBe('Hello world');
  });

  it('should handle search term at end of text', () => {
    const result = highlightText('Hello world', 'world', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelector('mark')?.textContent).toBe('world');
    expect(container.textContent).toBe('Hello world');
  });

  it('should handle search term as entire text', () => {
    const result = highlightText('match', 'match', 'highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelectorAll('mark')).toHaveLength(1);
    expect(container.textContent).toBe('match');
  });

  it('should preserve original case in highlighted text', () => {
    const result = highlightText('Hello HELLO hello', 'hello', 'highlight');
    const { container } = render(<>{result}</>);

    const marks = container.querySelectorAll('mark');
    expect(marks).toHaveLength(3);
    expect(marks[0].textContent).toBe('Hello');
    expect(marks[1].textContent).toBe('HELLO');
    expect(marks[2].textContent).toBe('hello');
  });

  it('should apply the provided highlight class', () => {
    const result = highlightText('test text', 'text', 'custom-highlight');
    const { container } = render(<>{result}</>);

    expect(container.querySelector('mark')?.className).toBe('custom-highlight');
  });
});
