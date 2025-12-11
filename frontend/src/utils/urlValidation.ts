export interface UrlValidationResult {
  valid: boolean;
  error?: string;
}

/**
 * Validates that a URL is a valid HTTP/HTTPS URL suitable for rendering.
 * Returns detailed error messages for user-facing validation.
 */
export function validateHttpUrl(url: string): UrlValidationResult {
  if (!url.trim()) {
    return { valid: false, error: 'URL is required' };
  }

  try {
    const parsed = new URL(url);

    // Only allow http/https
    if (!['http:', 'https:'].includes(parsed.protocol)) {
      return { valid: false, error: 'URL must use http or https' };
    }

    // Block local URLs
    const host = parsed.hostname.toLowerCase();
    if (host === 'localhost' || host === '127.0.0.1' ||
        host.startsWith('192.168.') || host === '::1' || host === '[::1]') {
      return { valid: false, error: 'Local URLs are not allowed' };
    }

    return { valid: true };
  } catch {
    return { valid: false, error: 'Please enter a valid URL' };
  }
}

/**
 * Simple boolean check for URL validity.
 */
export function isValidHttpUrl(url: string): boolean {
  return validateHttpUrl(url).valid;
}
