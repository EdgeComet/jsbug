import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Panel } from './Panel';
import { ConfigProvider } from '../../context/ConfigContext';
import { mockLeftPanel, mockRightPanel } from '../../data/mockData';
import type { PanelData } from '../../hooks/useRenderPanel';

const renderWithProvider = (ui: React.ReactElement) => {
  return render(<ConfigProvider>{ui}</ConfigProvider>);
};

// Convert mock data to PanelData format
const leftPanelData: PanelData = {
  technical: mockLeftPanel.technical,
  indexation: mockLeftPanel.indexation,
  links: mockLeftPanel.links,
  images: mockLeftPanel.images,
  content: { ...mockLeftPanel.content, html: mockLeftPanel.html },
  network: mockLeftPanel.network,
  timeline: mockLeftPanel.timeline,
  console: mockLeftPanel.console,
};

const rightPanelData: PanelData = {
  technical: mockRightPanel.technical,
  indexation: mockRightPanel.indexation,
  links: mockRightPanel.links,
  images: mockRightPanel.images,
  content: mockRightPanel.content,
  network: null,
  timeline: null,
  console: null,
};

describe('Panel', () => {
  it('renders left panel with JS Rendered label', () => {
    renderWithProvider(
      <Panel
        side="left"
        data={leftPanelData}
        jsEnabled={true}
      />
    );
    expect(screen.getByText('JS Rendered')).toBeInTheDocument();
  });

  it('renders right panel with Non JS label', () => {
    renderWithProvider(
      <Panel
        side="right"
        data={rightPanelData}
        jsEnabled={false}
      />
    );
    expect(screen.getByText('Non JS')).toBeInTheDocument();
  });

  it('renders technical card with status code', () => {
    renderWithProvider(
      <Panel
        side="left"
        data={leftPanelData}
        jsEnabled={false}
      />
    );
    expect(screen.getByText('200')).toBeInTheDocument();
  });

  it('renders content card with title', () => {
    renderWithProvider(
      <Panel
        side="left"
        data={leftPanelData}
        jsEnabled={false}
      />
    );
    expect(screen.getByText(mockLeftPanel.content.title)).toBeInTheDocument();
  });

  it('renders tabs when jsEnabled is true', () => {
    renderWithProvider(
      <Panel
        side="left"
        data={leftPanelData}
        jsEnabled={true}
      />
    );
    expect(screen.getByText('Network')).toBeInTheDocument();
    expect(screen.getByText('Timeline')).toBeInTheDocument();
    expect(screen.getByText('Console')).toBeInTheDocument();
    expect(screen.getByText('HTML')).toBeInTheDocument();
  });

  it('shows NoJSInfo when jsEnabled is false', () => {
    renderWithProvider(
      <Panel
        side="right"
        data={rightPanelData}
        jsEnabled={false}
      />
    );
    expect(screen.getByText(/Network, Timeline, and Console data are only available/i)).toBeInTheDocument();
  });
});
