import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ResultTabs } from './ResultTabs';
import { mockLeftPanel } from '../../data/mockData';

describe('ResultTabs', () => {
  const defaultProps = {
    networkData: mockLeftPanel.network,
    timelineData: mockLeftPanel.timeline,
    consoleData: mockLeftPanel.console,
    htmlData: mockLeftPanel.html,
    onOpenHTMLModal: vi.fn(),
  };

  it('renders all tab buttons', () => {
    render(<ResultTabs {...defaultProps} />);
    expect(screen.getByText('Network')).toBeInTheDocument();
    expect(screen.getByText('Timeline')).toBeInTheDocument();
    expect(screen.getByText('Console')).toBeInTheDocument();
    expect(screen.getByText('HTML')).toBeInTheDocument();
  });

  it('shows Network tab content by default', () => {
    render(<ResultTabs {...defaultProps} />);
    expect(screen.getByText('Requests')).toBeInTheDocument();
    expect(screen.getByText('Transferred')).toBeInTheDocument();
  });

  it('switches to Timeline tab when clicked', () => {
    render(<ResultTabs {...defaultProps} />);
    fireEvent.click(screen.getByText('Timeline'));
    const timelineContent = document.querySelector('[data-tab="timeline"]');
    expect(timelineContent).toBeInTheDocument();
  });

  it('switches to Console tab when clicked', () => {
    render(<ResultTabs {...defaultProps} />);
    fireEvent.click(screen.getByText('Console'));
    // Console shows summary stats by default - entries are in modal
    expect(screen.getByText(/lines/)).toBeInTheDocument();
    expect(screen.getByText(/errors/)).toBeInTheDocument();
  });

  it('switches to HTML tab when clicked', () => {
    render(<ResultTabs {...defaultProps} />);
    fireEvent.click(screen.getByText('HTML'));
    expect(screen.getByText('Copy')).toBeInTheDocument();
    expect(screen.getByText('Download')).toBeInTheDocument();
  });

  it('displays request count badge on Network tab', () => {
    render(<ResultTabs {...defaultProps} />);
    // 8 requests in mock data - appears in tab badge
    const elements = screen.getAllByText('8');
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });
});
