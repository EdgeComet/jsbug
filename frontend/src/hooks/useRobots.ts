import { useState, useCallback } from 'react';
import { checkRobots } from '../services/robots';

/**
 * Robots.txt check result
 */
export interface RobotsResult {
  isAllowed: boolean;
}

/**
 * Return type of the useRobots hook
 */
export interface UseRobotsResult {
  data: RobotsResult | null;
  error: string | null;
  isLoading: boolean;
  check: (url: string) => Promise<void>;
  reset: () => void;
}

/**
 * Hook for managing robots.txt check state
 */
export function useRobots(): UseRobotsResult {
  const [data, setData] = useState<RobotsResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const check = useCallback(async (url: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await checkRobots(url);

      if (response.success && response.data) {
        setData({ isAllowed: response.data.is_allowed });
        setError(null);
      } else if (response.error) {
        // On error, assume allowed per spec
        setData({ isAllowed: true });
        setError(response.error.message);
      } else {
        // Unknown error, assume allowed per spec
        setData({ isAllowed: true });
        setError(null);
      }
    } catch (_err) {
      // Network error, assume allowed per spec
      setData({ isAllowed: true });
      setError('Failed to check robots.txt');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const reset = useCallback(() => {
    setData(null);
    setError(null);
    setIsLoading(false);
  }, []);

  return {
    data,
    error,
    isLoading,
    check,
    reset,
  };
}
