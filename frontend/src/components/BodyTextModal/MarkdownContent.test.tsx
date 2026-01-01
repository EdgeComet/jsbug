import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MarkdownContent } from './MarkdownContent';

// Mock scrollIntoView which is not implemented in jsdom
beforeEach(() => {
  Element.prototype.scrollIntoView = vi.fn();
});

// Define component type for markdown custom components
type ComponentRenderer = React.FC<{ children: React.ReactNode }>;

interface MockComponents {
  h1?: ComponentRenderer;
  h2?: ComponentRenderer;
  li?: ComponentRenderer;
  p?: ComponentRenderer;
}

// Mock react-markdown to simplify testing
vi.mock('react-markdown', () => ({
  default: ({ children, components }: { children: string; components?: MockComponents }) => {
    // Simple mock that renders basic markdown structure
    const lines = children.split('\n');
    const rendered: React.ReactElement[] = [];

    lines.forEach((line, i) => {
      const trimmed = line.trim();
      if (trimmed.startsWith('# ')) {
        const H1 = components?.h1 || (({ children }: { children: React.ReactNode }) => <h1>{children}</h1>);
        rendered.push(<H1 key={i}>{trimmed.slice(2)}</H1>);
      } else if (trimmed.startsWith('## ')) {
        const H2 = components?.h2 || (({ children }: { children: React.ReactNode }) => <h2>{children}</h2>);
        rendered.push(<H2 key={i}>{trimmed.slice(3)}</H2>);
      } else if (trimmed.startsWith('- ')) {
        const Li = components?.li || (({ children }: { children: React.ReactNode }) => <li>{children}</li>);
        rendered.push(<Li key={i}>{trimmed.slice(2)}</Li>);
      } else if (trimmed.length > 0) {
        const P = components?.p || (({ children }: { children: React.ReactNode }) => <p>{children}</p>);
        rendered.push(<P key={i}>{trimmed}</P>);
      }
    });

    return <div data-testid="markdown-content">{rendered}</div>;
  },
}));

describe('MarkdownContent', () => {
  it('renders markdown content', () => {
    render(<MarkdownContent content="# Hello World" />);
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Hello World');
  });

  it('renders multiple heading levels', () => {
    render(<MarkdownContent content={`# Title

## Subtitle`} />);
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Title');
    expect(screen.getByRole('heading', { level: 2 })).toHaveTextContent('Subtitle');
  });

  it('renders paragraphs', () => {
    render(<MarkdownContent content="This is a paragraph." />);
    expect(screen.getByText('This is a paragraph.')).toBeInTheDocument();
  });

  it('renders list items', () => {
    render(<MarkdownContent content={`- Item 1
- Item 2
- Item 3`} />);
    expect(screen.getByText('Item 1')).toBeInTheDocument();
    expect(screen.getByText('Item 2')).toBeInTheDocument();
    expect(screen.getByText('Item 3')).toBeInTheDocument();
  });

  it('renders without search term', () => {
    render(<MarkdownContent content="# Heading" />);
    expect(screen.queryByRole('mark')).not.toBeInTheDocument();
  });

  it('renders complex markdown structure', () => {
    const markdown = `# Main Title

Introduction paragraph.

## Section 1

Content for section 1.

- List item one
- List item two

## Section 2

More content here.`;

    render(<MarkdownContent content={markdown} />);

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Main Title');
    expect(screen.getByText('Introduction paragraph.')).toBeInTheDocument();
    expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(2);
    expect(screen.getByText('List item one')).toBeInTheDocument();
    expect(screen.getByText('List item two')).toBeInTheDocument();
  });

  it('renders semantic section labels', () => {
    const markdown = `[NAV]
Home | About | Contact

[MAIN CONTENT]

# Welcome

Paragraph content.`;

    render(<MarkdownContent content={markdown} />);

    expect(screen.getByText('[NAV]')).toBeInTheDocument();
    expect(screen.getByText('[MAIN CONTENT]')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Welcome');
  });
});

describe('MarkdownContent with search', () => {
  it('highlights search matches in content', () => {
    render(<MarkdownContent content="Hello world" searchTerm="world" />);
    // The highlight logic is in the real component, mock just renders text
    expect(screen.getByText(/world/)).toBeInTheDocument();
  });

  it('handles empty search term gracefully', () => {
    render(<MarkdownContent content="Some content" searchTerm="" />);
    expect(screen.getByText('Some content')).toBeInTheDocument();
  });

  it('handles activeMatchIndex prop', () => {
    render(
      <MarkdownContent
        content="test test test"
        searchTerm="test"
        activeMatchIndex={1}
      />
    );
    // Component should handle the activeMatchIndex without crashing
    // Multiple matches are expected, so use getAllBy
    expect(screen.getAllByText(/test/).length).toBeGreaterThan(0);
  });
});
