import { RefObject } from 'react';
import styles from './TurnstileModal.module.css';

interface TurnstileModalProps {
  /** Whether the modal is open */
  isOpen: boolean;
  /** Ref to the container where Turnstile widget will be rendered */
  containerRef: RefObject<HTMLDivElement | null>;
  /** Called when user clicks outside the modal to close it */
  onClose: () => void;
}

/**
 * Modal component for displaying Cloudflare Turnstile challenges
 * The container is always rendered (hidden) so Turnstile can render the widget.
 * The visible overlay only appears when Cloudflare needs user interaction.
 */
export function TurnstileModal({ isOpen, containerRef, onClose }: TurnstileModalProps) {
  const handleOverlayClick = (e: React.MouseEvent) => {
    // Close only if clicking the overlay, not the modal content
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div
      className={styles.overlay}
      style={{ display: isOpen ? 'flex' : 'none' }}
      onClick={handleOverlayClick}
    >
      <div className={styles.modal}>
        <div ref={containerRef} className={styles.widgetContainer} />
      </div>
    </div>
  );
}
