import type { PanelConfig } from '../types/config';
import type { RenderRequest, RenderResponse, RenderError } from '../types/api';
import { getBaseApiUrl, API_ENDPOINTS } from '../constants/api';

/**
 * Get the full URL for a screenshot by ID
 */
export function getScreenshotUrl(id: string): string {
  return `${getBaseApiUrl()}/screenshot/${id}`;
}

/**
 * Check if an error is a pool exhausted error (should trigger retry)
 */
export function isPoolExhaustedError(error: RenderError | null): boolean {
  return error?.code === 'POOL_EXHAUSTED';
}

/**
 * Build a RenderRequest from the URL and panel configuration
 */
export function buildRenderRequest(
  url: string,
  config: PanelConfig,
  sessionToken?: string
): RenderRequest {
  const blockedTypes: string[] = [];

  if (config.blocking.imagesMedia) {
    blockedTypes.push('image');
  }

  const request: RenderRequest = {
    url,
    js_enabled: config.jsEnabled,
    follow_redirects: false, // Always false per spec
    user_agent: config.userAgent === 'custom'
      ? config.customUserAgent ?? 'chrome'
      : config.userAgent,
    timeout: config.timeout,
    wait_event: config.waitFor,
    block_analytics: true,  // Always true (trackingScripts locked)
    block_ads: true,        // Always true
    block_social: true,     // Always true
    blocked_types: blockedTypes,
  };

  // Add session token if provided
  if (sessionToken) {
    request.session_token = sessionToken;
  }

  return request;
}

/**
 * Send a render request to the API
 */
export async function renderPage(
  url: string,
  config: PanelConfig,
  sessionToken?: string
): Promise<RenderResponse> {
  const apiUrl = getBaseApiUrl() + API_ENDPOINTS.RENDER;
  const request = buildRenderRequest(url, config, sessionToken);

  try {
    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      // Try to parse error response from backend
      try {
        const errorData = await response.json();
        if (errorData?.error?.code && errorData?.error?.message) {
          return {
            success: false,
            data: null,
            error: {
              code: errorData.error.code,
              message: errorData.error.message,
            },
          };
        }
      } catch {
        // JSON parsing failed, fall through to generic error
      }

      return {
        success: false,
        data: null,
        error: {
          code: 'HTTP_ERROR',
          message: `Request failed: ${response.status}`,
        },
      };
    }

    const data: RenderResponse = await response.json();
    return data;
  } catch (_error) {
    return {
      success: false,
      data: null,
      error: {
        code: 'NETWORK_ERROR',
        message: 'Failed to connect to server',
      },
    };
  }
}
