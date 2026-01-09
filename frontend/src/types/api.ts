// API Request/Response types for the rendering API

/**
 * Request body sent to POST /api/render
 */
export interface RenderRequest {
  request_id?: string;
  url: string;
  js_enabled: boolean;
  follow_redirects?: boolean;
  user_agent?: string;
  timeout?: number;
  wait_event?: string;
  block_analytics?: boolean;
  block_ads?: boolean;
  block_social?: boolean;
  blocked_types?: string[];
  session_token?: string;
}

/**
 * Top-level API response wrapper
 */
export interface RenderResponse {
  success: boolean;
  data: RenderData | null;
  error: RenderError | null;
}

/**
 * Error response from the API
 */
export interface RenderError {
  code: string;
  message: string;
}

/**
 * Successful render data from the API
 */
export interface RenderData {
  status_code: number;
  final_url: string;
  redirect_url?: string | null;
  canonical_url?: string | null;
  page_size_bytes: number;
  render_time: number;
  screenshot_id?: string | null;
  meta_robots?: string | null;
  x_robots_tag?: string | null;
  meta_indexable: boolean;
  meta_follow: boolean;

  title: string;
  meta_description?: string | null;
  h1?: string[] | null;
  h2?: string[] | null;
  h3?: string[] | null;
  word_count: number;
  body_text_tokens_count: number;
  body_text?: string | null;
  body_markdown?: string | null;
  text_html_ratio: number;
  open_graph?: Record<string, string> | null;
  structured_data?: unknown[] | null;

  hreflang?: ApiHrefLang[] | null;
  links?: ApiLink[] | null;
  images?: ApiImage[] | null;
  requests?: ApiNetworkRequest[] | null;
  lifecycle?: ApiLifecycleEvent[] | null;
  console?: ApiConsoleMessage[] | null;
  js_errors?: ApiJSError[] | null;

  html?: string | null;
}

/**
 * Alternate language link from hreflang tags
 */
export interface ApiHrefLang {
  lang: string;
  url: string;
  source: string;
}

/**
 * Extracted link from the page
 */
export interface ApiLink {
  href: string;
  text: string;
  is_external: boolean;
  is_dofollow: boolean;
  is_image_link: boolean;
  is_absolute: boolean;
  is_social: boolean;
  is_ugc: boolean;
  is_sponsored: boolean;
}

/**
 * Extracted image from the page
 */
export interface ApiImage {
  src: string;
  alt: string;
  is_external: boolean;
  is_absolute: boolean;
  is_in_link: boolean;
  link_href: string;
  size: number;
}

/**
 * Network request captured during rendering (JS mode only)
 */
export interface ApiNetworkRequest {
  id: string;
  url: string;
  method: string;
  status: number;
  type: string;
  size: number;
  time: number;
  is_internal: boolean;
  blocked: boolean;
  failed: boolean;
}

/**
 * Page lifecycle event (JS mode only)
 */
export interface ApiLifecycleEvent {
  event: string;
  time: number;
}

/**
 * Console message captured during rendering (JS mode only)
 */
export interface ApiConsoleMessage {
  id: string;
  level: string;
  message: string;
  time: number;
}

/**
 * JavaScript error captured during rendering (JS mode only)
 * Note: Not currently implemented in the frontend
 */
export interface ApiJSError {
  message: string;
  source: string;
  line: number;
  column: number;
  stack_trace: string;
  timestamp: number;
}
