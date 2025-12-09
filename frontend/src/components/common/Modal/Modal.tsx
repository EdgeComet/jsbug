import { ReactNode, RefObject } from 'react';
import { Icon } from '../Icon';
import { useModal } from '../../../hooks/useModal';
import styles from './Modal.module.css';

export type ModalSize = 'sm' | 'md' | 'lg' | 'xl' | 'wide';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: ReactNode;
  size?: ModalSize;
  children: ReactNode;
  headerExtra?: ReactNode;
  footer?: ReactNode;
  searchInputRef?: RefObject<HTMLInputElement | null>;
}

const sizeClasses: Record<ModalSize, string> = {
  sm: styles.modalSm,
  md: styles.modalMd,
  lg: styles.modalLg,
  xl: styles.modalXl,
  wide: styles.modalWide,
};

export function Modal({
  isOpen,
  onClose,
  title,
  size = 'md',
  children,
  headerExtra,
  footer,
  searchInputRef,
}: ModalProps) {
  useModal(isOpen, onClose, searchInputRef);

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className={styles.modalOverlay} onClick={handleOverlayClick}>
      <div className={`${styles.modalDialog} ${sizeClasses[size]}`}>
        <div className={styles.modalHeader}>
          <div className={styles.modalHeaderLeft}>
            <h3 className={styles.modalTitle}>{title}</h3>
            {headerExtra}
          </div>
          <button className={styles.modalClose} onClick={onClose} aria-label="Close">
            <Icon name="x" size={20} />
          </button>
        </div>
        <div className={styles.modalBody}>
          {children}
        </div>
        {footer && (
          <div className={styles.modalFooter}>
            {footer}
          </div>
        )}
      </div>
    </div>
  );
}
