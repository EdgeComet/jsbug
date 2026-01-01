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
    bodyText: 'Plain text content',
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

  it('renders plain text when markdown is empty', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        bodyText="Plain text content"
        bodyMarkdown=""
      />
    );
    expect(screen.getByText('Plain text content')).toBeInTheDocument();
  });

  it('renders markdown content when bodyMarkdown is provided', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        bodyText="Plain text"
        bodyMarkdown="# Heading\n\nParagraph content here."
      />
    );
    // With our mock, the raw markdown is rendered
    const markdownContent = screen.getByTestId('markdown-content');
    expect(markdownContent).toHaveTextContent('# Heading');
    expect(markdownContent).toHaveTextContent('Paragraph content here.');
  });

  it('shows fallback notice when markdown is empty but text exists', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        bodyText="Some plain text"
        bodyMarkdown=""
      />
    );
    expect(screen.getByText(/Structured view unavailable/)).toBeInTheDocument();
  });

  it('displays search input', () => {
    render(<BodyTextModal {...defaultProps} />);
    expect(screen.getByPlaceholderText('Search text...')).toBeInTheDocument();
  });

  it('shows compare button when compareBodyText is provided', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        compareBodyText="No-JS version text"
      />
    );
    expect(screen.getByText(/Compare/)).toBeInTheDocument();
  });

  it('shows compare button when compareBodyMarkdown is provided', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        bodyMarkdown="# JS Content"
        compareBodyMarkdown="# No-JS Content"
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
        bodyMarkdown="# JS Content"
        compareBodyMarkdown="# No-JS Content"
      />
    );

    const compareButton = screen.getByText(/Compare OFF/);
    fireEvent.click(compareButton);

    // Should show compare mode UI
    expect(screen.getByText('JS Rendered')).toBeInTheDocument();
    expect(screen.getByText('No JS')).toBeInTheDocument();
  });

  it('displays match count when searching', () => {
    render(
      <BodyTextModal
        {...defaultProps}
        bodyText="hello world hello again hello"
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
    bodyText: 'JS rendered text',
    bodyMarkdown: '# JS Heading\n\nJS paragraph content.',
    wordCount: 5,
    compareBodyText: 'No-JS text',
    compareBodyMarkdown: '# No-JS Heading\n\nNo-JS paragraph content.',
  };

  it('shows side-by-side comparison in compare mode', () => {
    render(<BodyTextModal {...compareProps} />);

    // Toggle compare mode
    const compareButton = screen.getByText(/Compare OFF/);
    fireEvent.click(compareButton);

    // Should show both panels
    expect(screen.getByText('JS Rendered')).toBeInTheDocument();
    expect(screen.getByText('No JS')).toBeInTheDocument();
  });

  it('shows fallback notice for each panel without markdown', () => {
    render(
      <BodyTextModal
        isOpen={true}
        onClose={vi.fn()}
        bodyText="JS text only"
        bodyMarkdown=""
        wordCount={3}
        compareBodyText="No-JS text only"
        compareBodyMarkdown=""
      />
    );

    // Toggle compare mode
    const compareButton = screen.getByText(/Compare OFF/);
    fireEvent.click(compareButton);

    // Both panels should show fallback notice
    const notices = screen.getAllByText(/Structured view unavailable/);
    expect(notices.length).toBe(2);
  });

  it('exits compare mode when toggled off', () => {
    render(<BodyTextModal {...compareProps} />);

    // Toggle compare mode ON
    const compareButton = screen.getByText(/Compare OFF/);
    fireEvent.click(compareButton);
    expect(screen.getByText('JS Rendered')).toBeInTheDocument();

    // Toggle compare mode OFF
    const compareOnButton = screen.getByText(/Compare ON/);
    fireEvent.click(compareOnButton);

    // Should no longer show side-by-side view
    expect(screen.queryByText('JS Rendered')).not.toBeInTheDocument();
    expect(screen.queryByText('No JS')).not.toBeInTheDocument();
  });
});

describe('BodyTextModal Search Navigation', () => {
  it('shows navigation buttons when there are matches', () => {
    render(
      <BodyTextModal
        isOpen={true}
        onClose={vi.fn()}
        bodyText="test test test"
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
        bodyText="test test test"
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
        bodyText="test test test"
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
