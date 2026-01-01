import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BodyTextModal } from './BodyTextModal';

// Mock scrollIntoView which is not implemented in jsdom
beforeEach(() => {
  Element.prototype.scrollIntoView = vi.fn();
});

// Mock react-markdown to simplify testing
vi.mock('react-markdown', () => ({
  default: ({ children }: { children: string }) => {
    // Simple mock that renders raw markdown content for testing
    return <div data-testid="markdown-content">{children}</div>;
  },
}));

describe('BodyTextModal', () => {
  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    leftBodyMarkdown: 'Plain text content',
    wordCount: 3,
  };

  it('renders modal when isOpen is true', () => {
    render(<BodyTextModal {...defaultProps} />);
    expect(screen.getByText(/Body Text/)).toBeInTheDocument();
    expect(screen.getByText(/3 words/)).toBeInTheDocument();
  });

  it('does not render when isOpen is false', () => {
    render(<BodyTextModal {...defaultProps} isOpen={false} />);
    expect(screen.queryByText(/Body Text/)).not.toBeInTheDocument();
  });

  it('renders markdown content', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        leftBodyMarkdown="# Heading\n\nParagraph content here."
        defaultCompareMode={false}
      />
    );
    // With our mock, the raw markdown is rendered
    const markdownContent = screen.getByTestId('markdown-content');
    expect(markdownContent).toHaveTextContent('# Heading');
    expect(markdownContent).toHaveTextContent('Paragraph content here.');
  });

  it('displays search input', () => {
    render(<BodyTextModal {...defaultProps} />);
    expect(screen.getByPlaceholderText('Search text...')).toBeInTheDocument();
  });

  it('shows compare button when rightBodyMarkdown is provided', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        leftBodyMarkdown="# JS Content"
        rightBodyMarkdown="# No-JS Content"
      />
    );
    expect(screen.getByText(/Compare/)).toBeInTheDocument();
  });

  it('does not show compare button when no compare data is provided', () => {
    render(<BodyTextModal {...defaultProps} />);
    expect(screen.queryByText(/Compare/)).not.toBeInTheDocument();
  });

  it('toggles compare mode when Compare button is clicked', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        leftBodyMarkdown="# JS Content"
        rightBodyMarkdown="# No-JS Content"
      />
    );

    // Compare mode is ON by default
    expect(screen.getByText('Left Panel')).toBeInTheDocument();
    expect(screen.getByText('Right Panel')).toBeInTheDocument();

    // Toggle compare mode OFF
    const compareButton = screen.getByText(/Compare ON/);
    fireEvent.click(compareButton);

    // Should hide compare mode UI
    expect(screen.queryByText('Left Panel')).not.toBeInTheDocument();
    expect(screen.queryByText('Right Panel')).not.toBeInTheDocument();
  });

  it('displays match count when searching', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        leftBodyMarkdown="hello world hello again hello"
      />
    );

    const searchInput = screen.getByPlaceholderText('Search text...');
    fireEvent.change(searchInput, { target: { value: 'hello' } });

    // Should show 3 matches (format: "3 matches" in header, "1 / 3" in navigation)
    expect(screen.getByText(/3 matches/)).toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn();
    render(<BodyTextModal {...defaultProps} onClose={onClose} />);

    const closeButton = screen.getByLabelText('Close');
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalled();
  });

  it('formats word count with locale separators', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        wordCount={1234567}
      />
    );
    // Should format as "1,234,567 words" (locale-dependent)
    expect(screen.getByText(/1,234,567 words/)).toBeInTheDocument();
  });
});

describe('BodyTextModal Compare Mode', () => {
  const compareProps = {
    isOpen: true,
    onClose: vi.fn(),
    leftBodyMarkdown: '# JS Heading\n\nJS paragraph content.',
    wordCount: 5,
    rightBodyMarkdown: '# No-JS Heading\n\nNo-JS paragraph content.',
  };

  it('shows side-by-side comparison in compare mode by default', () => {
    render(<BodyTextModal {...compareProps} />);

    // Compare mode is ON by default - should show both panels
    expect(screen.getByText('Left Panel')).toBeInTheDocument();
    expect(screen.getByText('Right Panel')).toBeInTheDocument();
  });

  it('exits compare mode when toggled off', () => {
    render(<BodyTextModal {...compareProps} />);

    // Compare mode is ON by default
    expect(screen.getByText('Left Panel')).toBeInTheDocument();

    // Toggle compare mode OFF
    const compareButton = screen.getByText(/Compare ON/);
    fireEvent.click(compareButton);

    // Should no longer show side-by-side view
    expect(screen.queryByText('Left Panel')).not.toBeInTheDocument();
    expect(screen.queryByText('Right Panel')).not.toBeInTheDocument();
  });
});

describe('BodyTextModal Search Navigation', () => {
  it('shows navigation buttons when there are matches', () => {
    render(
      <BodyTextModal
        isOpen={true}
        onClose={vi.fn()}
        leftBodyMarkdown="test test test"
        wordCount={3}
      />
    );

    const searchInput = screen.getByPlaceholderText('Search text...');
    fireEvent.change(searchInput, { target: { value: 'test' } });

    // Navigation buttons should be visible
    expect(screen.getByLabelText('Previous match')).toBeInTheDocument();
    expect(screen.getByLabelText('Next match')).toBeInTheDocument();
  });

  it('navigates to next match on Enter key', () => {
    render(
      <BodyTextModal
        isOpen={true}
        onClose={vi.fn()}
        leftBodyMarkdown="test test test"
        wordCount={3}
      />
    );

    const searchInput = screen.getByPlaceholderText('Search text...');
    fireEvent.change(searchInput, { target: { value: 'test' } });
    fireEvent.keyDown(searchInput, { key: 'Enter' });

    // Verify navigation buttons and that we moved to match 2
    expect(screen.getByLabelText('Previous match')).toBeInTheDocument();
    expect(screen.getByLabelText('Next match')).toBeInTheDocument();
    // The match info span should show "2 / 3"
    expect(screen.getByText(/3 matches/)).toBeInTheDocument();
  });

  it('navigates to previous match on Shift+Enter key', () => {
    render(
      <BodyTextModal
        isOpen={true}
        onClose={vi.fn()}
        leftBodyMarkdown="test test test"
        wordCount={3}
      />
    );

    const searchInput = screen.getByPlaceholderText('Search text...');
    fireEvent.change(searchInput, { target: { value: 'test' } });

    // Navigate forward twice
    fireEvent.keyDown(searchInput, { key: 'Enter' });
    fireEvent.keyDown(searchInput, { key: 'Enter' });

    // Now navigate back
    fireEvent.keyDown(searchInput, { key: 'Enter', shiftKey: true });

    // Counter should show navigation (text is split across elements)
    expect(screen.getByLabelText('Previous match')).toBeInTheDocument();
    expect(screen.getByLabelText('Next match')).toBeInTheDocument();
  });
});
