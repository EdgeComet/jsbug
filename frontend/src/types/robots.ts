// Types for the robots.txt checking API

/**
 * Request body sent to POST /api/robots
 */
export interface RobotsRequest {
  url: string;
}

/**
 * Successful robots.txt check data
 */
export interface RobotsData {
  url: string;
  is_allowed: boolean;
}

/**
 * Error response from the robots API
 */
export interface RobotsError {
  code: string;
  message: string;
}

/**
 * Response from POST /api/robots
 */
export interface RobotsResponse {
  success: boolean;
  data?: RobotsData;
  error?: RobotsError;
}
