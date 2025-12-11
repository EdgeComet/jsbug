/**
 * Cloudflare Turnstile captcha configuration
 *
 * To enable captcha:
 * 1. Get site key from Cloudflare Turnstile dashboard
 * 2. Set VITE_CAPTCHA_ENABLED=true in .env
 * 3. Set VITE_CAPTCHA_SITE_KEY=your_site_key in .env
 */

export const captchaConfig = {
  enabled: import.meta.env.VITE_CAPTCHA_ENABLED === 'true',
  siteKey: import.meta.env.VITE_CAPTCHA_SITE_KEY || '',
} as const;

/**
 * Check if captcha is fully configured and enabled
 * Returns false if disabled or if site key is missing
 */
export function isCaptchaEnabled(): boolean {
  return captchaConfig.enabled && captchaConfig.siteKey !== '';
}
