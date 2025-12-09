import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ConfigModal } from './ConfigModal';
import { ConfigProvider } from '../../context/ConfigContext';

const renderWithProvider = (ui: React.ReactElement) => {
  return render(<ConfigProvider>{ui}</ConfigProvider>);
};

describe('ConfigModal', () => {
  it('does not render when isOpen is false', () => {
    renderWithProvider(<ConfigModal isOpen={false} onClose={() => {}} />);
    expect(screen.queryByText('Render Configuration')).not.toBeInTheDocument();
  });

  it('renders when isOpen is true', () => {
    renderWithProvider(<ConfigModal isOpen={true} onClose={() => {}} />);
    expect(screen.getByText('Render Configuration')).toBeInTheDocument();
  });

  it('renders both panel columns', () => {
    renderWithProvider(<ConfigModal isOpen={true} onClose={() => {}} />);
    expect(screen.getByText('Left Panel')).toBeInTheDocument();
    expect(screen.getByText('Right Panel')).toBeInTheDocument();
  });

  it('calls onClose when Cancel button clicked', () => {
    const onClose = vi.fn();
    renderWithProvider(<ConfigModal isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByText('Cancel'));
    expect(onClose).toHaveBeenCalled();
  });

  it('calls onClose when Apply button clicked', () => {
    const onClose = vi.fn();
    renderWithProvider(<ConfigModal isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByText('Apply'));
    expect(onClose).toHaveBeenCalled();
  });

  it('calls onClose when close button clicked', () => {
    const onClose = vi.fn();
    renderWithProvider(<ConfigModal isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('Close'));
    expect(onClose).toHaveBeenCalled();
  });

  it('renders JS toggle controls', () => {
    renderWithProvider(<ConfigModal isOpen={true} onClose={() => {}} />);
    // Each panel has ON and OFF buttons for the JS toggle
    const onButtons = screen.getAllByRole('button', { name: 'ON' });
    const offButtons = screen.getAllByRole('button', { name: 'OFF' });
    expect(onButtons.length).toBe(2); // One for each panel
    expect(offButtons.length).toBe(2); // One for each panel
  });

  it('renders timeout sliders', () => {
    renderWithProvider(<ConfigModal isOpen={true} onClose={() => {}} />);
    const sliders = screen.getAllByRole('slider');
    expect(sliders.length).toBe(2);
  });
});
