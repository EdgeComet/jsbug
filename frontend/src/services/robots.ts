import type { RobotsRequest, RobotsResponse } from '../types/robots';
import { getBaseApiUrl, API_ENDPOINTS } from '../constants/api';

/**
 * Validate URL format
 * - Cannot be empty
 * - Must be a valid URL
 * - Scheme must be http or https
 * - Must have a valid host
 */
export function validateRobotsUrl(url: string): { valid: boolean; error?: { code: string; message: string } } {
  if (!url || url.trim() === '') {
    return { valid: false, error: { code: 'INVALID_URL', message: 'URL cannot be empty' } };
  }

  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch {
    return { valid: false, error: { code: 'INVALID_URL', message: 'Invalid URL format' } };
  }

  if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
    return { valid: false, error: { code: 'INVALID_URL', message: 'URL must use http or https scheme' } };
  }

  if (!parsed.hostname) {
    return { valid: false, error: { code: 'INVALID_URL', message: 'URL must have a valid host' } };
  }

  return { valid: true };
}

/**
 * Check if a URL is allowed by robots.txt
 */
export async function checkRobots(url: string): Promise<RobotsResponse> {
  // Client-side validation
  const validation = validateRobotsUrl(url);
  if (!validation.valid) {
    return {
      success: false,
      error: validation.error,
    };
  }

  const apiUrl = getBaseApiUrl() + API_ENDPOINTS.ROBOTS;
  const request: RobotsRequest = { url };

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
            error: {
              code: errorData.error.code,
              message: errorData.error.message,
            },
          };
        }
      } catch {
        // JSON parsing failed, fall through to generic error
      }

      if (response.status === 405) {
        return {
          success: false,
          error: {
            code: 'METHOD_NOT_ALLOWED',
            message: 'Method not allowed (must use POST)',
          },
        };
      }

      return {
        success: false,
        error: {
          code: 'HTTP_ERROR',
          message: `Request failed: ${response.status}`,
        },
      };
    }

    const data: RobotsResponse = await response.json();
    return data;
  } catch {
    return {
      success: false,
      error: {
        code: 'NETWORK_ERROR',
        message: 'Failed to connect to server',
      },
    };
  }
}
