import { describe, it, expect } from 'vitest';
import {
  USER_AGENT_SHORT,
  WAIT_FOR_SHORT,
  SHORT_USER_AGENT,
  SHORT_WAIT_FOR,
  serializePanelToParams,
  serializeToUrl,
  parsePanelFromParams,
  parseUrlState,
} from './urlState';
import type { UserAgent, WaitEvent, PanelConfig, AppConfig } from '../types/config';

describe('URL State Constants', () => {
  describe('USER_AGENT mappings', () => {
    const allUserAgents: UserAgent[] = [
      'googlebot',
      'googlebot-mobile',
      'chrome',
      'chrome-mobile',
      'bingbot',
      'claudebot',
      'claude-user',
      'gptbot',
      'chatgpt-user',
      'custom',
    ];

    it('should have forward mapping for all UserAgent values', () => {
      for (const ua of allUserAgents) {
        expect(USER_AGENT_SHORT[ua]).toBeDefined();
      }
    });

    it('should have reverse mapping for all short codes', () => {
      for (const ua of allUserAgents) {
        const shortCode = USER_AGENT_SHORT[ua];
        expect(SHORT_USER_AGENT[shortCode]).toBe(ua);
      }
    });

    it('should return correct reverse lookups', () => {
      expect(SHORT_USER_AGENT['gb']).toBe('googlebot');
      expect(SHORT_USER_AGENT['gm']).toBe('googlebot-mobile');
      expect(SHORT_USER_AGENT['c']).toBe('chrome');
      expect(SHORT_USER_AGENT['cm']).toBe('chrome-mobile');
      expect(SHORT_USER_AGENT['bb']).toBe('bingbot');
      expect(SHORT_USER_AGENT['cb']).toBe('claudebot');
      expect(SHORT_USER_AGENT['cu']).toBe('claude-user');
      expect(SHORT_USER_AGENT['gp']).toBe('gptbot');
      expect(SHORT_USER_AGENT['gu']).toBe('chatgpt-user');
      expect(SHORT_USER_AGENT['x']).toBe('custom');
    });
  });

  describe('WAIT_FOR mappings', () => {
    const allWaitEvents: WaitEvent[] = [
      'DOMContentLoaded',
      'load',
      'networkIdle',
      'networkAlmostIdle',
    ];

    it('should have forward mapping for all WaitEvent values', () => {
      for (const event of allWaitEvents) {
        expect(WAIT_FOR_SHORT[event]).toBeDefined();
      }
    });

    it('should have reverse mapping for all short codes', () => {
      for (const event of allWaitEvents) {
        const shortCode = WAIT_FOR_SHORT[event];
        expect(SHORT_WAIT_FOR[shortCode]).toBe(event);
      }
    });

    it('should return correct reverse lookups', () => {
      expect(SHORT_WAIT_FOR['d']).toBe('DOMContentLoaded');
      expect(SHORT_WAIT_FOR['l']).toBe('load');
      expect(SHORT_WAIT_FOR['ni']).toBe('networkIdle');
      expect(SHORT_WAIT_FOR['na']).toBe('networkAlmostIdle');
    });
  });
});

