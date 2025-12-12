/**
 * Session Token Service
 * Manages JWT session tokens in localStorage for captcha verification
 */

const STORAGE_KEY_TOKEN = 'jsbug_session_token'
const STORAGE_KEY_EXPIRY = 'jsbug_session_expiry'

/**
 * Get the stored session token
 * Returns null if no token exists or localStorage is unavailable
 */
export function get(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEY_TOKEN)
  } catch {
    // localStorage unavailable (e.g., private browsing in Safari)
    return null
  }
}

/**
 * Store a session token with its expiry time
 * @param token JWT session token
 * @param expiresAt ISO 8601 timestamp when token expires
 */
export function set(token: string, expiresAt: string): void {
  try {
    localStorage.setItem(STORAGE_KEY_TOKEN, token)
    localStorage.setItem(STORAGE_KEY_EXPIRY, expiresAt)
  } catch {
    // localStorage unavailable - token won't persist
    console.warn('Unable to store session token: localStorage unavailable')
  }
}

/**
 * Clear the stored session token and expiry
 */
export function clear(): void {
  try {
    localStorage.removeItem(STORAGE_KEY_TOKEN)
    localStorage.removeItem(STORAGE_KEY_EXPIRY)
  } catch {
    // localStorage unavailable
  }
}

/**
 * Check if the stored token is still valid (not expired)
 * Returns false if no token exists, token is expired, or localStorage unavailable
 */
export function isValid(): boolean {
  try {
    const expiry = localStorage.getItem(STORAGE_KEY_EXPIRY)
    if (!expiry) {
      return false
    }

    const expiryDate = new Date(expiry)
    const now = new Date()

    // Add 5 minute buffer before actual expiry to avoid edge cases
    const bufferMs = 5 * 60 * 1000
    return expiryDate.getTime() - bufferMs > now.getTime()
  } catch {
    return false
  }
}

/**
 * Get a valid session token, or null if none exists or token is expired
 */
export function getValidToken(): string | null {
  if (!isValid()) {
    return null
  }
  return get()
}

export const sessionToken = {
  get,
  set,
  clear,
  isValid,
  getValidToken,
}
