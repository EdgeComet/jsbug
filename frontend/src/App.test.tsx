import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import App from './App';

describe('App', () => {
  it('renders without crashing', () => {
    render(<App />);
    expect(screen.getByText('jsbug')).toBeInTheDocument();
  });

  it('does not render panels before first analysis', () => {
    render(<App />);
    // Panels should not be visible initially
    expect(screen.queryByText('JS Rendered')).not.toBeInTheDocument();
    expect(screen.queryByText('Non JS')).not.toBeInTheDocument();
    expect(screen.queryByText('VS')).not.toBeInTheDocument();
  });

  it('renders panels after clicking Analyze', async () => {
    render(<App />);
    fireEvent.click(screen.getByText('ANALYZE'));

    // After clicking analyze, panels should appear with loading states
    expect(screen.getByText('JS Rendered')).toBeInTheDocument();
    expect(screen.getByText('Non JS')).toBeInTheDocument();
    expect(screen.getByText('VS')).toBeInTheDocument();
  });

  it('renders URL input with default value', () => {
    render(<App />);
    const input = screen.getByDisplayValue('https://example.com/');
    expect(input).toBeInTheDocument();
  });

  it('opens config modal when config button clicked', () => {
    render(<App />);
    fireEvent.click(screen.getByTitle('Configure render settings'));
    expect(screen.getByText('Render Configuration')).toBeInTheDocument();
  });

  it('closes config modal when Cancel clicked', () => {
    render(<App />);
    fireEvent.click(screen.getByTitle('Configure render settings'));
    expect(screen.getByText('Render Configuration')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Cancel'));
    expect(screen.queryByText('Render Configuration')).not.toBeInTheDocument();
  });

  it('shows loading state in panels when Analyze clicked', async () => {
    render(<App />);
    fireEvent.click(screen.getByText('ANALYZE'));

    // Should show loading states in panels
    const loadingTexts = screen.getAllByText('Analyzing page...');
    expect(loadingTexts.length).toBe(2); // Both panels should show loading
  });
});