describe('serializePanelToParams', () => {
  const defaultLeftConfig: PanelConfig = {
    jsEnabled: true,
    userAgent: 'chrome-mobile',
    timeout: 15,
    waitFor: 'networkIdle',
    blocking: { imagesMedia: false, css: false, trackingScripts: true },
  };

  it('should return empty params when panel matches defaults exactly', () => {
    const params = serializePanelToParams(defaultLeftConfig, defaultLeftConfig, 'l');
    expect(params.toString()).toBe('');
  });

  it('should only include changed values in params', () => {
    const panel: PanelConfig = {
      ...defaultLeftConfig,
      timeout: 20,
    };
    const params = serializePanelToParams(panel, defaultLeftConfig, 'l');
    expect(params.get('l.t')).toBe('20');
    expect(params.get('l.ua')).toBeNull();
    expect(params.get('l.wf')).toBeNull();
  });

  it('should serialize timeout correctly', () => {
    const panel: PanelConfig = { ...defaultLeftConfig, timeout: 20 };
    const params = serializePanelToParams(panel, defaultLeftConfig, 'l');
    expect(params.get('l.t')).toBe('20');
  });

  it('should serialize userAgent to short code', () => {
    const panel: PanelConfig = { ...defaultLeftConfig, userAgent: 'googlebot' };
    const params = serializePanelToParams(panel, defaultLeftConfig, 'l');
    expect(params.get('l.ua')).toBe('gb');
  });

  it('should serialize waitFor to short code', () => {
    const panel: PanelConfig = { ...defaultLeftConfig, waitFor: 'DOMContentLoaded' };
    const params = serializePanelToParams(panel, defaultLeftConfig, 'l');
    expect(params.get('l.wf')).toBe('d');
  });

  it('should serialize blocking booleans to 1/0', () => {
    const panel: PanelConfig = {
      ...defaultLeftConfig,
      blocking: { imagesMedia: true, css: true, trackingScripts: true },
    };
    const params = serializePanelToParams(panel, defaultLeftConfig, 'l');
    expect(params.get('l.bi')).toBe('1');
    expect(params.get('l.bc')).toBe('1');
  });

  it('should serialize blocking false as 0 when default is true', () => {
    const defaultWithBlocking: PanelConfig = {
      ...defaultLeftConfig,
      blocking: { imagesMedia: true, css: true, trackingScripts: true },
    };
    const panel: PanelConfig = {
      ...defaultLeftConfig,
      blocking: { imagesMedia: false, css: false, trackingScripts: true },
    };
    const params = serializePanelToParams(panel, defaultWithBlocking, 'l');
    expect(params.get('l.bi')).toBe('0');
    expect(params.get('l.bc')).toBe('0');
  });

  it('should only include customUserAgent when userAgent is custom', () => {
    const panel: PanelConfig = {
      ...defaultLeftConfig,
      userAgent: 'custom',
      customUserAgent: 'My Bot/1.0',
    };
    const params = serializePanelToParams(panel, defaultLeftConfig, 'l');
    expect(params.get('l.ua')).toBe('x');
    expect(params.get('l.cua')).toBe('My Bot/1.0');
  });

  it('should not include customUserAgent when userAgent is not custom', () => {
    const panel: PanelConfig = {
      ...defaultLeftConfig,
      userAgent: 'googlebot',
      customUserAgent: 'Some value', // Should be ignored
    };
    const params = serializePanelToParams(panel, defaultLeftConfig, 'l');
    expect(params.get('l.cua')).toBeNull();
  });

  it('should use r prefix for right panel', () => {
    const panel: PanelConfig = { ...defaultLeftConfig, timeout: 10 };
    const params = serializePanelToParams(panel, defaultLeftConfig, 'r');
    expect(params.get('r.t')).toBe('10');
    expect(params.get('l.t')).toBeNull();
  });
});

