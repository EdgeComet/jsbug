import { useState, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { Icon } from '../common/Icon';
import { getScreenshotUrl } from '../../services/api';
import styles from '../Panel/Panel.module.css';

interface ScreenshotCardProps {
  screenshotId: string;
}

export function ScreenshotCard({ screenshotId }: ScreenshotCardProps) {
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showFullSize, setShowFullSize] = useState(false);

  useEffect(() => {
    setLoading(true);
    setError(null);
    setImageUrl(null);

    const url = getScreenshotUrl(screenshotId);
    const img = new Image();

    img.onload = () => {
      setImageUrl(url);
      setLoading(false);
    };

    img.onerror = () => {
      setError('Failed to load screenshot');
      setLoading(false);
    };

    img.src = url;

    return () => {
      img.onload = null;
      img.onerror = null;
    };
  }, [screenshotId]);

  // Handle escape key to close modal
  useEffect(() => {
    if (!showFullSize) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setShowFullSize(false);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [showFullSize]);

  const handleImageClick = useCallback(() => {
    setShowFullSize(true);
  }, []);

  const handleCloseModal = useCallback(() => {
    setShowFullSize(false);
  }, []);

  return (
    <>
      <div className={styles.resultCard}>
        <div className={styles.resultCardHeader}>
          <span className={styles.resultCardTitle}>
            <Icon name="image" size={14} />
            Screenshot
          </span>
        </div>
        <div className={styles.resultCardBody}>
          <div style={{ padding: '12px' }}>
            {loading && (
              <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px',
                padding: '24px',
                color: 'var(--text-muted)'
              }}>
                <Icon name="loader" size={18} />
                <span>Loading screenshot...</span>
              </div>
            )}

            {error && (
              <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px',
                padding: '24px',
                color: 'var(--text-error, #ef4444)'
              }}>
                <Icon name="alert-circle" size={18} />
                <span>{error}</span>
              </div>
            )}

            {imageUrl && !loading && !error && (
              <div style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: '8px'
              }}>
                <div
                  onClick={handleImageClick}
                  style={{
                    maxWidth: '100%',
                    maxHeight: '200px',
                    overflow: 'hidden',
                    borderRadius: '4px',
                    border: '1px solid var(--border-light)',
                    cursor: 'pointer',
                    transition: 'box-shadow 0.15s ease'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.boxShadow = 'none';
                  }}
                  title="Click to view full size"
                >
                  <img
                    src={imageUrl}
                    alt="Page screenshot"
                    style={{
                      maxWidth: '100%',
                      height: 'auto',
                      objectFit: 'contain',
                      display: 'block'
                    }}
                  />
                </div>
                <span style={{
                  fontSize: '12px',
                  color: 'var(--text-muted)'
                }}>
                  Click to view full size
                </span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Full-size modal */}
      {showFullSize && imageUrl && createPortal(
        <div
          onClick={handleCloseModal}
          style={{
            position: 'fixed',
            inset: 0,
            backgroundColor: 'rgba(0, 0, 0, 0.85)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 9999,
            padding: '24px'
          }}
        >
          {/* Close button */}
          <button
            onClick={handleCloseModal}
            style={{
              position: 'absolute',
              top: '16px',
              right: '16px',
              width: '40px',
              height: '40px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: 'rgba(255, 255, 255, 0.1)',
              border: 'none',
              borderRadius: '50%',
              cursor: 'pointer',
              color: 'white',
              transition: 'background 0.15s ease'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.2)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.1)';
            }}
            title="Close (Esc)"
          >
            <Icon name="x" size={24} />
          </button>

          {/* Image container - stop propagation so clicking image doesn't close modal */}
          <img
            onClick={(e) => e.stopPropagation()}
            src={imageUrl}
            alt="Page screenshot - full size"
            style={{
              maxWidth: '100%',
              maxHeight: '100%',
              objectFit: 'contain',
              borderRadius: '4px',
              boxShadow: '0 4px 24px rgba(0, 0, 0, 0.5)'
            }}
          />

          {/* Help text */}
          <span style={{
            position: 'absolute',
            bottom: '16px',
            left: '50%',
            transform: 'translateX(-50%)',
            color: 'rgba(255, 255, 255, 0.6)',
            fontSize: '12px'
          }}>
            Press Esc or click outside to close
          </span>
        </div>,
        document.body
      )}
    </>
  );
}
