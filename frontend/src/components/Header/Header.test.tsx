import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { Header } from './Header';
import { ConfigProvider } from '../../context/ConfigContext';

const renderWithProvider = (ui: React.ReactElement) => {
  return render(<ConfigProvider>{ui}</ConfigProvider>);
};

describe('Header', () => {
  it('renders without crashing', () => {
    renderWithProvider(
      <Header
        url=""
        onUrlChange={() => {}}
        onOpenConfig={() => {}}
        onCompare={() => {}}
      />
    );
    expect(screen.getByText('jsbug')).toBeInTheDocument();
  });

  it('renders URL input with placeholder', () => {
    renderWithProvider(
      <Header
        url=""
        onUrlChange={() => {}}
        onOpenConfig={() => {}}
        onCompare={() => {}}
      />
    );
    expect(screen.getByPlaceholderText(/Enter URL to compare/i)).toBeInTheDocument();
  });

  it('calls onOpenConfig when config button clicked', () => {
    const onOpenConfig = vi.fn();
    renderWithProvider(
      <Header
        url=""
        onUrlChange={() => {}}
        onOpenConfig={onOpenConfig}
        onCompare={() => {}}
      />
    );
    fireEvent.click(screen.getByTitle('Configure render settings'));
    expect(onOpenConfig).toHaveBeenCalled();
  });

  it('calls onCompare when compare button clicked', () => {
    const onCompare = vi.fn();
    renderWithProvider(
      <Header
        url=""
        onUrlChange={() => {}}
        onOpenConfig={() => {}}
        onCompare={onCompare}
      />
    );
    fireEvent.click(screen.getByText('ANALYZE'));
    expect(onCompare).toHaveBeenCalled();
  });

  it('updates URL when input changes', () => {
    const onUrlChange = vi.fn();
    renderWithProvider(
      <Header
        url=""
        onUrlChange={onUrlChange}
        onOpenConfig={() => {}}
        onCompare={() => {}}
      />
    );
    const input = screen.getByPlaceholderText(/Enter URL to compare/i);
    fireEvent.change(input, { target: { value: 'https://example.com' } });
    expect(onUrlChange).toHaveBeenCalledWith('https://example.com');
  });
});
