import type { UserAgent, WaitEvent, PanelConfig, AppConfig } from '../types/config';

// Forward mappings: full value → short code
export const USER_AGENT_SHORT = {
  'googlebot': 'gb',
  'googlebot-mobile': 'gm',
  'chrome': 'c',
  'chrome-mobile': 'cm',
  'bingbot': 'bb',
  'claudebot': 'cb',
  'claude-user': 'cu',
  'gptbot': 'gp',
  'chatgpt-user': 'gu',
  'custom': 'x',
} as const;

export const WAIT_FOR_SHORT = {
  'DOMContentLoaded': 'd',
  'load': 'l',
  'networkIdle': 'ni',
  'networkAlmostIdle': 'na',
} as const;

// Reverse mappings: short code → full value
export const SHORT_USER_AGENT = {
  'gb': 'googlebot',
  'gm': 'googlebot-mobile',
  'c': 'chrome',
  'cm': 'chrome-mobile',
  'bb': 'bingbot',
  'cb': 'claudebot',
  'cu': 'claude-user',
  'gp': 'gptbot',
  'gu': 'chatgpt-user',
  'x': 'custom',
} as const satisfies Record<string, UserAgent>;

export const SHORT_WAIT_FOR = {
  'd': 'DOMContentLoaded',
  'l': 'load',
  'ni': 'networkIdle',
  'na': 'networkAlmostIdle',
} as const satisfies Record<string, WaitEvent>;

// Short parameter keys
export const PARAM_KEYS = {
  timeout: 't',
  userAgent: 'ua',
  customUserAgent: 'cua',
  waitFor: 'wf',
  blockImages: 'bi',
  blockCSS: 'bc',
} as const;

/**
 * Serialize a panel's non-default config values to URL search params
 */
export function serializePanelToParams(
  panel: PanelConfig,
  defaults: PanelConfig,
  prefix: string // 'l' or 'r'
): URLSearchParams {
  const params = new URLSearchParams();

  // Timeout
  if (panel.timeout !== defaults.timeout) {
    params.set(`${prefix}.${PARAM_KEYS.timeout}`, String(panel.timeout));
  }

  // UserAgent
  if (panel.userAgent !== defaults.userAgent) {
    params.set(`${prefix}.${PARAM_KEYS.userAgent}`, USER_AGENT_SHORT[panel.userAgent]);
  }

  // CustomUserAgent (only if userAgent is 'custom')
  if (panel.userAgent === 'custom' && panel.customUserAgent) {
    params.set(`${prefix}.${PARAM_KEYS.customUserAgent}`, panel.customUserAgent);
  }

  // WaitFor
  if (panel.waitFor !== defaults.waitFor) {
    params.set(`${prefix}.${PARAM_KEYS.waitFor}`, WAIT_FOR_SHORT[panel.waitFor]);
  }

  // Blocking options
  if (panel.blocking.imagesMedia !== defaults.blocking.imagesMedia) {
    params.set(`${prefix}.${PARAM_KEYS.blockImages}`, panel.blocking.imagesMedia ? '1' : '0');
  }

  if (panel.blocking.css !== defaults.blocking.css) {
    params.set(`${prefix}.${PARAM_KEYS.blockCSS}`, panel.blocking.css ? '1' : '0');
  }

  // Skip jsEnabled (handled separately) and trackingScripts (always true)

  return params;
}

/**
 * Serialize complete app state to a shareable URL
 */
export function serializeToUrl(
  targetUrl: string,
  config: AppConfig,
  defaults: AppConfig
): string {
  // Build base path
  const basePath = `/u/${targetUrl}`;

  // Get params for both panels
  const leftParams = serializePanelToParams(config.left, defaults.left, 'l');
  const rightParams = serializePanelToParams(config.right, defaults.right, 'r');

  // Merge params
  const mergedParams = new URLSearchParams();
  for (const [key, value] of leftParams) {
    mergedParams.set(key, value);
  }
  for (const [key, value] of rightParams) {
    mergedParams.set(key, value);
  }

  // Build final URL
  const paramString = mergedParams.toString();
  if (paramString) {
    return `${basePath}#${paramString}`;
  }
  return basePath;
}

/**
 * Parse URL hash params back to a partial panel config
 */
export function parsePanelFromParams(
  params: URLSearchParams,
  prefix: string // 'l' or 'r'
): Partial<PanelConfig> {
  const result: Partial<PanelConfig> = {};

  // Timeout
  const timeoutStr = params.get(`${prefix}.${PARAM_KEYS.timeout}`);
  if (timeoutStr) {
    const timeout = parseInt(timeoutStr, 10);
    if (!isNaN(timeout) && timeout >= 1 && timeout <= 30) {
      result.timeout = timeout;
    }
  }

  // UserAgent
  const uaCode = params.get(`${prefix}.${PARAM_KEYS.userAgent}`);
  if (uaCode && uaCode in SHORT_USER_AGENT) {
    result.userAgent = SHORT_USER_AGENT[uaCode as keyof typeof SHORT_USER_AGENT];
  }

  // CustomUserAgent
  const customUa = params.get(`${prefix}.${PARAM_KEYS.customUserAgent}`);
  if (customUa) {
    result.customUserAgent = decodeURIComponent(customUa);
  }

  // WaitFor
  const wfCode = params.get(`${prefix}.${PARAM_KEYS.waitFor}`);
  if (wfCode && wfCode in SHORT_WAIT_FOR) {
    result.waitFor = SHORT_WAIT_FOR[wfCode as keyof typeof SHORT_WAIT_FOR];
  }

  // Blocking options
  const blockImages = params.get(`${prefix}.${PARAM_KEYS.blockImages}`);
  const blockCSS = params.get(`${prefix}.${PARAM_KEYS.blockCSS}`);

  if (blockImages !== null || blockCSS !== null) {
    result.blocking = {
      imagesMedia: blockImages === '1',
      css: blockCSS === '1',
      trackingScripts: true, // Always true
    };
  }

  return result;
}

/**
 * Parsed URL state result
 */
export interface ParsedUrlState {
  targetUrl: string | null;
  leftConfig: Partial<PanelConfig>;
  rightConfig: Partial<PanelConfig>;
}

/**
 * Parse a complete URL into target URL and config
 */
export function parseUrlState(pathname: string, hash: string): ParsedUrlState {
  // Check if pathname starts with '/u/'
  if (!pathname.startsWith('/u/')) {
    return { targetUrl: null, leftConfig: {}, rightConfig: {} };
  }

  // Extract target URL (everything after '/u/')
  const targetUrl = pathname.slice(3);

  // Parse hash (remove leading '#' if present)
  const hashStr = hash.startsWith('#') ? hash.slice(1) : hash;
  const params = new URLSearchParams(hashStr);

  // Parse both panels
  const leftConfig = parsePanelFromParams(params, 'l');
  const rightConfig = parsePanelFromParams(params, 'r');

  return { targetUrl, leftConfig, rightConfig };
}
