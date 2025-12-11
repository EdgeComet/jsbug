import { useState, useCallback, useRef, useEffect } from 'react';
import { isCaptchaEnabled, captchaConfig } from '../config/captcha';

// Turnstile type declarations
declare global {
  interface Window {
    turnstile?: {
      render: (container: HTMLElement, options: TurnstileRenderOptions) => string;
      reset: (widgetId: string) => void;
      remove: (widgetId: string) => void;
    };
  }
}

interface TurnstileRenderOptions {
  sitekey: string;
  callback: (token: string) => void;
  'error-callback': () => void;
  'expired-callback': () => void;
  'timeout-callback': () => void;
  'before-interactive-callback'?: () => void;
  appearance?: 'always' | 'execute' | 'interaction-only';
  theme?: 'light' | 'dark' | 'auto';
}

export interface UseTurnstileResult {
  /** Get a fresh captcha token. Returns null if cancelled or failed. */
  getToken: () => Promise<string | null>;
  /** Whether Turnstile is currently loading/verifying */
  isLoading: boolean;
  /** Whether the modal should be shown */
  showModal: boolean;
  /** Ref to attach to the modal container div */
  modalContainerRef: React.RefObject<HTMLDivElement | null>;
  /** Close the modal and cancel the current token request */
  closeModal: () => void;
}

/**
 * Hook to manage Cloudflare Turnstile captcha
 * Uses managed/interaction-only mode - widget is invisible unless challenge needed
 */
export function useTurnstile(): UseTurnstileResult {
  const [isLoading, setIsLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const modalContainerRef = useRef<HTMLDivElement | null>(null);
  const widgetIdRef = useRef<string | null>(null);
  const resolveRef = useRef<((token: string | null) => void) | null>(null);

  // Load Turnstile script on mount
  useEffect(() => {
    if (!isCaptchaEnabled()) return;
    if (document.getElementById('turnstile-script')) return;

    const script = document.createElement('script');
    script.id = 'turnstile-script';
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js';
    script.async = true;
    script.defer = true;
    document.head.appendChild(script);

    return () => {
      // Cleanup widget on unmount
      if (widgetIdRef.current && window.turnstile) {
        window.turnstile.remove(widgetIdRef.current);
      }
    };
  }, []);

  const closeModal = useCallback(() => {
    setShowModal(false);
    setIsLoading(false);

    if (widgetIdRef.current && window.turnstile) {
      window.turnstile.remove(widgetIdRef.current);
      widgetIdRef.current = null;
    }

    // Resolve with null if modal was closed without getting token
    if (resolveRef.current) {
      resolveRef.current(null);
      resolveRef.current = null;
    }
  }, []);

  const getToken = useCallback((): Promise<string | null> => {
    // If captcha is disabled, resolve immediately with null
    if (!isCaptchaEnabled()) {
      return Promise.resolve(null);
    }

    return new Promise((resolve) => {
      resolveRef.current = resolve;
      setIsLoading(true);
      // Don't show modal yet - only show when Turnstile needs interaction

      // Wait for turnstile script to load
      const renderWidget = () => {
        if (!window.turnstile || !modalContainerRef.current) {
          // Script not loaded yet, retry
          setTimeout(renderWidget, 50);
          return;
        }

        // Clear any existing widget
        if (widgetIdRef.current) {
          window.turnstile.remove(widgetIdRef.current);
          widgetIdRef.current = null;
        }

        // Render new widget
        widgetIdRef.current = window.turnstile.render(modalContainerRef.current, {
          sitekey: captchaConfig.siteKey,
          appearance: 'interaction-only', // Managed mode - invisible unless challenge needed
          theme: 'auto',
          'before-interactive-callback': () => {
            // Only show modal when Turnstile needs user interaction
            setShowModal(true);
          },
          callback: (token: string) => {
            // Success - got token
            setIsLoading(false);
            setShowModal(false);
            if (resolveRef.current) {
              resolveRef.current(token);
              resolveRef.current = null;
            }
          },
          'error-callback': () => {
            // Error occurred - resolve with null to unblock caller
            setIsLoading(false);
            setShowModal(false);
            if (resolveRef.current) {
              resolveRef.current(null);
              resolveRef.current = null;
            }
          },
          'expired-callback': () => {
            // Token expired before use - reset for retry
            if (widgetIdRef.current && window.turnstile) {
              window.turnstile.reset(widgetIdRef.current);
            }
          },
          'timeout-callback': () => {
            // Timeout - close and resolve null
            setIsLoading(false);
            setShowModal(false);
            if (resolveRef.current) {
              resolveRef.current(null);
              resolveRef.current = null;
            }
          },
        });
      };

      renderWidget();
    });
  }, []);

  return {
    getToken,
    isLoading,
    showModal,
    modalContainerRef,
    closeModal,
  };
}
