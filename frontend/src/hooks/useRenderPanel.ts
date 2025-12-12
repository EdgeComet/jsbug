import { useState, useCallback, useRef, useEffect } from 'react';
import type { PanelConfig } from '../types/config';
import type { RenderData } from '../types/api';
import type { TechnicalData, IndexationData, ContentData, LinksData, ImagesData, TimelineData } from '../types/content';
import type { NetworkData, ResourceType } from '../types/network';
import type { ConsoleEntry, ConsoleLevel } from '../types/console';
import { renderPage, isPoolExhaustedError } from '../services/api';
import { checkIndexability } from '../utils/indexationUtils';

// Pool exhaustion retry configuration
export const MAX_POOL_RETRIES = 5;
const POOL_RETRY_DELAY_MS = 3000;

/**
 * Combined panel data from API response
 */
export interface PanelData {
  technical: TechnicalData;
  indexation: IndexationData;
  content: ContentData;
  links: LinksData;
  images: ImagesData;
  network: NetworkData | null;
  timeline: TimelineData | null;
  console: ConsoleEntry[] | null;
}

/**
 * Return type of the useRenderPanel hook
 */
export interface UseRenderPanelResult {
  data: PanelData | null;
  error: string | null;
  errorCode: string | null;
  isLoading: boolean;
  isRetrying: boolean;
  retryCount: number;
  render: (url: string, config: PanelConfig, sessionToken?: string) => Promise<{ errorCode: string | null }>;
  reset: () => void;
  cancelRetry: () => void;
}

// Message shown when pool exhaustion retries are exhausted
const POOL_EXHAUSTED_MESSAGE =
  'All rendering slots are currently in use. Please wait a moment and try again.';

/**
 * Transform API response to frontend types
 */
function transformResponse(response: RenderData, jsEnabled: boolean): PanelData {
  return {
    technical: {
      statusCode: response.status_code,
      pageSize: response.page_size_bytes,
      loadTime: response.render_time,
      redirectUrl: response.redirect_url ?? undefined,
      finalUrl: response.final_url,
    },
    indexation: {
      metaIndexable: response.meta_indexable,
      metaFollow: response.meta_follow,
      canonicalUrl: response.canonical_url ?? '',
      hrefLangs: (response.hreflang ?? []).map(h => ({
        lang: h.lang,
        url: h.url,
        source: h.source,
      })),
      metaRobots: response.meta_robots ?? undefined,
      xRobotsTag: response.x_robots_tag ?? undefined,
      ...checkIndexability(response.status_code, response.meta_indexable, response.canonical_url ?? '', response.final_url),
    },
    content: {
      title: response.title,
      metaDescription: response.meta_description ?? '',
      h1: response.h1 ?? [],
      h2: response.h2 ?? [],
      h3: response.h3 ?? [],
      bodyWords: response.word_count,
      textHtmlRatio: response.text_html_ratio,
      bodyText: response.body_text ?? '',
      openGraph: response.open_graph ?? undefined,
      structuredData: response.structured_data ?? undefined,
      html: response.html ?? undefined,
    },
    links: {
      links: (response.links ?? []).map(link => ({
        href: link.href,
        text: link.text,
        isExternal: link.is_external,
        isDofollow: link.is_dofollow,
        isImageLink: link.is_image_link,
        isAbsolute: link.is_absolute,
        isSocial: link.is_social,
        isUgc: link.is_ugc,
        isSponsored: link.is_sponsored,
      })),
    },
    images: {
      images: (response.images ?? []).map(img => ({
        src: img.src,
        alt: img.alt,
        isExternal: img.is_external,
        isAbsolute: img.is_absolute,
        isInLink: img.is_in_link,
        linkHref: img.link_href,
        size: img.size,
      })),
    },
    network: jsEnabled ? {
      requests: (response.requests ?? []).map(req => ({
        id: req.id,
        url: req.url,
        method: req.method,
        status: req.status,
        type: req.type as ResourceType,
        size: req.url.startsWith('data:image/') ? req.url.length : req.size,
        time: req.time,
        blocked: req.blocked,
        failed: req.failed,
        isInternal: req.is_internal,
      })),
    } : null,
    timeline: jsEnabled ? {
      lifeTimeEvents: (response.lifecycle ?? []).map(evt => ({
        event: evt.event,
        time: evt.time,
      })),
    } : null,
    console: jsEnabled ? (response.console ?? []).map(msg => ({
      id: msg.id,
      level: msg.level as ConsoleLevel,
      message: msg.message,
      time: msg.time,
    })) : null,
  };
}

