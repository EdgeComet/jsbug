export interface IndexabilityResult {
  isIndexable: boolean;
  indexabilityReason: string;
}

function urlsMatch(canonicalUrl: string, currentUrl: string | undefined): boolean {
  if (!currentUrl) return false;
  if (canonicalUrl === currentUrl) return true;

  // For index pages only, normalize trailing slashes
  // Index page = URL path is empty or just "/"
  try {
    const canonical = new URL(canonicalUrl);
    const current = new URL(currentUrl);

    const isCanonicalIndex = canonical.pathname === '/' || canonical.pathname === '';
    const isCurrentIndex = current.pathname === '/' || current.pathname === '';

    if (isCanonicalIndex && isCurrentIndex) {
      // Both are index pages - compare origin only
      return canonical.origin === current.origin;
    }
  } catch {
    // Invalid URL, fall back to strict comparison
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

  if (canonicalUrl && !urlsMatch(canonicalUrl, currentUrl)) {
    return { isIndexable: false, indexabilityReason: 'Canonical URL points to different page' };
  }

  return { isIndexable: true, indexabilityReason: 'Page is indexable' };
}