describe('serializeToUrl', () => {
  const defaultConfig: AppConfig = {
    left: {
      jsEnabled: true,
      userAgent: 'chrome-mobile',
      timeout: 15,
      waitFor: 'networkIdle',
      blocking: { imagesMedia: false, css: false, trackingScripts: true },
    },
    right: {
      jsEnabled: false,
      userAgent: 'chrome-mobile',
      timeout: 10,
      waitFor: 'load',
      blocking: { imagesMedia: false, css: false, trackingScripts: true },
    },
  };

  it('should produce URL with no hash when config matches defaults', () => {
    const url = serializeToUrl('https://example.com', defaultConfig, defaultConfig);
    expect(url).toBe('/u/https://example.com');
  });

  it('should preserve target URL query params', () => {
    const url = serializeToUrl('https://example.com?foo=1', defaultConfig, defaultConfig);
    expect(url).toBe('/u/https://example.com?foo=1');
  });

  it('should produce hash with single changed value', () => {
    const config: AppConfig = {
      ...defaultConfig,
      left: { ...defaultConfig.left, timeout: 20 },
    };
    const url = serializeToUrl('https://example.com', config, defaultConfig);
    expect(url).toBe('/u/https://example.com#l.t=20');
  });

  it('should produce hash with multiple changes', () => {
    const config: AppConfig = {
      ...defaultConfig,
      left: { ...defaultConfig.left, timeout: 20 },
      right: { ...defaultConfig.right, userAgent: 'googlebot' },
    };
    const url = serializeToUrl('https://example.com', config, defaultConfig);
    expect(url).toContain('/u/https://example.com#');
    expect(url).toContain('l.t=20');
    expect(url).toContain('r.ua=gb');
  });

  it('should handle both panels having different changes', () => {
    const config: AppConfig = {
      left: { ...defaultConfig.left, timeout: 20, waitFor: 'load' },
      right: { ...defaultConfig.right, timeout: 5, userAgent: 'bingbot' },
    };
    const url = serializeToUrl('https://example.com', config, defaultConfig);
    expect(url).toContain('l.t=20');
    expect(url).toContain('l.wf=l');
    expect(url).toContain('r.t=5');
    expect(url).toContain('r.ua=bb');
  });

  it('should return empty hash when all values match defaults', () => {
    const url = serializeToUrl('https://example.com', defaultConfig, defaultConfig);
    expect(url).not.toContain('#');
  });
});

describe('parsePanelFromParams', () => {
  it('should return empty object for empty params', () => {
    const params = new URLSearchParams('');
    const result = parsePanelFromParams(params, 'l');
    expect(result).toEqual({});
  });

  it('should parse timeout correctly', () => {
    const params = new URLSearchParams('l.t=20');
    const result = parsePanelFromParams(params, 'l');
    expect(result.timeout).toBe(20);
  });

  it('should ignore invalid timeout (non-numeric)', () => {
    const params = new URLSearchParams('l.t=abc');
    const result = parsePanelFromParams(params, 'l');
    expect(result.timeout).toBeUndefined();
  });

  it('should ignore out of range timeout (too high)', () => {
    const params = new URLSearchParams('l.t=100');
    const result = parsePanelFromParams(params, 'l');
    expect(result.timeout).toBeUndefined();
  });

  it('should ignore out of range timeout (too low)', () => {
    const params = new URLSearchParams('l.t=0');
    const result = parsePanelFromParams(params, 'l');
    expect(result.timeout).toBeUndefined();
  });

  it('should parse userAgent short code', () => {
    const params = new URLSearchParams('l.ua=gb');
    const result = parsePanelFromParams(params, 'l');
    expect(result.userAgent).toBe('googlebot');
  });

  it('should ignore invalid userAgent code', () => {
    const params = new URLSearchParams('l.ua=zz');
    const result = parsePanelFromParams(params, 'l');
    expect(result.userAgent).toBeUndefined();
  });

  it('should parse waitFor short code', () => {
    const params = new URLSearchParams('l.wf=d');
    const result = parsePanelFromParams(params, 'l');
    expect(result.waitFor).toBe('DOMContentLoaded');
  });

  it('should ignore invalid waitFor code', () => {
    const params = new URLSearchParams('l.wf=invalid');
    const result = parsePanelFromParams(params, 'l');
    expect(result.waitFor).toBeUndefined();
  });

  it('should parse blocking booleans', () => {
    const params = new URLSearchParams('l.bi=1&l.bc=0');
    const result = parsePanelFromParams(params, 'l');
    expect(result.blocking).toEqual({
      imagesMedia: true,
      css: false,
      trackingScripts: true,
    });
  });

  it('should decode customUserAgent', () => {
    const params = new URLSearchParams('l.cua=My%20Bot');
    const result = parsePanelFromParams(params, 'l');
    expect(result.customUserAgent).toBe('My Bot');
  });

  it('should parse right prefix correctly', () => {
    const params = new URLSearchParams('r.t=10');
    const result = parsePanelFromParams(params, 'r');
    expect(result.timeout).toBe(10);
  });

  it('should not parse wrong prefix', () => {
    const params = new URLSearchParams('r.t=10');
    const result = parsePanelFromParams(params, 'l');
    expect(result.timeout).toBeUndefined();
  });

  it('should handle mixed valid/invalid params', () => {
    const params = new URLSearchParams('l.t=20&l.ua=invalid');
    const result = parsePanelFromParams(params, 'l');
    expect(result.timeout).toBe(20);
    expect(result.userAgent).toBeUndefined();
  });
});

