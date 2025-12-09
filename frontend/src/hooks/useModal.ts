import { useEffect, useCallback, RefObject } from 'react';

export function useModal(
  isOpen: boolean,
  onClose: () => void,
  searchInputRef?: RefObject<HTMLInputElement | null>
) {
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    }
    // Cmd+F (Mac) or Ctrl+F (Windows/Linux)
    if ((e.metaKey || e.ctrlKey) && e.key === 'f') {
      if (searchInputRef?.current) {
        e.preventDefault();
        searchInputRef.current.focus();
        searchInputRef.current.select();
      }
    }
  }, [onClose, searchInputRef]);

  useEffect(() => {
    if (!isOpen) return;

    document.addEventListener('keydown', handleKeyDown);
    document.body.style.overflow = 'hidden';

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = '';
    };
  }, [isOpen, handleKeyDown]);
}
