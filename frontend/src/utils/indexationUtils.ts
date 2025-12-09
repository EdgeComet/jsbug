export interface IndexabilityResult {
  isIndexable: boolean;
  indexabilityReason: string;
}

/**
 * Determines if a page is canonical.
 * Rules:
 * - Empty/missing canonical URL = true (implicitly self-canonical)
 * - Canonical URL (resolved to absolute) matches page URL = true
 * - Canonical URL differs from page URL = false (canonicalized elsewhere)
 */
export function isCanonical(canonicalUrl: string, pageUrl: string | undefined): boolean {
  // Empty canonical = implicitly self-canonical
  if (!canonicalUrl) return true;

  // Need page URL to resolve relative canonicals and compare
  if (!pageUrl) return false;

  try {
    // Resolve canonical URL to absolute using page URL as base
    // This handles relative URLs like "/path" or "../page"
    const resolvedCanonical = new URL(canonicalUrl, pageUrl).href;
    const normalizedPageUrl = new URL(pageUrl).href;

    // Exact match after resolution
    if (resolvedCanonical === normalizedPageUrl) return true;

    // Normalize trailing slashes for index pages only
    const canonical = new URL(resolvedCanonical);
    const current = new URL(normalizedPageUrl);

    const isCanonicalIndex = canonical.pathname === '/' || canonical.pathname === '';
    const isCurrentIndex = current.pathname === '/' || current.pathname === '';

    if (isCanonicalIndex && isCurrentIndex) {
      return canonical.origin === current.origin;
    }
  } catch {
    // Invalid URL, fall back to strict comparison
    return canonicalUrl === pageUrl;
  }

  return false;
}

export function checkIndexability(
  statusCode: number | null,
  metaIndexable: boolean,
  canonicalUrl: string,
  currentUrl: string | undefined
): IndexabilityResult {
  if (statusCode !== 200) {
    return { isIndexable: false, indexabilityReason: 'Status code is not 200' };
  }

  if (!metaIndexable) {
    return { isIndexable: false, indexabilityReason: 'Meta robots noindex directive' };
  }

  if (!isCanonical(canonicalUrl, currentUrl)) {
    return { isIndexable: false, indexabilityReason: 'Canonical URL points to different page' };
  }

  return { isIndexable: true, indexabilityReason: 'Page is indexable' };
}
