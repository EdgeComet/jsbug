export interface TechnicalData {
  statusCode: number | null;  // null for network/SSL errors
  pageSize: number;
  loadTime: number;
  errorMessage?: string;
  redirectUrl?: string;       // For 3xx responses
  finalUrl?: string;          // Final URL after redirects
}

export interface IndexationData {
  metaIndexable: boolean;
  metaFollow: boolean;
  canonicalUrl: string;
  hrefLangs: { lang: string; url: string; source: string }[];
  metaRobots?: string;
  xRobotsTag?: string;
  isIndexable: boolean;
  indexabilityReason: string;
}

export interface Link {
  href: string;
  text: string;
  isExternal: boolean;
  isDofollow: boolean;
  isImageLink: boolean;
  isAbsolute: boolean;
  isSocial: boolean;
  isUgc: boolean;
  isSponsored: boolean;
}

export interface LinksData {
  links: Link[];
}

export interface Image {
  src: string;
  alt: string;
  isExternal: boolean;
  isAbsolute: boolean;
  isInLink: boolean;
  linkHref: string;
  size: number;
}

export interface ImagesData {
  images: Image[];
}

export interface ContentData {
  title: string;
  metaDescription: string;
  h1: string[];
  h2: string[];
  h3: string[];
  bodyWords: number;
  textHtmlRatio: number;
  bodyText: string;
  openGraph?: Record<string, string>;
  structuredData?: unknown[];
  html?: string;
}

export interface LifecycleEvent {
  event: string;
  time: number;
}

export interface TimelineData {
  lifeTimeEvents: LifecycleEvent[];
}