describe('parseUrlState', () => {
  it('should return null targetUrl when no /u/ prefix', () => {
    const result = parseUrlState('/', '');
    expect(result.targetUrl).toBeNull();
    expect(result.leftConfig).toEqual({});
    expect(result.rightConfig).toEqual({});
  });

  it('should parse simple URL', () => {
    const result = parseUrlState('/u/https://example.com', '');
    expect(result.targetUrl).toBe('https://example.com');
    expect(result.leftConfig).toEqual({});
    expect(result.rightConfig).toEqual({});
  });

  it('should preserve target URL query params', () => {
    const result = parseUrlState('/u/https://example.com?foo=1', '');
    expect(result.targetUrl).toBe('https://example.com?foo=1');
  });

  it('should parse hash with left config', () => {
    const result = parseUrlState('/u/https://example.com', '#l.t=20');
    expect(result.targetUrl).toBe('https://example.com');
    expect(result.leftConfig.timeout).toBe(20);
    expect(result.rightConfig).toEqual({});
  });

  it('should parse hash with both panels', () => {
    const result = parseUrlState('/u/https://example.com', '#l.t=20&r.ua=gb');
    expect(result.leftConfig.timeout).toBe(20);
    expect(result.rightConfig.userAgent).toBe('googlebot');
  });

  it('should handle hash without # prefix', () => {
    const result = parseUrlState('/u/https://example.com', 'l.t=20');
    expect(result.leftConfig.timeout).toBe(20);
  });

  it('should handle empty hash', () => {
    const result = parseUrlState('/u/https://example.com', '');
    expect(result.leftConfig).toEqual({});
    expect(result.rightConfig).toEqual({});
  });

  it('should handle root path', () => {
    const result = parseUrlState('/', '');
    expect(result.targetUrl).toBeNull();
  });

  it('should handle other paths without /u/', () => {
    const result = parseUrlState('/about', '');
    expect(result.targetUrl).toBeNull();
  });

  // Round-trip test
  it('should round-trip serialize then parse', () => {
    const defaultConfig: AppConfig = {
      left: {
        jsEnabled: true,
        userAgent: 'chrome-mobile',
        timeout: 15,
        waitFor: 'networkIdle',
        blocking: { imagesMedia: false, css: false, trackingScripts: true },
      },
      right: {
        jsEnabled: false,
        userAgent: 'chrome-mobile',
        timeout: 10,
        waitFor: 'load',
        blocking: { imagesMedia: false, css: false, trackingScripts: true },
      },
    };

    const modifiedConfig: AppConfig = {
      left: { ...defaultConfig.left, timeout: 20, userAgent: 'googlebot' },
      right: { ...defaultConfig.right, timeout: 5, waitFor: 'DOMContentLoaded' },
    };

    // Serialize
    const url = serializeToUrl('https://example.com', modifiedConfig, defaultConfig);

    // Parse the URL parts
    const hashIndex = url.indexOf('#');
    const pathname = hashIndex >= 0 ? url.slice(0, hashIndex) : url;
    const hash = hashIndex >= 0 ? url.slice(hashIndex) : '';

    // Parse
    const parsed = parseUrlState(pathname, hash);

    // Verify
    expect(parsed.targetUrl).toBe('https://example.com');
    expect(parsed.leftConfig.timeout).toBe(20);
    expect(parsed.leftConfig.userAgent).toBe('googlebot');
    expect(parsed.rightConfig.timeout).toBe(5);
    expect(parsed.rightConfig.waitFor).toBe('DOMContentLoaded');
  });
});
