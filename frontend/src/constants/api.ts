export const API_ENDPOINTS = {
  RENDER: '/render',
  ROBOTS: '/robots',
  AUTH_CAPTCHA: '/auth/captcha',
} as const;

export function getBaseApiUrl(): string {
  return import.meta.env.VITE_API_URL || 'http://localhost:9301/api';
}
