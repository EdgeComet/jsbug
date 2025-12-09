import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { LoadingOverlay } from './LoadingOverlay';

describe('LoadingOverlay', () => {
  it('does not render when isVisible is false', () => {
    render(<LoadingOverlay isVisible={false} />);
    expect(screen.queryByText('Rendering page...')).not.toBeInTheDocument();
  });

  it('renders when isVisible is true', () => {
    render(<LoadingOverlay isVisible={true} />);
    expect(screen.getByText('Rendering page...')).toBeInTheDocument();
  });

  it('displays status message when provided', () => {
    render(<LoadingOverlay isVisible={true} status="Loading JavaScript..." />);
    expect(screen.getByText('Loading JavaScript...')).toBeInTheDocument();
  });

  it('does not display status when not provided', () => {
    render(<LoadingOverlay isVisible={true} />);
    expect(screen.queryByText('Loading JavaScript...')).not.toBeInTheDocument();
  });
});
