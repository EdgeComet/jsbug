import type { PanelConfig } from '../types/config';
import type { RenderRequest, RenderResponse } from '../types/api';

/**
 * Get the API URL from environment variable or use default
 */
export function getApiUrl(): string {
  return import.meta.env.VITE_API_URL || 'http://localhost:9301/api/render';
}

/**
 * Build a RenderRequest from the URL and panel configuration
 */
export function buildRenderRequest(
  url: string,
  config: PanelConfig,
  captchaToken?: string
): RenderRequest {
  const blockedTypes: string[] = [];

  if (config.blocking.imagesMedia) {
    blockedTypes.push('image');
  }
  if (config.blocking.css) {
    blockedTypes.push('stylesheet');
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

  // Add captcha token if provided
  if (captchaToken) {
    request.captcha_token = captchaToken;
  }

  return request;
}

/**
 * Send a render request to the API
 */
export async function renderPage(
  url: string,
  config: PanelConfig,
  captchaToken?: string
): Promise<RenderResponse> {
  const apiUrl = getApiUrl();
  const request = buildRenderRequest(url, config, captchaToken);

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