/**
 * Hook for managing panel render state
 */
export function useRenderPanel(): UseRenderPanelResult {
  const [data, setData] = useState<PanelData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [errorCode, setErrorCode] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isRetrying, setIsRetrying] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const retryTimeoutRef = useRef<number | null>(null);
  const cancelledRef = useRef(false);

  // Cancel any active retry timeout
  const cancelRetry = useCallback(() => {
    cancelledRef.current = true;
    if (retryTimeoutRef.current !== null) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }
    setIsRetrying(false);
    setRetryCount(0);
  }, []);

  // Cleanup on unmount to prevent memory leaks and state updates on unmounted component
  useEffect(() => {
    return () => {
      if (retryTimeoutRef.current !== null) {
        clearTimeout(retryTimeoutRef.current);
      }
    };
  }, []);

  const render = useCallback(async (url: string, config: PanelConfig, sessionToken?: string) => {
    // Cancel any active retry from previous render
    cancelRetry();
    cancelledRef.current = false; // Reset after cancel for this new render

    setIsLoading(true);
    setError(null);
    setErrorCode(null);

    // Internal retry loop for pool exhaustion
    let currentAttempt = 0;

    const attemptRender = async (): Promise<{ errorCode: string | null }> => {
      try {
        const response = await renderPage(url, config, sessionToken);

        if (response.success && response.data) {
          const transformedData = transformResponse(response.data, config.jsEnabled);
          setData(transformedData);
          setError(null);
          setErrorCode(null);
          setIsLoading(false);
          setIsRetrying(false);
          setRetryCount(0);
          return { errorCode: null };
        } else if (response.error) {
          // Check if this is a pool exhaustion error that should trigger retry
          if (isPoolExhaustedError(response.error) && currentAttempt < MAX_POOL_RETRIES) {
            currentAttempt++;
            setRetryCount(currentAttempt);
            setIsRetrying(true);
            setIsLoading(false);

            // Wait and retry
            await new Promise<void>((resolve) => {
              retryTimeoutRef.current = window.setTimeout(() => {
                retryTimeoutRef.current = null;
                resolve();
              }, POOL_RETRY_DELAY_MS);
            });

            // Check if we were cancelled during the wait
            if (!cancelledRef.current) {
              // Continue with retry
              return attemptRender();
            }
            // Cancelled - explicitly ensure panel is in valid idle state
            setIsLoading(false);
            setIsRetrying(false);
            setRetryCount(0);
            return { errorCode: null };
          } else {
            // Not a pool error, or retries exhausted - show error
            setData(null);
            setErrorCode(response.error.code);
            // Use friendly message for pool exhaustion after retries exhausted
            if (isPoolExhaustedError(response.error)) {
              setError(POOL_EXHAUSTED_MESSAGE);
            } else {
              setError(response.error.message);
            }
            setIsLoading(false);
            setIsRetrying(false);
            setRetryCount(0);
            return { errorCode: response.error.code };
          }
        } else {
          setData(null);
          setError('Unknown error occurred');
          setErrorCode(null);
          setIsLoading(false);
          setIsRetrying(false);
          setRetryCount(0);
          return { errorCode: null };
        }
      } catch (_err) {
        setData(null);
        setError('Failed to connect to server');
        setErrorCode(null);
        setIsLoading(false);
        setIsRetrying(false);
        setRetryCount(0);
        return { errorCode: null };
      }
    };

    return await attemptRender();
  }, [cancelRetry]);

  const reset = useCallback(() => {
    cancelRetry();
    setData(null);
    setError(null);
    setErrorCode(null);
    setIsLoading(true);
  }, [cancelRetry]);

  return {
    data,
    error,
    errorCode,
    isLoading,
    isRetrying,
    retryCount,
    render,
    reset,
    cancelRetry,
  };
}
