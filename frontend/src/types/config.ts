export type UserAgent = 'googlebot' | 'googlebot-mobile' | 'chrome' | 'chrome-mobile' | 'bingbot' | 'custom';
export type WaitEvent = 'DOMContentLoaded' | 'load' | 'networkIdle' | 'networkAlmostIdle';

export interface PanelConfig {
  jsEnabled: boolean;
  userAgent: UserAgent;
  customUserAgent?: string;
  timeout: number; // 1-30 seconds
  waitFor: WaitEvent;
  blocking: {
    imagesMedia: boolean;
    css: boolean;
    trackingScripts: boolean; // Always true, user can't change
  };
}

export interface AppConfig {
  left: PanelConfig;
  right: PanelConfig;
}
