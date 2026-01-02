import type { PanelConfig } from '../types/config';

export const userAgentLabels: Record<string, string> = {
  googlebot: 'Googlebot',
  'googlebot-mobile': 'Googlebot Mobile',
  'chrome-mobile': 'Chrome Mobile',
  chrome: 'Chrome Desktop',
  bingbot: 'Bingbot',
  custom: 'Custom UA',
};

export function getPanelLabel(config: PanelConfig): string {
  const jsLabel = config.jsEnabled ? 'JS' : 'Non-JS';
  const uaLabel = userAgentLabels[config.userAgent] || config.userAgent;
  return `${jsLabel} \u2022 ${uaLabel} \u2022 ${config.timeout}s`;
}
