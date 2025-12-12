import { useCallback } from 'react'
import { sessionToken } from '../services/sessionToken'
import { getBaseApiUrl, API_ENDPOINTS } from '../constants/api'

interface SessionResponse {
  session_token: string
  expires_at: string
}

interface SessionErrorResponse {
  error: {
    code: string
    message: string
  }
}

export interface UseSessionResult {
  /** Get a valid session token from localStorage, or null if none exists or expired */
  getValidToken: () => string | null
  /** Create a new session by exchanging a Turnstile token for a session token */
  createSession: (turnstileToken: string) => Promise<string | null>
  /** Clear the current session token */
  clearSession: () => void
}

/**
 * Hook to manage session tokens for captcha verification
 * Handles the exchange of Turnstile tokens for session tokens via the backend
 */
export function useSession(): UseSessionResult {
  /**
   * Get a valid session token from localStorage
   * Returns null if no token exists or token is expired
   */
  const getValidToken = useCallback((): string | null => {
    return sessionToken.getValidToken()
  }, [])

  /**
   * Exchange a Turnstile captcha token for a session token
   * Stores the session token in localStorage on success
   * Returns the session token on success, null on failure
   */
  const createSession = useCallback(async (turnstileToken: string): Promise<string | null> => {
    try {
      const apiUrl = getBaseApiUrl() + API_ENDPOINTS.AUTH_CAPTCHA

      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          captcha_token: turnstileToken,
        }),
      })

      if (!response.ok) {
        // Try to parse error response
        try {
          const errorData: SessionErrorResponse = await response.json()
          console.error('Session creation failed:', errorData.error.code, errorData.error.message)
        } catch {
          console.error('Session creation failed with status:', response.status)
        }
        return null
      }

      const data: SessionResponse = await response.json()

      // Store the session token in localStorage
      sessionToken.set(data.session_token, data.expires_at)

      return data.session_token
    } catch (error) {
      console.error('Failed to create session:', error)
      return null
    }
  }, [])

  /**
   * Clear the current session token from localStorage
   */
  const clearSession = useCallback((): void => {
    sessionToken.clear()
  }, [])

  return {
    getValidToken,
    createSession,
    clearSession,
  }
}
